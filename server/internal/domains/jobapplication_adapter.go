package jobs

import (
	"context"

	"github.com/google/uuid"

	"woragis-jobs-service/internal/domains/jobapplications"
	"woragis-jobs-service/internal/domains/resumes"
)

// jobApplicationServiceAdapter adapts the jobapplications.Service to the resumes.JobApplicationService interface
type jobApplicationServiceAdapter struct {
	service jobapplications.Service
}

// newJobApplicationServiceAdapter creates a new adapter
func newJobApplicationServiceAdapter(service jobapplications.Service) resumes.JobApplicationService {
	return &jobApplicationServiceAdapter{
		service: service,
	}
}

// GetJobApplication fetches a job application and converts it to the resume's expected format
func (a *jobApplicationServiceAdapter) GetJobApplication(ctx context.Context, applicationID uuid.UUID) (*resumes.JobApplication, error) {
	jobApp, err := a.service.GetJobApplication(ctx, applicationID)
	if err != nil {
		return nil, err
	}

	// Convert from jobapplications.JobApplication to resumes.JobApplication
	return &resumes.JobApplication{
		ID:             jobApp.ID,
		UserID:         jobApp.UserID,
		JobTitle:       jobApp.JobTitle,
		JobDescription: jobApp.JobDescription,
		Language:       jobApp.Language,
		CompanyName:    jobApp.CompanyName,
	}, nil
}

// UpdateJobApplicationResumeID updates the resume ID for a job application
func (a *jobApplicationServiceAdapter) UpdateJobApplicationResumeID(ctx context.Context, applicationID uuid.UUID, resumeID uuid.UUID) error {
	updates := jobapplications.UpdateJobApplicationRequest{
		ResumeID: &resumeID,
	}
	_, err := a.service.UpdateJobApplication(ctx, applicationID, updates)
	return err
}
