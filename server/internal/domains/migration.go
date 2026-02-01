package jobs

import (
	"gorm.io/gorm"

	"woragis-jobs-service/internal/domains/jobapplications"
	"woragis-jobs-service/internal/domains/jobapplications/interviewstages"
	"woragis-jobs-service/internal/domains/jobapplications/responses"
	"woragis-jobs-service/internal/domains/jobwebsites"
	"woragis-jobs-service/internal/domains/resumes"
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

	// Ensure language column can store freeform language strings (upcast to varchar(100)).
	// This is idempotent for Postgres: if the column is already large enough the ALTER will be a no-op.
	// Use a safe ALTER TYPE statement wrapped in Exec so it runs in production automatically.
	if err := db.Exec("ALTER TABLE job_applications ALTER COLUMN language TYPE varchar(100)").Error; err != nil {
		// Some environments may not have the column yet or may already match; ignore errors that indicate
		// missing column but propagate others.
		// We'll attempt a safer conditional check for Postgres by verifying column existence first.
		// Try to detect a missing column error by running a conditional ALTER only if column exists.
		if execErr := db.Exec(`DO $$
BEGIN
	IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='job_applications' AND column_name='language') THEN
		EXECUTE 'ALTER TABLE job_applications ALTER COLUMN language TYPE varchar(100)';
	END IF;
END$$;`).Error; execErr != nil {
			return execErr
		}
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
