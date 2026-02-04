package contracttypes

import (
	"time"

	"github.com/google/uuid"
)

// ContractType represents an employment contract type.
type ContractType struct {
	ID          uuid.UUID `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"column:name;type:varchar(50);uniqueIndex;not null" json:"name"`
	Description string    `gorm:"column:description;type:text" json:"description"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
}

// TableName returns the table name for ContractType.
func (ContractType) TableName() string {
	return "contract_types"
}

// PredefinedContractTypes returns all predefined contract types.
func PredefinedContractTypes() []ContractType {
	return []ContractType{
		{
			Name:        "full_time",
			Description: "Full-time employment contract",
		},
		{
			Name:        "part_time",
			Description: "Part-time employment contract",
		},
		{
			Name:        "internship",
			Description: "Internship contract",
		},
		{
			Name:        "contractor",
			Description: "Independent contractor",
		},
		{
			Name:        "freelance",
			Description: "Freelance work",
		},
	}
}
