package joblevels

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository defines persistence operations for job levels.
type Repository interface {
	CreateLevel(ctx context.Context, level *JobLevel) error
	GetLevel(ctx context.Context, levelID uuid.UUID) (*JobLevel, error)
	GetLevelByName(ctx context.Context, name string) (*JobLevel, error)
	ListLevels(ctx context.Context) ([]JobLevel, error)
	SeedLevels(ctx context.Context, levels []JobLevel) error
}

type gormRepository struct {
	db *gorm.DB
}

// NewGormRepository returns a GORM-backed repository.
func NewGormRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) CreateLevel(ctx context.Context, level *JobLevel) error {
	return r.db.WithContext(ctx).Create(level).Error
}

func (r *gormRepository) GetLevel(ctx context.Context, levelID uuid.UUID) (*JobLevel, error) {
	var level JobLevel
	if err := r.db.WithContext(ctx).Where("id = ?", levelID).First(&level).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, NewDomainError(ErrCodeNotFound, ErrJobLevelNotFound)
		}
		return nil, NewDomainError(ErrCodeRepositoryFailure, ErrUnableToFetch)
	}
	return &level, nil
}

func (r *gormRepository) GetLevelByName(ctx context.Context, name string) (*JobLevel, error) {
	var level JobLevel
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&level).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, NewDomainError(ErrCodeNotFound, ErrJobLevelNotFound)
		}
		return nil, NewDomainError(ErrCodeRepositoryFailure, ErrUnableToFetch)
	}
	return &level, nil
}

func (r *gormRepository) ListLevels(ctx context.Context) ([]JobLevel, error) {
	var levels []JobLevel
	if err := r.db.WithContext(ctx).Order("seniority ASC, intensity ASC").Find(&levels).Error; err != nil {
		return nil, NewDomainError(ErrCodeRepositoryFailure, ErrUnableToFetch)
	}
	return levels, nil
}

func (r *gormRepository) SeedLevels(ctx context.Context, levels []JobLevel) error {
	// Check if any levels already exist
	var count int64
	if err := r.db.WithContext(ctx).Model(&JobLevel{}).Count(&count).Error; err != nil {
		return NewDomainError(ErrCodeRepositoryFailure, ErrUnableToFetch)
	}

	// If levels already exist, don't seed
	if count > 0 {
		return nil
	}

	// Insert all levels
	for _, level := range levels {
		if err := r.db.WithContext(ctx).Create(&level).Error; err != nil {
			return NewDomainError(ErrCodeRepositoryFailure, ErrUnableToCreate)
		}
	}

	return nil
}
