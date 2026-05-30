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

func TestActivityLogHandler_List(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		query      string
		mockSetup  func(svc *mocks.MockActivityLogService)
		wantStatus int
	}{
		{
			name:  "returns 200 with no query params defaults to limit 25",
			query: "",
			mockSetup: func(svc *mocks.MockActivityLogService) {
				svc.EXPECT().List(mock.Anything, testUserID, testConceptID, "", 25).
					Return(domain.ActivityLogPage{Data: []domain.ActivityLog{}}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "returns 200 with explicit limit",
			query: "?limit=10",
			mockSetup: func(svc *mocks.MockActivityLogService) {
				svc.EXPECT().List(mock.Anything, testUserID, testConceptID, "", 10).
					Return(domain.ActivityLogPage{Data: []domain.ActivityLog{}}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "clamps limit to 100 when over",
			query: "?limit=200",
			mockSetup: func(svc *mocks.MockActivityLogService) {
				svc.EXPECT().List(mock.Anything, testUserID, testConceptID, "", 100).
					Return(domain.ActivityLogPage{Data: []domain.ActivityLog{}}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "passes cursor query param to service",
			query: "?cursor=abc123",
			mockSetup: func(svc *mocks.MockActivityLogService) {
				svc.EXPECT().List(mock.Anything, testUserID, testConceptID, "abc123", 25).
					Return(domain.ActivityLogPage{Data: []domain.ActivityLog{}}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "returns 500 on service error",
			query: "",
			mockSetup: func(svc *mocks.MockActivityLogService) {
				svc.EXPECT().List(mock.Anything, testUserID, testConceptID, "", 25).
					Return(domain.ActivityLogPage{}, errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := mocks.NewMockActivityLogService(t)
			tt.mockSetup(svc)

			w := httptest.NewRecorder()
			_, router := gin.CreateTestContext(w)

			handler := NewActivityLogHandler(svc)
			router.GET("/concepts/:id/logs", func(c *gin.Context) {
				c.Set("userID", testUserID)
				handler.List(c)
			})

			req := httptest.NewRequest(http.MethodGet, "/concepts/"+testConceptID+"/logs"+tt.query, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestActivityLogHandler_Create(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		body       any
		mockSetup  func(svc *mocks.MockActivityLogService)
		wantStatus int
	}{
		{
			name: "returns 201 on success",
			body: map[string]any{
				"activity_type":    "flashcards",
				"duration_minutes": 20,
				"logged_at":        "2026-01-15T09:00:00Z",
			},
			mockSetup: func(svc *mocks.MockActivityLogService) {
				svc.EXPECT().Create(mock.Anything, testUserID, testConceptID, mock.Anything).
					Return(domain.ActivityLog{ID: "log-1", ActivityType: "flashcards"}, nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "returns 400 on malformed JSON",
			body:       "not-json",
			mockSetup:  func(svc *mocks.MockActivityLogService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "returns 500 on service error",
			body: map[string]any{
				"activity_type": "reading",
				"logged_at":     "2026-01-15T09:00:00Z",
			},
			mockSetup: func(svc *mocks.MockActivityLogService) {
				svc.EXPECT().Create(mock.Anything, testUserID, testConceptID, mock.Anything).
					Return(domain.ActivityLog{}, errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := mocks.NewMockActivityLogService(t)
			tt.mockSetup(svc)

			w := httptest.NewRecorder()
			_, router := gin.CreateTestContext(w)

			handler := NewActivityLogHandler(svc)
			router.POST("/concepts/:id/logs", func(c *gin.Context) {
				c.Set("userID", testUserID)
				handler.Create(c)
			})

			var buf bytes.Buffer
			_ = json.NewEncoder(&buf).Encode(tt.body)

			req := httptest.NewRequest(http.MethodPost, "/concepts/"+testConceptID+"/logs", &buf)
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestActivityLogHandler_Update(t *testing.T) {
	t.Parallel()

	const testLogID = "log-uuid-1"

	tests := []struct {
		name       string
		body       any
		mockSetup  func(svc *mocks.MockActivityLogService)
		wantStatus int
	}{
		{
			name: "returns 200 on success",
			body: map[string]any{
				"activity_type": "practice",
			},
			mockSetup: func(svc *mocks.MockActivityLogService) {
				svc.EXPECT().Update(mock.Anything, testUserID, testLogID, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(domain.ActivityLog{ID: testLogID, ActivityType: "practice"}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 400 on malformed JSON",
			body:       "not-json",
			mockSetup:  func(svc *mocks.MockActivityLogService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "returns 404 when service returns ErrNotFound",
			body: map[string]any{"activity_type": "reading"},
			mockSetup: func(svc *mocks.MockActivityLogService) {
				svc.EXPECT().Update(mock.Anything, testUserID, testLogID, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(domain.ActivityLog{}, domain.ErrNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "returns 500 on unexpected service error",
			body: map[string]any{"activity_type": "reading"},
			mockSetup: func(svc *mocks.MockActivityLogService) {
				svc.EXPECT().Update(mock.Anything, testUserID, testLogID, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(domain.ActivityLog{}, errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := mocks.NewMockActivityLogService(t)
			tt.mockSetup(svc)

			w := httptest.NewRecorder()
			_, router := gin.CreateTestContext(w)

			handler := NewActivityLogHandler(svc)
			router.PATCH("/concepts/:id/logs/:lid", func(c *gin.Context) {
				c.Set("userID", testUserID)
				handler.Update(c)
			})

			var buf bytes.Buffer
			_ = json.NewEncoder(&buf).Encode(tt.body)

			req := httptest.NewRequest(http.MethodPatch, "/concepts/"+testConceptID+"/logs/"+testLogID, &buf)
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestActivityLogHandler_Delete(t *testing.T) {
	t.Parallel()

	const testLogID = "log-uuid-1"

	tests := []struct {
		name       string
		mockSetup  func(svc *mocks.MockActivityLogService)
		wantStatus int
	}{
		{
			name: "returns 204 on success",
			mockSetup: func(svc *mocks.MockActivityLogService) {
				svc.EXPECT().Delete(mock.Anything, testUserID, testLogID).
					Return(nil)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name: "returns 404 when service returns ErrNotFound",
			mockSetup: func(svc *mocks.MockActivityLogService) {
				svc.EXPECT().Delete(mock.Anything, testUserID, testLogID).
					Return(domain.ErrNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "returns 500 on unexpected service error",
			mockSetup: func(svc *mocks.MockActivityLogService) {
				svc.EXPECT().Delete(mock.Anything, testUserID, testLogID).
					Return(errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := mocks.NewMockActivityLogService(t)
			tt.mockSetup(svc)

			w := httptest.NewRecorder()
			_, router := gin.CreateTestContext(w)

			handler := NewActivityLogHandler(svc)
			router.DELETE("/concepts/:id/logs/:lid", func(c *gin.Context) {
				c.Set("userID", testUserID)
				handler.Delete(c)
			})

			req := httptest.NewRequest(http.MethodDelete, "/concepts/"+testConceptID+"/logs/"+testLogID, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
