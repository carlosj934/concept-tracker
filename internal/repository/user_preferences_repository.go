package repository

import (
	"context"
	"time"

	"concept-tracker/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserPreferencesRepository interface {
	GetUserPreferences(ctx context.Context, userID string) (domain.UserPreferences, error)
	Update(ctx context.Context, userID string, timezone string) (domain.UserPreferences, error)
}

type postgresUserPreferencesRepository struct {
	pool *pgxpool.Pool
}

func NewUserPreference(p *pgxpool.Pool) UserPreferencesRepository {
	return &postgresUserPreferencesRepository{pool: p}
}

func (r *postgresUserPreferencesRepository) GetUserPreferences(ctx context.Context, userID string) (domain.UserPreferences, error) {
	var u, timezone string
	var updatedAt time.Time

	err := r.pool.QueryRow(ctx, `
	SELECT user_id, timezone, updated_at
	FROM user_preferences
	WHERE user_id = $1
	`, userID).Scan(&u, &timezone, &updatedAt)
	if err != nil {
		return domain.UserPreferences{}, err
	}

	return domain.UserPreferences{
		UserID:    u,
		Timezone:  timezone,
		UpdatedAt: updatedAt,
	}, nil
}

func (r *postgresUserPreferencesRepository) Update(ctx context.Context, userID string, timezone string) (domain.UserPreferences, error) {
	var u, t string
	var updatedAt time.Time

	err := r.pool.QueryRow(ctx, `
	INSERT INTO user_preferences (user_id, timezone)
	VALUES ($1, $2)
	ON CONFLICT (user_id)
	DO UPDATE SET timezone = $2, updated_at = now() 
	RETURNING user_id, timezone, updated_at
	`, userID, timezone).Scan(&u, &t, &updatedAt)
	if err != nil {
		return domain.UserPreferences{}, err
	}

	return domain.UserPreferences{
		UserID:    u,
		Timezone:  t,
		UpdatedAt: updatedAt,
	}, nil
}
