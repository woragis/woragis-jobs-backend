package resumes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQPublisher publishes resume generation jobs to RabbitMQ for the resume worker.
type RabbitMQPublisher interface {
	PublishResumeGenerationJob(ctx context.Context, job *ResumeWorkerJob) error
	Close() error
}

// ResumeWorkerJob is the message published to RabbitMQ for the resume worker.
type ResumeWorkerJob struct {
	JobID          string                 `json:"jobId"`
	UserID         string                 `json:"userId"`
	UserEmail      string                 `json:"userEmail"`
	UserName       string                 `json:"userName"`
	JobDescription string                 `json:"jobDescription"`
	Metadata       map[string]interface{} `json:"metadata"`
}

const (
	resumeExchange   = "woragis.tasks"
	resumeQueue      = "resumes.queue"
	resumeRoutingKey = "resumes.generate"
)

type rabbitMQPublisher struct {
	channel *amqp.Channel
	confirms <-chan amqp.Confirmation
	logger  *slog.Logger
}

// NewRabbitMQPublisher creates a new RabbitMQ publisher for resume jobs.
func NewRabbitMQPublisher(channel *amqp.Channel, logger *slog.Logger) (RabbitMQPublisher, error) {
	// Declare exchange
	err := channel.ExchangeDeclare(
		resumeExchange,  // name
		"direct",        // kind
		true,            // durable
		false,           // auto-deleted
		false,           // internal
		false,           // no-wait
		nil,             // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Declare dead-letter exchange
	err = channel.ExchangeDeclare(
		"woragis.dlx", // name
		"direct",      // kind
		true,          // durable
		false,         // auto-deleted
		false,         // internal
		false,         // no-wait
		nil,           // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare dead-letter exchange: %w", err)
	}

	// Declare queue with DLX configuration (must match resume-worker)
	_, err = channel.QueueDeclare(
		resumeQueue, // name
		true,        // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		amqp.Table{
			"x-max-priority":            10,
			"x-dead-letter-exchange":    "woragis.dlx",
			"x-dead-letter-routing-key": "resumes.dead-letter",
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind queue to exchange
	err = channel.QueueBind(
		resumeQueue,      // queue name
		resumeRoutingKey, // routing key
		resumeExchange,   // exchange
		false,            // no-wait
		nil,              // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to bind queue: %w", err)
	}

	// Enable publisher confirms on the channel so we can wait for ack/nack
	if err := channel.Confirm(false); err != nil {
		// Not fatal — some environments may not support confirms, but prefer to log
		logger.Warn("failed to enable publisher confirms", "error", err.Error())
	}

	confirms := channel.NotifyPublish(make(chan amqp.Confirmation, 1))

	return &rabbitMQPublisher{
		channel:  channel,
		confirms: confirms,
		logger:   logger,
	}, nil
}

// helper to get publish retry configuration from env
func getPublishConfig() (maxAttempts int, timeout time.Duration, baseDelay time.Duration) {
	maxAttempts = 3
	if v := os.Getenv("JOBS_PUBLISH_MAX_ATTEMPTS"); v != "" {
		if i, err := strconv.Atoi(v); err == nil && i > 0 {
			maxAttempts = i
		}
	}

	timeout = 5 * time.Second
	if v := os.Getenv("JOBS_PUBLISH_WAIT_ACK_MS"); v != "" {
		if i, err := strconv.Atoi(v); err == nil && i > 0 {
			timeout = time.Duration(i) * time.Millisecond
		}
	}

	baseDelay = 500 * time.Millisecond
	if v := os.Getenv("JOBS_PUBLISH_BASE_DELAY_MS"); v != "" {
		if i, err := strconv.Atoi(v); err == nil && i > 0 {
			baseDelay = time.Duration(i) * time.Millisecond
		}
	}

	return
}

// PublishResumeGenerationJob publishes a resume generation job to RabbitMQ.
func (p *rabbitMQPublisher) PublishResumeGenerationJob(ctx context.Context, job *ResumeWorkerJob) error {
	if p.channel == nil {
		return fmt.Errorf("channel is not available")
	}

	body, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	// Retry loop with publisher confirms if available
	maxAttempts, ackWait, baseDelay := getPublishConfig()

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if attempt > 1 {
			// Exponential backoff
			delay := time.Duration(attempt) * baseDelay
			time.Sleep(delay)
		}

		// Publish
		err = p.channel.PublishWithContext(
			ctx,
			resumeExchange,   // exchange
			resumeRoutingKey, // routing key
			false,            // mandatory
			false,            // immediate
			amqp.Publishing{
				ContentType:  "application/json",
				Body:         body,
				DeliveryMode: amqp.Persistent,
			},
		)

		if err != nil {
			lastErr = err
			p.logger.Error("failed to publish job",
				slog.String("jobId", job.JobID),
				slog.String("error", err.Error()),
				slog.Int("attempt", attempt),
			)
			continue
		}

		// If confirms channel is available, wait for confirmation
		if p.confirms != nil {
			select {
			case conf := <-p.confirms:
				if conf.Ack {
					p.logger.Info("resume generation job published",
						slog.String("jobId", job.JobID),
						slog.String("userId", job.UserID),
					)
					return nil
				}
				lastErr = fmt.Errorf("message nack received")
				p.logger.Warn("publish received NACK from broker", "jobId", job.JobID, "attempt", attempt)
				continue
			case <-time.After(ackWait):
				lastErr = fmt.Errorf("timeout waiting for publish confirmation")
				p.logger.Warn("timeout waiting for publish confirmation", "jobId", job.JobID, "attempt", attempt)
				continue
			}
		}

		// If no confirms available, consider publish successful
		p.logger.Info("resume generation job published (no confirms)",
			slog.String("jobId", job.JobID),
			slog.String("userId", job.UserID),
		)
		return nil
	}

	// All attempts failed — attempt HTTP fallback to resume-worker if configured
	p.logger.Error("all publish attempts failed, attempting HTTP fallback if configured", "jobId", job.JobID, "lastError", fmt.Sprintf("%v", lastErr))

	fallbackURL := os.Getenv("RESUME_WORKER_FALLBACK_URL")
	if fallbackURL == "" {
		fallbackURL = "http://resume-worker:3005/fallback/resumes"
	}

	reqBody, _ := json.Marshal(job)
	httpClient := &http.Client{Timeout: 10 * time.Second}

	resp, err := httpClient.Post(fallbackURL, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		p.logger.Error("failed to call HTTP fallback", "error", err.Error(), "fallbackURL", fallbackURL)
		if lastErr != nil {
			return fmt.Errorf("publish failed: %w", lastErr)
		}
		return fmt.Errorf("publish failed and fallback failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		p.logger.Info("fallback accepted resume generation job", "jobId", job.JobID, "fallbackURL", fallbackURL)
		return nil
	}

	p.logger.Error("fallback rejected resume generation job", "status", resp.StatusCode, "fallbackURL", fallbackURL)
	if lastErr != nil {
		return fmt.Errorf("publish failed: %w", lastErr)
	}
	return fmt.Errorf("publish failed and fallback returned status %d", resp.StatusCode)
}

// Close closes the RabbitMQ channel.
func (p *rabbitMQPublisher) Close() error {
	if p.channel != nil {
		return p.channel.Close()
	}
	return nil
}

// NoOpPublisher is a no-op RabbitMQ publisher for when RabbitMQ is not available
type noOpPublisher struct {
	logger *slog.Logger
}

// NewNoOpPublisher creates a new no-op RabbitMQ publisher
func NewNoOpPublisher(logger *slog.Logger) RabbitMQPublisher {
	return &noOpPublisher{logger: logger}
}

// PublishResumeGenerationJob is a no-op implementation
func (p *noOpPublisher) PublishResumeGenerationJob(ctx context.Context, job *ResumeWorkerJob) error {
	p.logger.Warn("RabbitMQ publisher is not available, job will not be queued",
		slog.String("jobId", job.JobID),
		slog.String("userId", job.UserID),
	)
	return nil
}

// Close is a no-op implementation
func (p *noOpPublisher) Close() error {
	return nil
}
