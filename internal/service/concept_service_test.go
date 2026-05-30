package service

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"concept-tracker/internal/domain"
	"concept-tracker/internal/mocks"
)

func TestGetByID(t *testing.T) {
	t.Parallel()

	errSomething := errors.New("some error")

	tests := []struct {
		name      string
		mockSetup func(repo *mocks.MockConceptRepository)
		want      domain.ConceptWithChildren
		wantErr   error
	}{
		{
			name: "return domain.ErrNotFound if pgx.ErrNoRows",
			mockSetup: func(repo *mocks.MockConceptRepository) {
				repo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(domain.Concept{}, pgx.ErrNoRows)
			},
			want:    domain.ConceptWithChildren{},
			wantErr: domain.ErrNotFound,
		},
		{
			name: "if repo.GetByID returns some other error, service bubbles it up",
			mockSetup: func(repo *mocks.MockConceptRepository) {
				repo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(domain.Concept{}, errSomething)
			},
			want:    domain.ConceptWithChildren{},
			wantErr: errSomething,
		},
		{
			name: "if repo.GetByID succeeds but repo.GetChildren fails, service returns the error",
			mockSetup: func(repo *mocks.MockConceptRepository) {
				repo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(domain.Concept{}, nil)
				repo.On("GetChildren", mock.Anything, mock.Anything, mock.Anything).Return(nil, errSomething)
			},
			want:    domain.ConceptWithChildren{},
			wantErr: errSomething,
		},
		{
			name: "both repo.GetByID and repo.GetChildren succeed, return ConceptWithChildren",
			mockSetup: func(repo *mocks.MockConceptRepository) {
				repo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(domain.Concept{}, nil)
				repo.On("GetChildren", mock.Anything, mock.Anything, mock.Anything).Return([]domain.Concept{}, nil)
			},
			want: domain.ConceptWithChildren{
				Concept:  domain.Concept{},
				Children: []domain.Concept{},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := &mocks.MockConceptRepository{}

			tt.mockSetup(repo)

			svc := NewConceptService(repo)
			got, err := svc.GetByID(context.Background(), "user-123", "concept-123")

			assert.Equal(t, tt.want, got)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestDelete(t *testing.T) {
	t.Parallel()

	errSomething := errors.New("some error")

	tests := []struct {
		name      string
		mockSetup func(repo *mocks.MockConceptRepository)
		wantErr   error
	}{
		{
			name: "successfuly deletes",
			mockSetup: func(repo *mocks.MockConceptRepository) {
				repo.On("Delete", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "doesn't delete properly, returns error",
			mockSetup: func(repo *mocks.MockConceptRepository) {
				repo.On("Delete", mock.Anything, mock.Anything, mock.Anything).Return(errSomething)
			},
			wantErr: errSomething,
		},
		{
			name: "can't find concept to be deleted, return that it isn't found",
			mockSetup: func(repo *mocks.MockConceptRepository) {
				repo.On("Delete", mock.Anything, mock.Anything, mock.Anything).Return(pgx.ErrNoRows)
			},
			wantErr: domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := &mocks.MockConceptRepository{}

			tt.mockSetup(repo)

			svc := NewConceptService(repo)
			err := svc.Delete(context.Background(), "user-123", "concept-123")

			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestMove(t *testing.T) {
	t.Parallel()

	errSomething := errors.New("some error")

	tests := []struct {
		name      string
		mockSetup func(repo *mocks.MockConceptRepository)
		wantErr   error
	}{
		{
			name: "successfuly moves",
			mockSetup: func(repo *mocks.MockConceptRepository) {
				repo.On("Move", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "doesn't move properly, returns error",
			mockSetup: func(repo *mocks.MockConceptRepository) {
				repo.On("Move", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errSomething)
			},
			wantErr: errSomething,
		},
		{
			name: "can't find concept to be moved, return that it isn't found",
			mockSetup: func(repo *mocks.MockConceptRepository) {
				repo.On("Move", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(pgx.ErrNoRows)
			},
			wantErr: domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := &mocks.MockConceptRepository{}

			tt.mockSetup(repo)

			svc := NewConceptService(repo)
			err := svc.Move(context.Background(), "user-123", "concept-123", new("parent-123"))

			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
