package resumes

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"log/slog"
)

// Keep the same interface name so existing wiring needs minimal changes.
type RabbitMQPublisher interface {
    PublishResumeGenerationJob(ctx context.Context, job *ResumeWorkerJob) error
    Close() error
}

// ResumeWorkerJob mirrors the previous message structure used by RabbitMQ publisher.
type ResumeWorkerJob struct {
    JobID          string                 `json:"jobId"`
    UserID         string                 `json:"userId"`
    UserEmail      string                 `json:"userEmail"`
    UserName       string                 `json:"userName"`
    JobDescription string                 `json:"jobDescription"`
    Metadata       map[string]interface{} `json:"metadata"`
}

type httpPublisher struct {
    client *http.Client
    url    string
    logger *slog.Logger
}

// NewHTTPPublisher creates a publisher that POSTs jobs to the resume-generator HTTP API.
// It implements the same interface name used elsewhere to minimize code changes.
func NewHTTPPublisher(resumeGeneratorURL string, timeoutMs int, logger *slog.Logger) RabbitMQPublisher {
    if resumeGeneratorURL == "" {
        resumeGeneratorURL = "http://resume-generator:3000"
    }
    if timeoutMs <= 0 {
        timeoutMs = 5000
    }

    return &httpPublisher{
        client: &http.Client{Timeout: time.Duration(timeoutMs) * time.Millisecond},
        url:    resumeGeneratorURL,
        logger: logger,
    }
}

// PublishResumeGenerationJob sends a POST to /jobs/start on the resume-generator.
func (p *httpPublisher) PublishResumeGenerationJob(ctx context.Context, job *ResumeWorkerJob) error {
	if p == nil {
		return NewDomainError(ErrCodePublisherNotInitialized, fmt.Errorf("publisher not initialized"), map[string]interface{}{
			"jobId": job.JobID,
		})
	}

	// Build payload expected by resume-generator
	payload := map[string]interface{}{
		"userId": job.UserID,
	}

	// Extract jobApplicationId from metadata if present
	if v, ok := job.Metadata["job_application_id"]; ok {
		if s, ok := v.(string); ok && s != "" {
			payload["jobApplicationId"] = s
		}
	}

	// jobDescription and language
	if job.JobDescription != "" {
		payload["jobDescription"] = job.JobDescription
	}
	if v, ok := job.Metadata["language"]; ok {
		if s, ok := v.(string); ok && s != "" {
			payload["language"] = s
		}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return NewDomainError(ErrCodeInvalidPayload, err, map[string]interface{}{
			"jobId": job.JobID,
		})
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.url+"/jobs/start", io.NopCloser(bytesReader(body)))
	if err != nil {
		return NewDomainError(ErrCodeHTTPRequestFailed, err, map[string]interface{}{
			"jobId": job.JobID,
			"url":   p.url,
		})
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		p.logger.Error("resume generator HTTP request failed",
			"error_code", ErrCodeHTTPRequestFailed,
			"jobId", job.JobID,
			"url", p.url,
			"error", err.Error(),
		)
		return NewDomainError(ErrCodeHTTPRequestFailed, err, map[string]interface{}{
			"jobId": job.JobID,
			"url":   p.url,
		})
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		p.logger.Error("resume generator rejected job",
			"error_code", ErrCodePublisherFailed,
			"jobId", job.JobID,
			"status", resp.StatusCode,
			"response_body", string(bodyBytes),
		)
		return NewDomainError(ErrCodePublisherFailed, fmt.Errorf("status %d", resp.StatusCode), map[string]interface{}{
			"jobId":          job.JobID,
			"status":         resp.StatusCode,
			"resume_gen_url": p.url,
		})
	}

	p.logger.Info("resume generation job forwarded to resume-generator",
		"jobId", job.JobID,
		"userId", job.UserID,
	)
	return nil
}

func (p *httpPublisher) Close() error {
    // Nothing to close for HTTP client
    return nil
}

// NewNoOpPublisher returns a publisher that logs a warning and does nothing.
func NewNoOpPublisher(logger *slog.Logger) RabbitMQPublisher {
    return &noOpPublisher{logger: logger}
}

type noOpPublisher struct{
    logger *slog.Logger
}

func (p *noOpPublisher) PublishResumeGenerationJob(ctx context.Context, job *ResumeWorkerJob) error {
    if p.logger != nil {
        p.logger.Warn("No-op publisher: job will not be forwarded", "jobId", job.JobID)
    }
    return nil
}

func (p *noOpPublisher) Close() error { return nil }

// bytesReader returns an io.Reader from a byte slice without importing bytes everywhere.
func bytesReader(b []byte) *bytesReaderWrapper { return &bytesReaderWrapper{b: b} }

type bytesReaderWrapper struct{ b []byte; i int }

func (r *bytesReaderWrapper) Read(p []byte) (int, error) {
    if r.i >= len(r.b) {
        return 0, io.EOF
    }
    n := copy(p, r.b[r.i:])
    r.i += n
    return n, nil
}
