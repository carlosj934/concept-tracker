package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"concept-tracker/internal/domain"
	"concept-tracker/internal/mocks"
)

func TestUserPreferencesHandler_GetUserPreferences(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		mockSetup  func(svc *mocks.MockUserPreferencesService)
		wantStatus int
	}{
		{
			name: "returns 200 with preferences",
			mockSetup: func(svc *mocks.MockUserPreferencesService) {
				svc.EXPECT().GetUserPreferences(mock.Anything, testUserID).
					Return(domain.UserPreferences{UserID: testUserID, Timezone: "UTC"}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "returns 500 on service error",
			mockSetup: func(svc *mocks.MockUserPreferencesService) {
				svc.EXPECT().GetUserPreferences(mock.Anything, testUserID).
					Return(domain.UserPreferences{}, errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := mocks.NewMockUserPreferencesService(t)
			tt.mockSetup(svc)

			w := httptest.NewRecorder()
			_, router := gin.CreateTestContext(w)

			handler := NewUserPreferencesHandler(svc)
			router.GET("/me/preferences", func(c *gin.Context) {
				c.Set("userID", testUserID)
				handler.GetUserPreferences(c)
			})

			req := httptest.NewRequest(http.MethodGet, "/me/preferences", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestUserPreferencesHandler_Update(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		body       any
		mockSetup  func(svc *mocks.MockUserPreferencesService)
		wantStatus int
	}{
		{
			name: "returns 200 on success",
			body: map[string]any{"timezone": "America/Los_Angeles"},
			mockSetup: func(svc *mocks.MockUserPreferencesService) {
				svc.EXPECT().Update(mock.Anything, testUserID, "America/Los_Angeles").
					Return(domain.UserPreferences{UserID: testUserID, Timezone: "America/Los_Angeles"}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 400 when timezone field is missing",
			body:       map[string]any{},
			mockSetup:  func(svc *mocks.MockUserPreferencesService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 on malformed JSON",
			body:       "not-json",
			mockSetup:  func(svc *mocks.MockUserPreferencesService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "returns 500 on service error",
			body: map[string]any{"timezone": "America/Los_Angeles"},
			mockSetup: func(svc *mocks.MockUserPreferencesService) {
				svc.EXPECT().Update(mock.Anything, testUserID, "America/Los_Angeles").
					Return(domain.UserPreferences{}, errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := mocks.NewMockUserPreferencesService(t)
			tt.mockSetup(svc)

			w := httptest.NewRecorder()
			_, router := gin.CreateTestContext(w)

			handler := NewUserPreferencesHandler(svc)
			router.PATCH("/me/preferences", func(c *gin.Context) {
				c.Set("userID", testUserID)
				handler.Update(c)
			})

			var buf bytes.Buffer
			_ = json.NewEncoder(&buf).Encode(tt.body)

			req := httptest.NewRequest(http.MethodPatch, "/me/preferences", &buf)
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
