package joblevels

import (
	"time"

	"github.com/google/uuid"
)

// Seniority represents the seniority level of a job.
type Seniority string

const (
	SeniorityEntry  Seniority = "entry"
	SeniorityJunior Seniority = "junior"
	SeniorityMid    Seniority = "mid"
	SeniorityPleno  Seniority = "pleno"
	SenioritySenior Seniority = "senior"
)

// Intensity represents the intensity/difficulty level of a job.
type Intensity string

const (
	IntensityLow    Intensity = "low"
	IntensityMedium Intensity = "medium"
	IntensityHigh   Intensity = "high"
)

// JobLevel represents a job level classification combining seniority and intensity.
type JobLevel struct {
	ID        uuid.UUID `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name      string    `gorm:"column:name;type:varchar(50);uniqueIndex;not null" json:"name"`
	Seniority Seniority `gorm:"column:seniority;type:varchar(20);not null" json:"seniority"`
	Intensity Intensity `gorm:"column:intensity;type:varchar(20);not null" json:"intensity"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
}

// TableName returns the table name for JobLevel.
func (JobLevel) TableName() string {
	return "job_levels"
}

// ValidSeniorities returns all valid seniority levels.
func ValidSeniorities() []Seniority {
	return []Seniority{
		SeniorityEntry,
		SeniorityJunior,
		SeniorityMid,
		SeniorityPleno,
		SenioritySenior,
	}
}

// ValidIntensities returns all valid intensity levels.
func ValidIntensities() []Intensity {
	return []Intensity{
		IntensityLow,
		IntensityMedium,
		IntensityHigh,
	}
}

// PredefinedLevels returns all predefined job levels.
func PredefinedLevels() []JobLevel {
	levels := []JobLevel{}
	for _, seniority := range ValidSeniorities() {
		for _, intensity := range ValidIntensities() {
			levels = append(levels, JobLevel{
				Name:      string(seniority) + "_" + string(intensity),
				Seniority: seniority,
				Intensity: intensity,
			})
		}
	}
	return levels
}
