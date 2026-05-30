package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"concept-tracker/internal/domain"
	"concept-tracker/internal/repository"
)

type ConceptService interface {
	ListRoots(ctx context.Context, userID string) ([]domain.Concept, error)
	GetByID(ctx context.Context, userID string, id string) (domain.ConceptWithChildren, error)
	GetSubtree(ctx context.Context, userID string, id string) ([]domain.Concept, error)
	Create(ctx context.Context, userID string, concept domain.Concept) (domain.Concept, error)
	Update(ctx context.Context, userID string, id string, name string, description *string) (domain.Concept, error)
	Move(ctx context.Context, userID string, id string, newParentID *string) error
	Delete(ctx context.Context, userID string, id string) error
}

type conceptService struct {
	repo repository.ConceptRepository
}

func NewConceptService(repo repository.ConceptRepository) ConceptService {
	return &conceptService{
		repo: repo,
	}
}

func (c conceptService) GetByID(ctx context.Context, userID string, id string) (domain.ConceptWithChildren, error) {
	i, err := c.repo.GetByID(ctx, userID, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ConceptWithChildren{}, domain.ErrNotFound
		}
		return domain.ConceptWithChildren{}, err
	}

	g, err := c.repo.GetChildren(ctx, userID, id)
	if err != nil {
		return domain.ConceptWithChildren{}, err
	}

	return domain.ConceptWithChildren{
		Concept:  i,
		Children: g,
	}, nil
}

func (c conceptService) ListRoots(ctx context.Context, userID string) ([]domain.Concept, error) {
	l, err := c.repo.ListRoots(ctx, userID)
	if err != nil {
		return nil, err
	}

	return l, nil
}

func (c conceptService) GetSubtree(ctx context.Context, userID string, id string) ([]domain.Concept, error) {
	s, err := c.repo.GetSubtree(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (c conceptService) Create(ctx context.Context, userID string, concept domain.Concept) (domain.Concept, error) {
	create, err := c.repo.Create(ctx, userID, concept)
	if err != nil {
		return domain.Concept{}, err
	}
	return create, nil
}

func (c conceptService) Update(ctx context.Context, userID string, id string, name string, description *string) (domain.Concept, error) {
	u, err := c.repo.Update(ctx, userID, id, name, description)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Concept{}, domain.ErrNotFound
		}
		return domain.Concept{}, err
	}

	return u, nil
}

func (c conceptService) Move(ctx context.Context, userID string, id string, newParentID *string) error {
	if err := c.repo.Move(ctx, userID, id, newParentID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return err
	}

	return nil
}

func (c conceptService) Delete(ctx context.Context, userID string, id string) error {
	if err := c.repo.Delete(ctx, userID, id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return err
	}

	return nil
}
