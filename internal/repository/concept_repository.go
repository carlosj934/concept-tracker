package repository

import (
	"context"
	"time"

	"concept-tracker/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ConceptRepository interface {
	Create(ctx context.Context, userID string, concept domain.Concept) (domain.Concept, error)
	GetByID(ctx context.Context, userID string, id string) (domain.Concept, error)
	GetChildren(ctx context.Context, userID string, id string) ([]domain.Concept, error)
	GetSubtree(ctx context.Context, userID string, id string) ([]domain.Concept, error)
	Update(ctx context.Context, userID string, id string, name string, description *string) (domain.Concept, error) 
	Move(ctx context.Context, userID string, id string, newParentID *string) error
	Delete(ctx context.Context, userID string, id string) error
}

type postgresConceptRepository struct {
	pool *pgxpool.Pool
}

func New(p *pgxpool.Pool) ConceptRepository {
	return &postgresConceptRepository{pool: p}
}

func (r *postgresConceptRepository) Create(ctx context.Context, userID string, concept domain.Concept) (domain.Concept, error) {
	var id string
	var createdAt, updatedAt time.Time

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Concept{}, err	
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
	INSERT INTO concepts (user_id, parent_id, name, description)
	VALUES ($1, $2, $3, $4)
	RETURNING id, created_at, updated_at
	`, concept.UserID, concept.ParentID, concept.Name, concept.Description).Scan(&id, &createdAt, &updatedAt)
	if err != nil {
		return domain.Concept{}, err 
	}
	
	if concept.ParentID == nil {
		_, err = tx.Exec(ctx, `
		INSERT INTO concept_paths (ancestor_id, descendant_id, depth)
		VALUES ($1, $2, $3)
		`, id, id, 0)
		if err != nil {
			return domain.Concept{}, err
		}
	} else {
		// self row
		_, err = tx.Exec(ctx, `
		INSERT INTO concept_paths (ancestor_id, descendant_id, depth)
		VALUES($1, $2, $3)
		`, id, id, 0)
		if err != nil {
			return domain.Concept{}, err
		}
		
		// ancestor row
		_, err = tx.Exec(ctx, `
		INSERT INTO concept_paths (ancestor_id, descendant_id, depth)
		SELECT ancestor_id, $1, depth + 1
		FROM concept_paths
		WHERE descendant_id = $2
		`, id, *concept.ParentID)
		if err != nil {
			return domain.Concept{}, err
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return domain.Concept{}, err
	}

	return domain.Concept{
		ID: id,
		UserID: userID,
		ParentID: concept.ParentID,
		Name: concept.Name,
		Description: concept.Description,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

func (r *postgresConceptRepository) GetByID(ctx context.Context, userID string, id string) (domain.Concept, error) {
	var i, u, name string
	var createdAt, updatedAt time.Time
	var parentID, description *string

	err := r.pool.QueryRow(ctx, `
	SELECT id, user_id, parent_id, name, description, created_at, updated_at
	FROM concepts
	WHERE id = $1 AND user_id = $2
	`, id, userID).Scan(&i, &u, &parentID, &name, &description, &createdAt, &updatedAt)
	if err != nil {
		return domain.Concept{}, err
	}

	return domain.Concept{
		ID: i,
		UserID: u,
		ParentID: parentID,
		Name: name,
		Description: description,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
} 
