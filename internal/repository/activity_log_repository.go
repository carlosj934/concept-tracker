package repository

import (
	"context"
	"time"

	"concept-tracker/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ActivityLogRepository interface {
	List(ctx context.Context, userID string, conceptID string, cursor *domain.Cursor, limit int) ([]domain.ActivityLog, error)
	Create(ctx context.Context, userID string, conceptID string, activity domain.ActivityLog) (domain.ActivityLog, error)
	Update(ctx context.Context, userID string, id string, actType *string, duration *int64, notes *string, loggedAt *time.Time) (domain.ActivityLog, error)
	Delete(ctx context.Context, userID string, id string) error
}

type postgresActivityLogRepository struct {
	pool *pgxpool.Pool
}

func NewActivityLog(p *pgxpool.Pool) ActivityLogRepository {
	return &postgresActivityLogRepository{pool: p}
}

func (r *postgresActivityLogRepository) List(ctx context.Context, userID string, conceptID string, cursor *domain.Cursor, limit int) ([]domain.ActivityLog, error) {
	var rows pgx.Rows
	var err error

	if cursor == nil {
		rows, err = r.pool.Query(ctx, `
		SELECT id, concept_id, user_id, activity_type, duration_minutes, notes, logged_at, created_at
		FROM activity_logs
		WHERE user_id = $1 AND concept_id = $2
		ORDER BY logged_at DESC, id DESC
		LIMIT $3
		`, userID, conceptID, limit+1)
	} else {
		rows, err = r.pool.Query(ctx, `
		SELECT id, concept_id, user_id, activity_type, duration_minutes, notes, logged_at, created_at
		FROM activity_logs
		WHERE user_id = $1 AND concept_id = $2 AND (logged_at, id) < ($3, $4)
		ORDER BY logged_at DESC, id DESC
		LIMIT $5
		`, userID, conceptID, cursor.LoggedAt, cursor.ID, limit+1)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activityLogs []domain.ActivityLog
	for rows.Next() {
		var a domain.ActivityLog
		var id, conceptID, userID, activity string
		var duration *int64
		var notes *string
		var loggedAt, createdAt time.Time

		err := rows.Scan(&id, &conceptID, &userID, &activity, &duration, &notes, &loggedAt, &createdAt)

		if err != nil {
			return nil, err
		}
		a.ID = id
		a.ConceptID = conceptID
		a.UserID = userID
		a.ActivityType = activity
		a.DurationMins = duration
		a.Notes = notes
		a.LoggedAt = loggedAt
		a.CreatedAt = createdAt
		activityLogs = append(activityLogs, a)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return activityLogs, nil
}

func (r *postgresActivityLogRepository) Create(ctx context.Context, userID string, conceptID string, activity domain.ActivityLog) (domain.ActivityLog, error) {
	var id string
	var createdAt time.Time

	err := r.pool.QueryRow(ctx, `
	INSERT INTO activity_logs (concept_id, user_id, activity_type, duration_minutes, notes, logged_at)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id, created_at
	`, conceptID, userID, activity.ActivityType, activity.DurationMins, activity.Notes, activity.LoggedAt).Scan(&id, &createdAt)
	if err != nil {
		return domain.ActivityLog{}, err
	}

	return domain.ActivityLog{
		ID:           id,
		ConceptID:    conceptID,
		UserID:       userID,
		ActivityType: activity.ActivityType,
		DurationMins: activity.DurationMins,
		Notes:        activity.Notes,
		LoggedAt:     activity.LoggedAt,
		CreatedAt:    activity.CreatedAt,
	}, nil
}

func (r *postgresActivityLogRepository) Update(ctx context.Context, userID string, id string, actType *string, duration *int64, notes *string, loggedAt *time.Time) (domain.ActivityLog, error) {
	// using individual letters to not overshadow input params
	var u, conceptID, i, a string
	var n *string
	var d *int64
	var l, createdAt time.Time

	err := r.pool.QueryRow(ctx, `
	UPDATE activity_logs
	SET activity_type = COALESCE($1, activity_type), duration_minutes = COALESCE($2, duration_minutes), notes = COALESCE($3, notes), logged_at = COALESCE($4, logged_at)
	WHERE id = $5 AND user_id = $6
	RETURNING concept_id, user_id, id, activity_type, duration_minutes, notes, logged_at, created_at
	`, actType, duration, notes, loggedAt, id, userID).Scan(&conceptID, &u, &i, &a, &d, &n, &l, &createdAt)
	if err != nil {
		return domain.ActivityLog{}, err
	}

	return domain.ActivityLog{
		ID:           id,
		ConceptID:    conceptID,
		UserID:       userID,
		ActivityType: a,
		DurationMins: d,
		Notes:        n,
		LoggedAt:     l,
		CreatedAt:    createdAt,
	}, nil
}

func (r *postgresActivityLogRepository) Delete(ctx context.Context, userID string, id string) error {
	_, err := r.pool.Exec(ctx, `
	DELETE FROM activity_logs
	WHERE id = $1 AND user_id = $2
	`, id, userID)
	if err != nil {
		return err
	}

	return nil
}
