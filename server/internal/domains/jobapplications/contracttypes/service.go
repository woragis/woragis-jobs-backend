package contracttypes

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
)

// Service handles business logic for contract types.
type Service interface {
	GetType(ctx context.Context, typeID string) (*ContractType, error)
	ListTypes(ctx context.Context) ([]ContractType, error)
	SeedDefaultTypes(ctx context.Context) error
}

type service struct {
	repo   Repository
	logger *slog.Logger
}

// NewService creates a new contract type service.
func NewService(repo Repository, logger *slog.Logger) Service {
	return &service{
		repo:   repo,
		logger: logger,
	}
}

func (s *service) GetType(ctx context.Context, typeID string) (*ContractType, error) {
	// Parse string to UUID
	parsedID, err := uuid.Parse(typeID)
	if err != nil {
		return nil, NewDomainError(ErrCodeInvalidPayload, ErrInvalidName)
	}

	return s.repo.GetType(ctx, parsedID)
}

func (s *service) ListTypes(ctx context.Context) ([]ContractType, error) {
	return s.repo.ListTypes(ctx)
}

func (s *service) SeedDefaultTypes(ctx context.Context) error {
	types := PredefinedContractTypes()
	if err := s.repo.SeedTypes(ctx, types); err != nil {
		s.logger.Error("failed to seed contract types", "error", err)
		return err
	}
	s.logger.Info("contract types seeded successfully", "count", len(types))
	return nil
}
