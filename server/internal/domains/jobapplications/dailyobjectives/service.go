package dailyobjectives

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Service handles business logic for daily objectives and progress computation.
type Service interface {
	CreateObjective(ctx context.Context, userID uuid.UUID, req CreateObjectiveRequest) (*DailyObjective, error)
	GetObjective(ctx context.Context, userID uuid.UUID) (*DailyObjective, error)
	UpdateObjective(ctx context.Context, userID uuid.UUID, req CreateObjectiveRequest) (*DailyObjective, error)
	GetTodayProgress(ctx context.Context, userID uuid.UUID) (*DailyProgress, error)
	GetHistoricalProgress(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]DailyProgress, error)
}

type CreateObjectiveRequest struct {
	TotalTarget  int `json:"totalTarget"`
	JuniorTarget int `json:"juniorTarget"`
	PlenoTarget  int `json:"plenoTarget"`
	SeniorTarget int `json:"seniorTarget"`
}

type service struct {
	repo Repository
	db   *gorm.DB
}

// NewService creates a new daily objectives service.
func NewService(repo Repository, db *gorm.DB) Service {
	return &service{
		repo: repo,
		db:   db,
	}
}

func (s *service) CreateObjective(ctx context.Context, userID uuid.UUID, req CreateObjectiveRequest) (*DailyObjective, error) {
	// Validate request
	if err := validateRequest(req); err != nil {
		return nil, err
	}

	objective := &DailyObjective{
		ID:           uuid.New(),
		UserID:       userID,
		TotalTarget:  req.TotalTarget,
		JuniorTarget: req.JuniorTarget,
		PlenoTarget:  req.PlenoTarget,
		SeniorTarget: req.SeniorTarget,
	}

	if err := s.repo.CreateObjective(ctx, objective); err != nil {
		return nil, err
	}

	return objective, nil
}

func (s *service) GetObjective(ctx context.Context, userID uuid.UUID) (*DailyObjective, error) {
	return s.repo.GetObjective(ctx, userID)
}

func (s *service) UpdateObjective(ctx context.Context, userID uuid.UUID, req CreateObjectiveRequest) (*DailyObjective, error) {
	// Validate request
	if err := validateRequest(req); err != nil {
		return nil, err
	}

	objective := &DailyObjective{
		UserID:       userID,
		TotalTarget:  req.TotalTarget,
		JuniorTarget: req.JuniorTarget,
		PlenoTarget:  req.PlenoTarget,
		SeniorTarget: req.SeniorTarget,
	}

	if err := s.repo.UpdateObjective(ctx, objective); err != nil {
		return nil, err
	}

	// Fetch updated objective
	return s.repo.GetObjective(ctx, userID)
}

func (s *service) GetTodayProgress(ctx context.Context, userID uuid.UUID) (*DailyProgress, error) {
	today := time.Now().UTC()
	todayStart := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	tomorrowStart := todayStart.AddDate(0, 0, 1)

	stats, err := s.computeStats(ctx, userID, todayStart, tomorrowStart)
	if err != nil {
		return nil, err
	}

	objective, err := s.repo.GetObjective(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.calculateProgress(todayStart, stats, objective), nil
}

func (s *service) GetHistoricalProgress(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]DailyProgress, error) {
	objective, err := s.repo.GetObjective(ctx, userID)
	if err != nil {
		return nil, err
	}

	var results []DailyProgress

	// Iterate through each day in the range
	current := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)
	end := time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, 999999999, time.UTC)

	for current.Before(end) || current.Equal(end) {
		nextDay := current.AddDate(0, 0, 1)

		stats, err := s.computeStats(ctx, userID, current, nextDay)
		if err != nil {
			return nil, err
		}

		progress := s.calculateProgress(current, stats, objective)
		results = append(results, *progress)

		current = nextDay
	}

	return results, nil
}

// computeStats calculates application counts for a date range grouped by seniority.
func (s *service) computeStats(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) (*DailyStats, error) {
	var stats struct {
		Total  int
		Junior int
		Pleno  int
		Senior int
	}

	// Query job applications created between startDate and endDate
	query := s.db.WithContext(ctx).
		Table("job_applications ja").
		Select(`
			COUNT(ja.id) as total,
			COUNT(CASE WHEN jl.seniority = 'junior' THEN 1 END) as junior,
			COUNT(CASE WHEN jl.seniority = 'pleno' THEN 1 END) as pleno,
			COUNT(CASE WHEN jl.seniority = 'senior' THEN 1 END) as senior
		`).
		Joins("LEFT JOIN job_levels jl ON ja.job_level_id = jl.id").
		Where("ja.user_id = ? AND ja.created_at >= ? AND ja.created_at < ?", userID, startDate, endDate).
		Scan(&stats)

	   if query.Error != nil {
		   return nil, NewDomainError(ErrCodeRepositoryFailure, errors.New(ErrUnableToFetch))
	   }

	return &DailyStats{
		Date:        startDate,
		TotalCount:  stats.Total,
		JuniorCount: stats.Junior,
		PlenoCount:  stats.Pleno,
		SeniorCount: stats.Senior,
	}, nil
}

// calculateProgress calculates progress percentages.
func (s *service) calculateProgress(date time.Time, stats *DailyStats, objective *DailyObjective) *DailyProgress {
	progressTotal := 0.0
	if objective.TotalTarget > 0 {
		progressTotal = (float64(stats.TotalCount) / float64(objective.TotalTarget)) * 100.0
	}

	progressJunior := 0.0
	if objective.JuniorTarget > 0 {
		progressJunior = (float64(stats.JuniorCount) / float64(objective.JuniorTarget)) * 100.0
	}

	progressPleno := 0.0
	if objective.PlenoTarget > 0 {
		progressPleno = (float64(stats.PlenoCount) / float64(objective.PlenoTarget)) * 100.0
	}

	progressSenior := 0.0
	if objective.SeniorTarget > 0 {
		progressSenior = (float64(stats.SeniorCount) / float64(objective.SeniorTarget)) * 100.0
	}

	// Cap at 100% for display purposes
	if progressTotal > 100 {
		progressTotal = 100
	}
	if progressJunior > 100 {
		progressJunior = 100
	}
	if progressPleno > 100 {
		progressPleno = 100
	}
	if progressSenior > 100 {
		progressSenior = 100
	}

	return &DailyProgress{
		Date:            date,
		Stats:           *stats,
		Targets:         *objective,
		ProgressTotal:   progressTotal,
		ProgressJunior:  progressJunior,
		ProgressPleno:   progressPleno,
		ProgressSenior:  progressSenior,
	}
}

// validateRequest validates the create/update request.
func validateRequest(req CreateObjectiveRequest) error {
	// Check for negative values
	   if req.TotalTarget < 0 || req.JuniorTarget < 0 || req.PlenoTarget < 0 || req.SeniorTarget < 0 {
		   return NewDomainError(ErrCodeValidation, errors.New(ErrNegativeTargets))
	   }

	// Check that sum equals total
	sum := req.JuniorTarget + req.PlenoTarget + req.SeniorTarget
	   if sum != req.TotalTarget {
		   return NewDomainError(ErrCodeValidation, errors.New(ErrTargetSumMismatch))
	   }

	return nil
}
