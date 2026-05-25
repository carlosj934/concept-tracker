package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"concept-tracker/internal/domain"
	"concept-tracker/internal/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserPreferencesService_Update(t *testing.T) {
	t.Parallel()

	fixedTime := time.Date(2026, 1, 15, 9, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		timezone  string
		mockSetup func(repo *mocks.MockUserPreferencesRepository)
		want      domain.UserPreferences
		wantErr   bool
	}{
		{
			name:     "valid IANA timezone calls repo and returns preferences",
			timezone: "America/Los_Angeles",
			mockSetup: func(repo *mocks.MockUserPreferencesRepository) {
				repo.EXPECT().Update(mock.Anything, "user-123", "America/Los_Angeles").
					Return(domain.UserPreferences{
						UserID:    "user-123",
						Timezone:  "America/Los_Angeles",
						UpdatedAt: fixedTime,
					}, nil)
			},
			want: domain.UserPreferences{
				UserID:    "user-123",
				Timezone:  "America/Los_Angeles",
				UpdatedAt: fixedTime,
			},
			wantErr: false,
		},
		{
			name:      "invalid timezone returns error without calling repo",
			timezone:  "not/a/timezone",
			mockSetup: func(repo *mocks.MockUserPreferencesRepository) {},
			want:      domain.UserPreferences{},
			wantErr:   true,
		},
		{
			name:     "UTC is valid and calls repo",
			timezone: "UTC",
			mockSetup: func(repo *mocks.MockUserPreferencesRepository) {
				repo.EXPECT().Update(mock.Anything, "user-123", "UTC").
					Return(domain.UserPreferences{
						UserID:    "user-123",
						Timezone:  "UTC",
						UpdatedAt: fixedTime,
					}, nil)
			},
			want: domain.UserPreferences{
				UserID:    "user-123",
				Timezone:  "UTC",
				UpdatedAt: fixedTime,
			},
			wantErr: false,
		},
		{
			name:     "repo error is bubbled up",
			timezone: "Europe/London",
			mockSetup: func(repo *mocks.MockUserPreferencesRepository) {
				repo.EXPECT().Update(mock.Anything, "user-123", "Europe/London").
					Return(domain.UserPreferences{}, errors.New("db error"))
			},
			want:    domain.UserPreferences{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := mocks.NewMockUserPreferencesRepository(t)
			tt.mockSetup(repo)

			svc := NewUserPreferencesService(repo)
			got, err := svc.Update(context.Background(), "user-123", tt.timezone)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, domain.UserPreferences{}, got)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
