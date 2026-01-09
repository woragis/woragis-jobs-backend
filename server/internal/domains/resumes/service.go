package resumes

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
)

// Service orchestrates resume workflows.
type Service interface {
	CreateResume(ctx context.Context, userID uuid.UUID, title, filePath, fileName string, fileSize int64, tags JSONArray) (*Resume, error)
	UpdateResume(ctx context.Context, userID uuid.UUID, resumeID uuid.UUID, title string, tags JSONArray) (*Resume, error)
	DeleteResume(ctx context.Context, userID uuid.UUID, resumeID uuid.UUID) error
	GetResume(ctx context.Context, userID uuid.UUID, resumeID uuid.UUID) (*Resume, error)
	ListResumes(ctx context.Context, userID uuid.UUID) ([]Resume, error)
	ListResumesByTags(ctx context.Context, userID uuid.UUID, tags []string) ([]Resume, error)
	MarkAsMain(ctx context.Context, userID uuid.UUID, resumeID uuid.UUID) (*Resume, error)
	MarkAsFeatured(ctx context.Context, userID uuid.UUID, resumeID uuid.UUID) (*Resume, error)
	UnmarkAsMain(ctx context.Context, userID uuid.UUID, resumeID uuid.UUID) (*Resume, error)
	UnmarkAsFeatured(ctx context.Context, userID uuid.UUID, resumeID uuid.UUID) (*Resume, error)
	GetMainResume(ctx context.Context, userID uuid.UUID) (*Resume, error)
	GetFeaturedResume(ctx context.Context, userID uuid.UUID) (*Resume, error)
	GetBestResume(ctx context.Context, userID uuid.UUID) (*Resume, error) // Returns main > featured > most recent
	RecalculateResumeMetrics(ctx context.Context, resumeID uuid.UUID) error
	// Resume generation operations
	GenerateResume(ctx context.Context, userID uuid.UUID, jobDescription string, metadata map[string]interface{}) (jobID uuid.UUID, err error)
	GetResumeGenerationJobStatus(ctx context.Context, jobID uuid.UUID) (*ResumeGenerationJob, error)
	ListUserResumeGenerationJobs(ctx context.Context, userID uuid.UUID) ([]ResumeGenerationJob, error)
	CompleteResumeGeneration(ctx context.Context, jobID uuid.UUID, resumeID uuid.UUID) error
	FailResumeGeneration(ctx context.Context, jobID uuid.UUID, errorMessage string) error
}

// service implements Service.
type service struct {
	repo                 Repository
	rabbitMQPublisher    RabbitMQPublisher
	logger               *slog.Logger
}

// NewService creates a new resume service.
func NewService(repo Repository, publisher RabbitMQPublisher, logger *slog.Logger) Service {
	return &service{
		repo:                 repo,
		rabbitMQPublisher:    publisher,
		logger:               logger,
	}
}

// CreateResume creates a new resume.
func (s *service) CreateResume(ctx context.Context, userID uuid.UUID, title, filePath, fileName string, fileSize int64, tags JSONArray) (*Resume, error) {
	resume, err := NewResume(userID, title, filePath, fileName, fileSize, tags)
	if err != nil {
		return nil, err
	}

	if err := s.repo.CreateResume(ctx, resume); err != nil {
		return nil, err
	}

	return resume, nil
}

// UpdateResume updates an existing resume.
func (s *service) UpdateResume(ctx context.Context, userID uuid.UUID, resumeID uuid.UUID, title string, tags JSONArray) (*Resume, error) {
	resume, err := s.repo.GetResume(ctx, resumeID, userID)
	if err != nil {
		return nil, err
	}

	if title != "" {
		if err := resume.UpdateTitle(title); err != nil {
			return nil, err
		}
	}

	if tags != nil {
		if err := resume.UpdateTags(tags); err != nil {
			return nil, err
		}
	}

	if err := s.repo.UpdateResume(ctx, resume); err != nil {
		return nil, err
	}

	return resume, nil
}

// DeleteResume deletes a resume.
func (s *service) DeleteResume(ctx context.Context, userID uuid.UUID, resumeID uuid.UUID) error {
	return s.repo.DeleteResume(ctx, resumeID, userID)
}

// GetResume retrieves a resume by ID.
func (s *service) GetResume(ctx context.Context, userID uuid.UUID, resumeID uuid.UUID) (*Resume, error) {
	return s.repo.GetResume(ctx, resumeID, userID)
}

// ListResumes lists all resumes for a user.
func (s *service) ListResumes(ctx context.Context, userID uuid.UUID) ([]Resume, error) {
	return s.repo.ListResumes(ctx, userID)
}

// ListResumesByTags lists resumes filtered by tags.
func (s *service) ListResumesByTags(ctx context.Context, userID uuid.UUID, tags []string) ([]Resume, error) {
	return s.repo.ListResumesByTags(ctx, userID, tags)
}

// MarkAsMain marks a resume as main and unmarks others.
func (s *service) MarkAsMain(ctx context.Context, userID uuid.UUID, resumeID uuid.UUID) (*Resume, error) {
	// Unmark all other resumes as main
	if err := s.repo.UnmarkAllAsMain(ctx, userID); err != nil {
		return nil, err
	}

	// Mark this resume as main
	resume, err := s.repo.GetResume(ctx, resumeID, userID)
	if err != nil {
		return nil, err
	}

	resume.MarkAsMain()
	if err := s.repo.UpdateResume(ctx, resume); err != nil {
		return nil, err
	}

	return resume, nil
}

// MarkAsFeatured marks a resume as featured.
func (s *service) MarkAsFeatured(ctx context.Context, userID uuid.UUID, resumeID uuid.UUID) (*Resume, error) {
	resume, err := s.repo.GetResume(ctx, resumeID, userID)
	if err != nil {
		return nil, err
	}

	resume.MarkAsFeatured()
	if err := s.repo.UpdateResume(ctx, resume); err != nil {
		return nil, err
	}

	return resume, nil
}

// UnmarkAsMain removes the main flag from a resume.
func (s *service) UnmarkAsMain(ctx context.Context, userID uuid.UUID, resumeID uuid.UUID) (*Resume, error) {
	resume, err := s.repo.GetResume(ctx, resumeID, userID)
	if err != nil {
		return nil, err
	}

	resume.UnmarkAsMain()
	if err := s.repo.UpdateResume(ctx, resume); err != nil {
		return nil, err
	}

	return resume, nil
}

// UnmarkAsFeatured removes the featured flag from a resume.
func (s *service) UnmarkAsFeatured(ctx context.Context, userID uuid.UUID, resumeID uuid.UUID) (*Resume, error) {
	resume, err := s.repo.GetResume(ctx, resumeID, userID)
	if err != nil {
		return nil, err
	}

	resume.UnmarkAsFeatured()
	if err := s.repo.UpdateResume(ctx, resume); err != nil {
		return nil, err
	}

	return resume, nil
}

// GetMainResume retrieves the main resume.
func (s *service) GetMainResume(ctx context.Context, userID uuid.UUID) (*Resume, error) {
	return s.repo.GetMainResume(ctx, userID)
}

// GetFeaturedResume retrieves a featured resume.
func (s *service) GetFeaturedResume(ctx context.Context, userID uuid.UUID) (*Resume, error) {
	return s.repo.GetFeaturedResume(ctx, userID)
}

// GetBestResume returns the best resume using priority: main > featured > most recent.
func (s *service) GetBestResume(ctx context.Context, userID uuid.UUID) (*Resume, error) {
	// Try main first
	resume, err := s.repo.GetMainResume(ctx, userID)
	if err == nil {
		return resume, nil
	}

	// Try featured
	resume, err = s.repo.GetFeaturedResume(ctx, userID)
	if err == nil {
		return resume, nil
	}

	// Fallback to most recent
	resumes, err := s.repo.ListResumes(ctx, userID)
	if err != nil {
		return nil, err
	}

	if len(resumes) == 0 {
		return nil, NewDomainError(ErrCodeNotFound, ErrResumeNotFound)
	}

	// Return the most recent (first in list since it's ordered by created_at DESC)
	return &resumes[0], nil
}

// RecalculateResumeMetrics recalculates and updates metrics for a resume.
func (s *service) RecalculateResumeMetrics(ctx context.Context, resumeID uuid.UUID) error {
	metrics, err := s.repo.CalculateResumeMetrics(ctx, resumeID)
	if err != nil {
		return err
	}
	
	return s.repo.UpdateResumeMetrics(ctx, resumeID, metrics)
}

// GenerateResume creates a resume generation job and publishes it to the queue.
func (s *service) GenerateResume(ctx context.Context, userID uuid.UUID, jobDescription string, metadata map[string]interface{}) (uuid.UUID, error) {
	// Create a new resume generation job
	job := NewResumeGenerationJob(userID, jobDescription, metadata)
	
	// Persist the job to the database
	if err := s.repo.CreateResumeGenerationJob(ctx, job); err != nil {
		s.logger.Error("failed to create resume generation job", "error", err, "userId", userID)
		return uuid.Nil, err
	}
	
	// Convert to ResumeWorkerJob for publishing
	workerJob := &ResumeWorkerJob{
		JobID:          job.ID.String(),
		UserID:         job.UserID.String(),
		JobDescription: job.JobDescription,
		Metadata:       job.Metadata,
	}
	
	// Publish the job to RabbitMQ for the worker to process
	if err := s.rabbitMQPublisher.PublishResumeGenerationJob(ctx, workerJob); err != nil {
		s.logger.Error("failed to publish resume generation job", "error", err, "jobId", job.ID)
		// Mark the job as failed since we couldn't queue it
		job.MarkFailed("Failed to queue job for processing", "QUEUE_ERROR")
		_ = s.repo.UpdateResumeGenerationJob(ctx, job)
		return uuid.Nil, err
	}
	
	s.logger.Info("resume generation job created and queued", "jobId", job.ID, "userId", userID)
	return job.ID, nil
}

// GetResumeGenerationJobStatus retrieves the status of a resume generation job.
func (s *service) GetResumeGenerationJobStatus(ctx context.Context, jobID uuid.UUID) (*ResumeGenerationJob, error) {
	job, err := s.repo.GetResumeGenerationJob(ctx, jobID)
	if err != nil {
		s.logger.Error("failed to get resume generation job status", "error", err, "jobId", jobID)
		return nil, err
	}
	
	return job, nil
}

// ListUserResumeGenerationJobs retrieves all resume generation jobs for a user.
func (s *service) ListUserResumeGenerationJobs(ctx context.Context, userID uuid.UUID) ([]ResumeGenerationJob, error) {
	jobs, err := s.repo.ListUserResumeGenerationJobs(ctx, userID)
	if err != nil {
		s.logger.Error("failed to list user resume generation jobs", "error", err, "userId", userID)
		return nil, err
	}
	
	return jobs, nil
}

// CompleteResumeGeneration marks a resume generation job as completed with the generated resume ID.
func (s *service) CompleteResumeGeneration(ctx context.Context, jobID uuid.UUID, resumeID uuid.UUID) error {
	job, err := s.repo.GetResumeGenerationJob(ctx, jobID)
	if err != nil {
		s.logger.Error("failed to get resume generation job", "error", err, "jobId", jobID)
		return err
	}
	
	job.MarkCompleted(resumeID)
	if err := s.repo.UpdateResumeGenerationJob(ctx, job); err != nil {
		s.logger.Error("failed to update resume generation job", "error", err, "jobId", jobID)
		return err
	}
	
	s.logger.Info("resume generation job completed", "jobId", jobID, "resumeId", resumeID)
	return nil
}

// FailResumeGeneration marks a resume generation job as failed with an error message.
func (s *service) FailResumeGeneration(ctx context.Context, jobID uuid.UUID, errorMessage string) error {
	job, err := s.repo.GetResumeGenerationJob(ctx, jobID)
	if err != nil {
		s.logger.Error("failed to get resume generation job", "error", err, "jobId", jobID)
		return err
	}
	
	// Use a generic error code if none provided
	job.MarkFailed(errorMessage, "GENERATION_ERROR")
	if err := s.repo.UpdateResumeGenerationJob(ctx, job); err != nil {
		s.logger.Error("failed to update resume generation job", "error", err, "jobId", jobID)
		return err
	}
	
	s.logger.Info("resume generation job failed", "jobId", jobID, "error", errorMessage)
	return nil
}

