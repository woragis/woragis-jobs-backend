package contracttypes

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository defines persistence operations for contract types.
type Repository interface {
	CreateType(ctx context.Context, contractType *ContractType) error
	GetType(ctx context.Context, typeID uuid.UUID) (*ContractType, error)
	GetTypeByName(ctx context.Context, name string) (*ContractType, error)
	ListTypes(ctx context.Context) ([]ContractType, error)
	SeedTypes(ctx context.Context, types []ContractType) error
}

type gormRepository struct {
	db *gorm.DB
}

// NewGormRepository returns a GORM-backed repository.
func NewGormRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) CreateType(ctx context.Context, contractType *ContractType) error {
	return r.db.WithContext(ctx).Create(contractType).Error
}

func (r *gormRepository) GetType(ctx context.Context, typeID uuid.UUID) (*ContractType, error) {
	var contractType ContractType
	if err := r.db.WithContext(ctx).Where("id = ?", typeID).First(&contractType).Error; err != nil {
		   if errors.Is(err, gorm.ErrRecordNotFound) {
			   return nil, NewDomainError(ErrCodeNotFound, ErrContractTypeNotFound)
		   }
		   return nil, NewDomainError(ErrCodeRepositoryFailure, ErrUnableToFetch)
	}
	return &contractType, nil
}

func (r *gormRepository) GetTypeByName(ctx context.Context, name string) (*ContractType, error) {
	var contractType ContractType
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&contractType).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, NewDomainError(ErrCodeNotFound, ErrContractTypeNotFound)
		}
		return nil, NewDomainError(ErrCodeRepositoryFailure, ErrUnableToFetch)
	}
	return &contractType, nil
}

func (r *gormRepository) ListTypes(ctx context.Context) ([]ContractType, error) {
	var types []ContractType
	if err := r.db.WithContext(ctx).Order("name ASC").Find(&types).Error; err != nil {
		return nil, NewDomainError(ErrCodeRepositoryFailure, ErrUnableToFetch)
	}
	return types, nil
}

func (r *gormRepository) SeedTypes(ctx context.Context, types []ContractType) error {
	// Check if any types already exist
	var count int64
	if err := r.db.WithContext(ctx).Model(&ContractType{}).Count(&count).Error; err != nil {
		return NewDomainError(ErrCodeRepositoryFailure, ErrUnableToFetch)
	}

	// If types already exist, don't seed
	if count > 0 {
		return nil
	}

	// Insert all types
	   for _, contractType := range types {
		   if err := r.db.WithContext(ctx).Create(&contractType).Error; err != nil {
			   return NewDomainError(ErrCodeRepositoryFailure, ErrUnableToCreate)
		   }
	   }

	return nil
}
