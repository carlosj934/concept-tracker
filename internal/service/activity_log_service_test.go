package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"concept-tracker/internal/domain"
	"concept-tracker/internal/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func validCursor(loggedAt time.Time, id string) string {
	raw := fmt.Sprintf("%s_%s", loggedAt.Format(time.RFC3339), id)
	return base64.StdEncoding.EncodeToString([]byte(raw))
}

func TestActivityLogService_List(t *testing.T) {
	t.Parallel()

	fixedTime := time.Date(2026, 1, 15, 9, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		cursor    string
		limit     int
		mockSetup func(repo *mocks.MockActivityLogRepository)
		wantPage  domain.ActivityLogPage
		wantErr   bool
	}{
		{
			name:   "no cursor passes nil to repo and returns page without next cursor",
			cursor: "",
			limit:  2,
			mockSetup: func(repo *mocks.MockActivityLogRepository) {
				repo.EXPECT().List(mock.Anything, mock.Anything, mock.Anything, (*domain.Cursor)(nil), 2).
					Return([]domain.ActivityLog{
						{ID: "log-1", LoggedAt: fixedTime},
						{ID: "log-2", LoggedAt: fixedTime.Add(-time.Hour)},
					}, nil)
			},
			wantPage: domain.ActivityLogPage{
				Data: []domain.ActivityLog{
					{ID: "log-1", LoggedAt: fixedTime},
					{ID: "log-2", LoggedAt: fixedTime.Add(-time.Hour)},
				},
				NextCursor: nil,
				HasMore:    false,
			},
			wantErr: false,
		},
		{
			name:   "valid cursor is decoded and passed to repo",
			cursor: validCursor(fixedTime, "log-99"),
			limit:  2,
			mockSetup: func(repo *mocks.MockActivityLogRepository) {
				repo.EXPECT().List(mock.Anything, mock.Anything, mock.Anything, &domain.Cursor{LoggedAt: fixedTime, ID: "log-99"}, 2).
					Return([]domain.ActivityLog{
						{ID: "log-100", LoggedAt: fixedTime.Add(-time.Hour)},
					}, nil)
			},
			wantPage: domain.ActivityLogPage{
				Data: []domain.ActivityLog{
					{ID: "log-100", LoggedAt: fixedTime.Add(-time.Hour)},
				},
				NextCursor: nil,
				HasMore:    false,
			},
			wantErr: false,
		},
		{
			name:      "malformed base64 cursor returns error without calling repo",
			cursor:    "not-valid-base64!!!",
			limit:     2,
			mockSetup: func(repo *mocks.MockActivityLogRepository) {},
			wantPage:  domain.ActivityLogPage{},
			wantErr:   true,
		},
		{
			name:      "valid base64 but invalid timestamp inside returns error without calling repo",
			cursor:    base64.StdEncoding.EncodeToString([]byte("not-a-timestamp_log-1")),
			limit:     2,
			mockSetup: func(repo *mocks.MockActivityLogRepository) {},
			wantPage:  domain.ActivityLogPage{},
			wantErr:   true,
		},
		{
			name:   "limit+1 results sets HasMore true, trims slice, and encodes NextCursor",
			cursor: "",
			limit:  2,
			mockSetup: func(repo *mocks.MockActivityLogRepository) {
				repo.EXPECT().List(mock.Anything, mock.Anything, mock.Anything, (*domain.Cursor)(nil), 2).
					Return([]domain.ActivityLog{
						{ID: "log-1", LoggedAt: fixedTime},
						{ID: "log-2", LoggedAt: fixedTime.Add(-time.Hour)},
						{ID: "log-3", LoggedAt: fixedTime.Add(-2 * time.Hour)},
					}, nil)
			},
			wantPage: domain.ActivityLogPage{
				Data: []domain.ActivityLog{
					{ID: "log-1", LoggedAt: fixedTime},
					{ID: "log-2", LoggedAt: fixedTime.Add(-time.Hour)},
				},
				NextCursor: func() *string {
					s := validCursor(fixedTime.Add(-time.Hour), "log-2")
					return &s
				}(),
				HasMore: true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := mocks.NewMockActivityLogRepository(t)
			tt.mockSetup(repo)

			svc := NewActivityLogService(repo)
			got, err := svc.List(context.Background(), "user-123", "concept-123", tt.cursor, tt.limit)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, domain.ActivityLogPage{}, got)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantPage, got)
		})
	}
}
