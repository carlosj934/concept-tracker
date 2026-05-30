package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"concept-tracker/internal/domain"
)

type ResourceRepository interface {
	ListConceptResources(ctx context.Context, userID string, conceptID string) ([]domain.ConceptResource, error)
	Create(ctx context.Context, userID string, conceptID string, resource domain.ConceptResource) (domain.ConceptResource, error)
	Update(ctx context.Context, userID string, id string, url *string, title *string) (domain.ConceptResource, error)
	Delete(ctx context.Context, userID string, id string) error
}

type postgresResourceRepository struct {
	pool *pgxpool.Pool
}

func NewResource(p *pgxpool.Pool) ResourceRepository {
	return &postgresResourceRepository{pool: p}
}

func (r *postgresResourceRepository) ListConceptResources(ctx context.Context, userID string, conceptID string) ([]domain.ConceptResource, error) {
	rows, err := r.pool.Query(ctx, `
	SELECT id, concept_id, user_id, provider, external_id, url, title, meta, created_at
	FROM concept_resources
	WHERE user_id = $1 AND concept_id = $2
	`, userID, conceptID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conceptResources []domain.ConceptResource
	for rows.Next() {
		var c domain.ConceptResource
		var id, conceptID, userID, provider, externalID, URL, title string
		var meta []byte
		var createdAt time.Time

		err := rows.Scan(&id, &conceptID, &userID, &provider, &externalID, &URL, &title, &meta, &createdAt)
		if err != nil {
			return nil, err
		}
		c.ID = id
		c.UserID = userID
		c.ConceptID = conceptID
		c.Provider = provider
		c.ExternalID = externalID
		c.URL = URL
		c.Title = title
		c.Meta = meta
		c.CreatedAt = createdAt
		conceptResources = append(conceptResources, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return conceptResources, nil
}

func (r *postgresResourceRepository) Create(ctx context.Context, userID string, conceptID string, resource domain.ConceptResource) (domain.ConceptResource, error) {
	var id string
	var createdAt time.Time

	err := r.pool.QueryRow(ctx, `
	INSERT INTO concept_resources (concept_id, user_id, provider, external_id, url, title, meta)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	RETURNING id, created_at
	`, conceptID, userID, resource.Provider, resource.ExternalID, resource.URL, resource.Title, resource.Meta).Scan(&id, &createdAt)
	if err != nil {
		return domain.ConceptResource{}, err
	}

	return domain.ConceptResource{
		ID:         id,
		ConceptID:  conceptID,
		UserID:     userID,
		Provider:   resource.Provider,
		ExternalID: resource.ExternalID,
		URL:        resource.URL,
		Title:      resource.Title,
		Meta:       resource.Meta,
		CreatedAt:  createdAt,
	}, nil
}

func (r *postgresResourceRepository) Update(ctx context.Context, userID string, id string, url *string, title *string) (domain.ConceptResource, error) {
	// using individual letters to not overshadow input params
	var URL, t, u, i, conceptID, provider, externalID string
	var meta []byte
	var createdAt time.Time

	err := r.pool.QueryRow(ctx, `
	UPDATE concept_resources
	SET url = COALESCE($1, url), title = COALESCE($2, title)
	WHERE id = $3 AND user_id = $4
	RETURNING url, title, id, user_id, concept_id, provider, external_id, meta, created_at
	`, url, title, id, userID).Scan(&URL, &t, &i, &u, &conceptID, &provider, &externalID, &meta, &createdAt)
	if err != nil {
		return domain.ConceptResource{}, err
	}

	return domain.ConceptResource{
		ID:         i,
		ConceptID:  conceptID,
		UserID:     u,
		Provider:   provider,
		ExternalID: externalID,
		URL:        URL,
		Title:      t,
		Meta:       meta,
		CreatedAt:  createdAt,
	}, nil
}

func (r *postgresResourceRepository) Delete(ctx context.Context, userID string, id string) error {
	_, err := r.pool.Exec(ctx, `
	DELETE FROM concept_resources
	WHERE id = $1 AND user_id = $2
	`, id, userID)
	if err != nil {
		return err
	}

	return nil
}
