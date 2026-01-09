package jobs

import (
	"gorm.io/gorm"

	"woragis-jobs-service/internal/domains/jobapplications"
	"woragis-jobs-service/internal/domains/jobapplications/responses"
	"woragis-jobs-service/internal/domains/jobapplications/interviewstages"
	"woragis-jobs-service/internal/domains/resumes"
	"woragis-jobs-service/internal/domains/jobwebsites"
)

// MigrateJobsTables runs database migrations for jobs service
func MigrateJobsTables(db *gorm.DB) error {
	// Enable UUID extension if not already enabled
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		return err
	}

	// Enable gen_random_uuid function if not already available
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"pgcrypto\"").Error; err != nil {
		return err
	}

	// Migrate job applications tables
	if err := db.AutoMigrate(
		&jobapplications.JobApplication{},
	); err != nil {
		return err
	}

	// Migrate resumes tables
	if err := db.AutoMigrate(
		&resumes.Resume{},
		&resumes.ResumeGenerationJob{},
	); err != nil {
		return err
	}

	// Migrate job websites tables
	if err := db.AutoMigrate(
		&jobwebsites.JobWebsite{},
	); err != nil {
		return err
	}

	// Migrate subdomain tables
	if err := db.AutoMigrate(
		&responses.Response{},
		&interviewstages.InterviewStage{},
	); err != nil {
		return err
	}

	return nil
}
