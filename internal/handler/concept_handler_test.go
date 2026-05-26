package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"concept-tracker/internal/domain"
	"concept-tracker/internal/mocks"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	testUserID    = "user_abc123"
	testConceptID = "concept-uuid-1"
)

func TestConceptHandler_GetByID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		conceptID  string
		userID     string
		mockSetup  func(svc *mocks.MockConceptService)
		wantStatus int
	}{
		{
			name:      "returns concept with children",
			conceptID: testConceptID,
			userID:    testUserID,
			mockSetup: func(svc *mocks.MockConceptService) {
				svc.EXPECT().GetByID(mock.Anything, testUserID, testConceptID).
					Return(domain.ConceptWithChildren{
						Concept:  domain.Concept{ID: testConceptID, Name: "GoLang"},
						Children: []domain.Concept{{ID: "child-1", Name: "Concurrency"}},
					}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:      "returns 404 when concept not found",
			conceptID: "nonexistent",
			userID:    testUserID,
			mockSetup: func(svc *mocks.MockConceptService) {
				svc.EXPECT().GetByID(mock.Anything, testUserID, "nonexistent").
					Return(domain.ConceptWithChildren{}, domain.ErrNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:      "returns 500 on unexpected error",
			conceptID: testConceptID,
			userID:    testUserID,
			mockSetup: func(svc *mocks.MockConceptService) {
				svc.EXPECT().GetByID(mock.Anything, testUserID, testConceptID).
					Return(domain.ConceptWithChildren{}, errors.New("db connection lost"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := mocks.NewMockConceptService(t)
			tt.mockSetup(svc)

			w := httptest.NewRecorder()
			_, router := gin.CreateTestContext(w)

			handler := NewConceptHandler(svc)
			router.GET("/concepts/:id", func(c *gin.Context) {
				c.Set("userID", tt.userID)
				handler.GetByID(c)
			})

			req := httptest.NewRequest(http.MethodGet, "/concepts/"+tt.conceptID, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestConceptHandler_ListRoots(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		userID     string
		mockSetup  func(svc *mocks.MockConceptService)
		wantStatus int
	}{
		{
			name:   "returns list of root concepts",
			userID: testUserID,
			mockSetup: func(svc *mocks.MockConceptService) {
				svc.EXPECT().ListRoots(mock.Anything, testUserID).
					Return([]domain.Concept{
						{ID: "root-1", Name: "GoLang"},
						{ID: "root-2", Name: "Databases"},
					}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "returns empty list when user has no concepts",
			userID: testUserID,
			mockSetup: func(svc *mocks.MockConceptService) {
				svc.EXPECT().ListRoots(mock.Anything, testUserID).
					Return([]domain.Concept{}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "returns 500 on service error",
			userID: testUserID,
			mockSetup: func(svc *mocks.MockConceptService) {
				svc.EXPECT().ListRoots(mock.Anything, testUserID).
					Return(nil, errors.New("unexpected error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := mocks.NewMockConceptService(t)
			tt.mockSetup(svc)

			w := httptest.NewRecorder()
			_, router := gin.CreateTestContext(w)

			handler := NewConceptHandler(svc)
			router.GET("/concepts", func(c *gin.Context) {
				c.Set("userID", tt.userID)
				handler.ListRoots(c)
			})

			req := httptest.NewRequest(http.MethodGet, "/concepts", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestConceptHandler_GetSubtree(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		conceptID  string
		userID     string
		mockSetup  func(svc *mocks.MockConceptService)
		wantStatus int
	}{
		{
			name:      "returns full subtree",
			conceptID: testConceptID,
			userID:    testUserID,
			mockSetup: func(svc *mocks.MockConceptService) {
				svc.EXPECT().GetSubtree(mock.Anything, testUserID, testConceptID).
					Return([]domain.Concept{
						{ID: testConceptID, Name: "GoLang"},
						{ID: "child-1", Name: "Concurrency"},
						{ID: "grandchild-1", Name: "Goroutines"},
					}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:      "returns 500 on service error",
			conceptID: testConceptID,
			userID:    testUserID,
			mockSetup: func(svc *mocks.MockConceptService) {
				svc.EXPECT().GetSubtree(mock.Anything, testUserID, testConceptID).
					Return(nil, errors.New("unexpected error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := mocks.NewMockConceptService(t)
			tt.mockSetup(svc)

			w := httptest.NewRecorder()
			_, router := gin.CreateTestContext(w)

			handler := NewConceptHandler(svc)
			router.GET("/concepts/:id/tree", func(c *gin.Context) {
				c.Set("userID", tt.userID)
				handler.GetSubtree(c)
			})

			req := httptest.NewRequest(http.MethodGet, "/concepts/"+tt.conceptID+"/tree", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestConceptHandler_Create(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		userID     string
		body       any
		mockSetup  func(svc *mocks.MockConceptService)
		wantStatus int
	}{
		{
			name:   "creates a root concept",
			userID: testUserID,
			body: map[string]any{
				"name":        "GoLang",
				"description": "Go programming language",
				"parent_id":   nil,
			},
			mockSetup: func(svc *mocks.MockConceptService) {
				svc.EXPECT().Create(mock.Anything, testUserID, mock.MatchedBy(func(c domain.Concept) bool {
					return c.Name == "GoLang" && c.ParentID == nil
				})).Return(domain.Concept{ID: testConceptID, Name: "GoLang"}, nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:   "creates a child concept",
			userID: testUserID,
			body: map[string]any{
				"name":      "Concurrency",
				"parent_id": "parent-uuid",
			},
			mockSetup: func(svc *mocks.MockConceptService) {
				svc.EXPECT().Create(mock.Anything, testUserID, mock.MatchedBy(func(c domain.Concept) bool {
					return c.Name == "Concurrency" && c.ParentID != nil && *c.ParentID == "parent-uuid"
				})).Return(domain.Concept{ID: "child-1", Name: "Concurrency"}, nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "returns 400 on malformed JSON",
			userID:     testUserID,
			body:       "not-json",
			mockSetup:  func(svc *mocks.MockConceptService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "returns 500 on service error",
			userID: testUserID,
			body:   map[string]any{"name": "GoLang"},
			mockSetup: func(svc *mocks.MockConceptService) {
				svc.EXPECT().Create(mock.Anything, testUserID, mock.Anything).
					Return(domain.Concept{}, errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := mocks.NewMockConceptService(t)
			tt.mockSetup(svc)

			w := httptest.NewRecorder()
			_, router := gin.CreateTestContext(w)

			handler := NewConceptHandler(svc)
			router.POST("/concepts", func(c *gin.Context) {
				c.Set("userID", tt.userID)
				handler.Create(c)
			})

			var buf bytes.Buffer
			_ = json.NewEncoder(&buf).Encode(tt.body)

			req := httptest.NewRequest(http.MethodPost, "/concepts", &buf)
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestConceptHandler_Update(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		conceptID  string
		userID     string
		body       any
		mockSetup  func(svc *mocks.MockConceptService)
		wantStatus int
	}{
		{
			name:      "updates name and description",
			conceptID: testConceptID,
			userID:    testUserID,
			body: map[string]any{
				"name":        "GoLang Updated",
				"description": "Updated description",
			},
			mockSetup: func(svc *mocks.MockConceptService) {
				desc := "Updated description"
				svc.EXPECT().Update(mock.Anything, testUserID, testConceptID, "GoLang Updated", &desc).
					Return(domain.Concept{ID: testConceptID, Name: "GoLang Updated"}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:      "returns 404 when concept not found",
			conceptID: "nonexistent",
			userID:    testUserID,
			body:      map[string]any{"name": "Anything"},
			mockSetup: func(svc *mocks.MockConceptService) {
				svc.EXPECT().Update(mock.Anything, testUserID, "nonexistent", "Anything", (*string)(nil)).
					Return(domain.Concept{}, domain.ErrNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "returns 400 on malformed JSON",
			conceptID:  testConceptID,
			userID:     testUserID,
			body:       "not-json",
			mockSetup:  func(svc *mocks.MockConceptService) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := mocks.NewMockConceptService(t)
			tt.mockSetup(svc)

			w := httptest.NewRecorder()
			_, router := gin.CreateTestContext(w)

			handler := NewConceptHandler(svc)
			router.PATCH("/concepts/:id", func(c *gin.Context) {
				c.Set("userID", tt.userID)
				handler.Update(c)
			})

			var buf bytes.Buffer
			_ = json.NewEncoder(&buf).Encode(tt.body)

			req := httptest.NewRequest(http.MethodPatch, "/concepts/"+tt.conceptID, &buf)
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestConceptHandler_Move(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		conceptID  string
		userID     string
		body       any
		mockSetup  func(svc *mocks.MockConceptService)
		wantStatus int
	}{
		{
			name:      "moves concept to a new parent",
			conceptID: testConceptID,
			userID:    testUserID,
			body:      map[string]any{"newParentID": "new-parent-uuid"},
			mockSetup: func(svc *mocks.MockConceptService) {
				newParent := "new-parent-uuid"
				svc.EXPECT().Move(mock.Anything, testUserID, testConceptID, &newParent).
					Return(nil)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:      "moves concept to root (nil parent)",
			conceptID: testConceptID,
			userID:    testUserID,
			body:      map[string]any{"newParentID": nil},
			mockSetup: func(svc *mocks.MockConceptService) {
				svc.EXPECT().Move(mock.Anything, testUserID, testConceptID, (*string)(nil)).
					Return(nil)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:      "returns 404 when concept not found",
			conceptID: "nonexistent",
			userID:    testUserID,
			body:      map[string]any{"newParentID": "parent-uuid"},
			mockSetup: func(svc *mocks.MockConceptService) {
				svc.EXPECT().Move(mock.Anything, testUserID, "nonexistent", mock.Anything).
					Return(domain.ErrNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "returns 400 on malformed JSON",
			conceptID:  testConceptID,
			userID:     testUserID,
			body:       "not-json",
			mockSetup:  func(svc *mocks.MockConceptService) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := mocks.NewMockConceptService(t)
			tt.mockSetup(svc)

			w := httptest.NewRecorder()
			_, router := gin.CreateTestContext(w)

			handler := NewConceptHandler(svc)
			router.PATCH("/concepts/:id/move", func(c *gin.Context) {
				c.Set("userID", tt.userID)
				handler.Move(c)
			})

			var buf bytes.Buffer
			_ = json.NewEncoder(&buf).Encode(tt.body)

			req := httptest.NewRequest(http.MethodPatch, "/concepts/"+tt.conceptID+"/move", &buf)
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestConceptHandler_Delete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		conceptID  string
		userID     string
		mockSetup  func(svc *mocks.MockConceptService)
		wantStatus int
	}{
		{
			name:      "deletes concept and subtree",
			conceptID: testConceptID,
			userID:    testUserID,
			mockSetup: func(svc *mocks.MockConceptService) {
				svc.EXPECT().Delete(mock.Anything, testUserID, testConceptID).
					Return(nil)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:      "returns 404 when concept not found",
			conceptID: "nonexistent",
			userID:    testUserID,
			mockSetup: func(svc *mocks.MockConceptService) {
				svc.EXPECT().Delete(mock.Anything, testUserID, "nonexistent").
					Return(domain.ErrNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:      "returns 500 on unexpected error",
			conceptID: testConceptID,
			userID:    testUserID,
			mockSetup: func(svc *mocks.MockConceptService) {
				svc.EXPECT().Delete(mock.Anything, testUserID, testConceptID).
					Return(errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := mocks.NewMockConceptService(t)
			tt.mockSetup(svc)

			w := httptest.NewRecorder()
			_, router := gin.CreateTestContext(w)

			handler := NewConceptHandler(svc)
			router.DELETE("/concepts/:id", func(c *gin.Context) {
				c.Set("userID", tt.userID)
				handler.Delete(c)
			})

			req := httptest.NewRequest(http.MethodDelete, "/concepts/"+tt.conceptID, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
