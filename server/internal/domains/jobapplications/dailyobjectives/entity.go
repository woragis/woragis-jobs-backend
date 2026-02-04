package dailyobjectives

import (
	"time"

	"github.com/google/uuid"
)

// DailyObjective represents a user's daily application targets.
type DailyObjective struct {
	ID           uuid.UUID `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID       uuid.UUID `gorm:"column:user_id;type:uuid;index;not null;uniqueIndex" json:"userId"`
	TotalTarget  int       `gorm:"column:total_target;not null" json:"totalTarget"`
	JuniorTarget int       `gorm:"column:junior_target;not null" json:"juniorTarget"`
	PlenoTarget  int       `gorm:"column:pleno_target;not null" json:"plenoTarget"`
	SeniorTarget int       `gorm:"column:senior_target;not null" json:"seniorTarget"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
}

// TableName returns the table name for DailyObjective.
func (DailyObjective) TableName() string {
	return "daily_objectives"
}

// DailyStats represents the count of applications for a specific date.
type DailyStats struct {
	Date       time.Time `json:"date"`
	TotalCount int       `json:"totalCount"`
	JuniorCount int       `json:"juniorCount"`
	PlenoCount  int       `json:"plenoCount"`
	SeniorCount int       `json:"seniorCount"`
}

// DailyProgress combines stats with targets and calculates progress percentages.
type DailyProgress struct {
	Date            time.Time      `json:"date"`
	Stats           DailyStats     `json:"stats"`
	Targets         DailyObjective `json:"targets"`
	ProgressTotal   float64        `json:"progressTotal"`   // 0-100
	ProgressJunior  float64        `json:"progressJunior"`  // 0-100
	ProgressPleno   float64        `json:"progressPleno"`   // 0-100
	ProgressSenior  float64        `json:"progressSenior"`  // 0-100
}
