package dailyobjectives

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository defines persistence operations for daily objectives.
type Repository interface {
	CreateObjective(ctx context.Context, objective *DailyObjective) error
	GetObjective(ctx context.Context, userID uuid.UUID) (*DailyObjective, error)
	UpdateObjective(ctx context.Context, objective *DailyObjective) error
	DeleteObjective(ctx context.Context, userID uuid.UUID) error
}

type gormRepository struct {
	db *gorm.DB
}

// NewGormRepository returns a GORM-backed repository.
func NewGormRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) CreateObjective(ctx context.Context, objective *DailyObjective) error {
	if err := r.db.WithContext(ctx).Create(objective).Error; err != nil {
		return NewDomainError(ErrCodeRepositoryFailure, err)
	}
	return nil
}

func (r *gormRepository) GetObjective(ctx context.Context, userID uuid.UUID) (*DailyObjective, error) {
	var objective DailyObjective
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&objective).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, NewDomainError(ErrCodeNotFound, err, map[string]interface{}{
				"user_id": userID.String(),
			})
		}
		return nil, NewDomainError(ErrCodeRepositoryFailure, err)
	}
	return &objective, nil
}

func (r *gormRepository) UpdateObjective(ctx context.Context, objective *DailyObjective) error {
	result := r.db.WithContext(ctx).Model(&DailyObjective{}).Where("user_id = ?", objective.UserID).Updates(objective)
	if result.Error != nil {
		return NewDomainError(ErrCodeRepositoryFailure, result.Error)
	}
	if result.RowsAffected == 0 {
		return NewDomainError(ErrCodeNotFound, nil, map[string]interface{}{
			"user_id": objective.UserID.String(),
		})
	}
	return nil
}

func (r *gormRepository) DeleteObjective(ctx context.Context, userID uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&DailyObjective{})
	if result.Error != nil {
		return NewDomainError(ErrCodeRepositoryFailure, result.Error)
	   }
	return nil
}
