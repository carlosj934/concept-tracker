package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"concept-tracker/internal/domain"
	"concept-tracker/internal/repository"
)

type ActivityLogService interface {
	List(ctx context.Context, userID string, conceptID string, cursor string, limit int) (domain.ActivityLogPage, error)
	Create(ctx context.Context, userID string, conceptID string, activity domain.ActivityLog) (domain.ActivityLog, error)
	Update(ctx context.Context, userID string, id string, actType *string, duration *int64, notes *string, loggedAt *time.Time) (domain.ActivityLog, error)
	Delete(ctx context.Context, userID string, id string) error
}

type activityLogService struct {
	repo repository.ActivityLogRepository
}

func NewActivityLogService(repo repository.ActivityLogRepository) ActivityLogService {
	return &activityLogService{
		repo: repo,
	}
}

func (a activityLogService) List(ctx context.Context, userID string, conceptID string, cursor string, limit int) (domain.ActivityLogPage, error) {
	var c *domain.Cursor

	if cursor != "" {
		data, err := base64.StdEncoding.DecodeString(cursor)
		if err != nil {
			return domain.ActivityLogPage{}, err
		}

		s := strings.Split(string(data), "_")
		t, err := time.Parse(time.RFC3339, s[0])
		if err != nil {
			return domain.ActivityLogPage{}, err
		}

		c = &domain.Cursor{LoggedAt: t, ID: s[1]}
	}

	l, err := a.repo.List(ctx, userID, conceptID, c, limit)
	if err != nil {
		return domain.ActivityLogPage{}, err
	}

	hasMore := len(l) > limit
	var NextCursor *string
	if hasMore {
		l = l[:limit]

		cursorStr := fmt.Sprintf("%s_%s", l[len(l)-1].LoggedAt.Format(time.RFC3339), l[len(l)-1].ID)
		encoded := base64.StdEncoding.EncodeToString([]byte(cursorStr))
		NextCursor = &encoded
	}

	return domain.ActivityLogPage{
		Data:       l,
		NextCursor: NextCursor,
		HasMore:    hasMore,
	}, nil
}

func (a activityLogService) Create(ctx context.Context, userID string, conceptID string, activity domain.ActivityLog) (domain.ActivityLog, error) {
	c, err := a.repo.Create(ctx, userID, conceptID, activity)
	if err != nil {
		return domain.ActivityLog{}, err
	}

	return c, nil
}

func (a activityLogService) Update(ctx context.Context, userID string, id string, actType *string, duration *int64, notes *string, loggedAt *time.Time) (domain.ActivityLog, error) {
	u, err := a.repo.Update(ctx, userID, id, actType, duration, notes, loggedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.ActivityLog{}, domain.ErrNotFound
		}
		return domain.ActivityLog{}, err
	}

	return u, nil
}

func (a activityLogService) Delete(ctx context.Context, userID string, id string) error {
	if err := a.repo.Delete(ctx, userID, id); err != nil {
		if err == pgx.ErrNoRows {
			return domain.ErrNotFound
		}
		return err
	}

	return nil
}
