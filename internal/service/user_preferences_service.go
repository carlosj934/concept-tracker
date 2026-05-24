package service

import (
	"context"
	"fmt"
	"time"

	"concept-tracker/internal/domain"
	"concept-tracker/internal/repository"
)

type UserPreferencesService interface {
	GetUserPreferences(ctx context.Context, userID string) (domain.UserPreferences, error)
	Update(ctx context.Context, userID string, timezone string) (domain.UserPreferences, error)
}

type userPreferencesService struct {
	repo repository.UserPreferencesRepository
}

func NewUserPreferencesService(repo repository.UserPreferencesRepository) UserPreferencesService {
	return &userPreferencesService{
		repo: repo,
	}
}

func (s userPreferencesService) GetUserPreferences(ctx context.Context, userID string) (domain.UserPreferences, error) {
	g, err := s.repo.GetUserPreferences(ctx, userID)
	if err != nil {
		return domain.UserPreferences{}, err
	}

	return g, nil
}

func (s userPreferencesService) Update(ctx context.Context, userID string, timezone string) (domain.UserPreferences, error) {
	_, err := time.LoadLocation(timezone)
	if err != nil {
		return domain.UserPreferences{}, fmt.Errorf("timezone needs to be formatted in the standard IANA timezone convention")
	}

	u, err := s.repo.Update(ctx, userID, timezone)
	if err != nil {
		return domain.UserPreferences{}, err
	}

	return u, nil
}
