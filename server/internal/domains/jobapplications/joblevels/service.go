package joblevels

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
)

// Service handles business logic for job levels.
type Service interface {
	GetLevel(ctx context.Context, levelID string) (*JobLevel, error)
	ListLevels(ctx context.Context) ([]JobLevel, error)
	SeedDefaultLevels(ctx context.Context) error
}

type service struct {
	repo   Repository
	logger *slog.Logger
}

// NewService creates a new job level service.
func NewService(repo Repository, logger *slog.Logger) Service {
	return &service{
		repo:   repo,
		logger: logger,
	}
}

func (s *service) GetLevel(ctx context.Context, levelID string) (*JobLevel, error) {
	// Parse string to UUID
	parsedID, err := uuid.Parse(levelID)
	if err != nil {
		return nil, NewDomainError(ErrCodeInvalidPayload, err, map[string]interface{}{
			"field":  "level_id",
			"reason": "invalid UUID format",
		})
	}

	return s.repo.GetLevel(ctx, parsedID)
}

func (s *service) ListLevels(ctx context.Context) ([]JobLevel, error) {
	return s.repo.ListLevels(ctx)
}

func (s *service) SeedDefaultLevels(ctx context.Context) error {
	levels := PredefinedLevels()
	if err := s.repo.SeedLevels(ctx, levels); err != nil {
		s.logger.Error("failed to seed job levels", "error", err)
		return err
	}
	s.logger.Info("job levels seeded successfully", "count", len(levels))
	return nil
}
