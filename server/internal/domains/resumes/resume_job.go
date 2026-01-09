package resumes

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ResumeJobStatus represents the status of a resume generation job
type ResumeJobStatus string

const (
	ResumeJobStatusPending    ResumeJobStatus = "pending"
	ResumeJobStatusProcessing ResumeJobStatus = "processing"
	ResumeJobStatusCompleted  ResumeJobStatus = "completed"
	ResumeJobStatusFailed     ResumeJobStatus = "failed"
	ResumeJobStatusCancelled  ResumeJobStatus = "cancelled"
)

// ResumeGenerationJob tracks a resume generation request
type ResumeGenerationJob struct {
	ID             uuid.UUID       `gorm:"column:id;type:uuid;primaryKey" json:"id"`
	UserID         uuid.UUID       `gorm:"column:user_id;type:uuid;index;not null" json:"userId"`
	JobDescription string          `gorm:"column:job_description;type:text;not null" json:"jobDescription"`
	Status         ResumeJobStatus `gorm:"column:status;type:varchar(50);default:'pending'" json:"status"`
	Metadata       JSONMetadata    `gorm:"column:metadata;type:jsonb;default:'{}'" json:"metadata"`
	ErrorMessage   string          `gorm:"column:error_message;type:text" json:"errorMessage,omitempty"`
	ErrorCode      string          `gorm:"column:error_code;type:varchar(50)" json:"errorCode,omitempty"`
	ResumeID       *uuid.UUID      `gorm:"column:resume_id;type:uuid;index" json:"resumeId,omitempty"`
	CreatedAt      time.Time       `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt      time.Time       `gorm:"column:updated_at" json:"updatedAt"`
}

// TableName specifies the table name for ResumeGenerationJob
func (ResumeGenerationJob) TableName() string {
	return "resume_jobs"
}

// JSONMetadata is a custom type for storing JSON metadata in PostgreSQL
type JSONMetadata map[string]interface{}

// Value implements the driver.Valuer interface
func (j JSONMetadata) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface
func (j *JSONMetadata) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONMetadata)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return json.Unmarshal([]byte(value.(string)), j)
	}
	return json.Unmarshal(bytes, j)
}

// NewResumeGenerationJob creates a new resume generation job
func NewResumeGenerationJob(userID uuid.UUID, jobDescription string, metadata map[string]interface{}) *ResumeGenerationJob {
	md := make(JSONMetadata)
	for k, v := range metadata {
		md[k] = v
	}
	
	return &ResumeGenerationJob{
		ID:             uuid.New(),
		UserID:         userID,
		JobDescription: jobDescription,
		Status:         ResumeJobStatusPending,
		Metadata:       md,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}
}

// MarkProcessing marks the job as being processed
func (j *ResumeGenerationJob) MarkProcessing() {
	j.Status = ResumeJobStatusProcessing
	j.UpdatedAt = time.Now().UTC()
}

// MarkCompleted marks the job as completed and links the generated resume
func (j *ResumeGenerationJob) MarkCompleted(resumeID uuid.UUID) {
	j.Status = ResumeJobStatusCompleted
	j.ResumeID = &resumeID
	j.ErrorMessage = ""
	j.ErrorCode = ""
	j.UpdatedAt = time.Now().UTC()
}

// MarkFailed marks the job as failed with an error message
func (j *ResumeGenerationJob) MarkFailed(errorMessage string, errorCode string) {
	j.Status = ResumeJobStatusFailed
	j.ErrorMessage = errorMessage
	j.ErrorCode = errorCode
	j.UpdatedAt = time.Now().UTC()
}

// MarkCancelled marks the job as cancelled
func (j *ResumeGenerationJob) MarkCancelled() {
	j.Status = ResumeJobStatusCancelled
	j.UpdatedAt = time.Now().UTC()
}
