package service

import (
	"context"

	"github.com/jackc/pgx/v5"

	"concept-tracker/internal/domain"
	"concept-tracker/internal/repository"
)

type ResourceService interface {
	ListConceptResources(ctx context.Context, userID string, conceptID string) ([]domain.ConceptResource, error)
	Create(ctx context.Context, userID string, conceptID string, resource domain.ConceptResource) (domain.ConceptResource, error)
	Update(ctx context.Context, userID string, id string, url *string, title *string) (domain.ConceptResource, error)
	Delete(ctx context.Context, userID string, id string) error
}

type resourceService struct {
	repo repository.ResourceRepository
}

func NewResourceService(repo repository.ResourceRepository) ResourceService {
	return &resourceService{
		repo: repo,
	}
}

func (r resourceService) ListConceptResources(ctx context.Context, userID string, conceptID string) ([]domain.ConceptResource, error) {
	l, err := r.repo.ListConceptResources(ctx, userID, conceptID)
	if err != nil {
		return nil, err
	}

	return l, nil
}

func (r resourceService) Create(ctx context.Context, userID string, conceptID string, resource domain.ConceptResource) (domain.ConceptResource, error) {
	c, err := r.repo.Create(ctx, userID, conceptID, resource)
	if err != nil {
		return domain.ConceptResource{}, err
	}

	return c, nil
}

func (r resourceService) Update(ctx context.Context, userID string, id string, url *string, title *string) (domain.ConceptResource, error) {
	u, err := r.repo.Update(ctx, userID, id, url, title)
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.ConceptResource{}, domain.ErrNotFound
		}
		return domain.ConceptResource{}, err
	}

	return u, nil
}

func (r resourceService) Delete(ctx context.Context, userID string, id string) error {
	if err := r.repo.Delete(ctx, userID, id); err != nil {
		if err == pgx.ErrNoRows {
			return domain.ErrNotFound
		}
		return err
	}

	return nil
}
