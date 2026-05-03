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

func (r *postgresConceptRepository) GetChildren(ctx context.Context, userID string, id string) ([]domain.Concept, error) {
	rows, err := r.pool.Query(ctx, `
	SELECT c.id, c.user_id, c.parent_id, c.name, c.description, c.created_at, c.updated_at
	FROM concepts c
	JOIN concept_paths cp ON cp.descendant_id = c.id
	WHERE cp.ancestor_id = $1
	AND cp.depth = 1
	AND c.user_id = $2
	`, id, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var concepts []domain.Concept
	for rows.Next() {
		var c domain.Concept
		var parentID, description *string
		var createdAt, updatedAt time.Time

		err := rows.Scan(&c.ID, &c.UserID, &parentID, &c.Name, &description, &createdAt, &updatedAt)
		if err != nil {
			return nil, err
		}
		c.ParentID = parentID
		c.Description = description
		c.CreatedAt = createdAt
		c.UpdatedAt = updatedAt
		concepts = append(concepts, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return concepts, nil
}

func (r *postgresConceptRepository) GetSubtree(ctx context.Context, userID string, id string) ([]domain.Concept, error) {
	rows, err := r.pool.Query(ctx, `
	SELECT c.id, c.user_id, c.parent_id, c.name, c.description, c.created_at, c.updated_at
	FROM concepts c
	JOIN concept_paths cp ON cp.descendant_id = c.id
	WHERE cp.ancestor_id = $1 AND cp.depth > 0
	AND c.user_id = $2
	`, id, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var concepts []domain.Concept
	for rows.Next() {
		var c domain.Concept
		var parentID, description *string
		var createdAt, updatedAt time.Time

		err := rows. Scan(&c.ID, &c.UserID, &parentID, &c.Name, &description, &createdAt, &updatedAt)
		if err != nil {
			return nil, err
		}
		c.ParentID = parentID
		c.Description = description
		c.CreatedAt = createdAt
		c.UpdatedAt = updatedAt
		concepts = append(concepts, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return concepts, nil
}

func (r *postgresConceptRepository) Update(ctx context.Context, userID string, id string, name string, description *string) (domain.Concept, error) {
	var u, i , n string
	var d, parentID *string
	var createdAt, updatedAt time.Time

	err := r.pool.QueryRow(ctx, `
	UPDATE concepts
	SET name = $1, description = $2, updated_at = now()
	WHERE id = $3 AND user_id = $4
	RETURNING name, description, user_id, id, parent_id, created_at, updated_at 
	`, name, description, id, userID).Scan(&n, &d, &u, &i, &parentID, &createdAt, &updatedAt)
	if err != nil {
		return domain.Concept{}, err
	}
	
	return domain.Concept{
		ID: i,
		UserID: u,
		ParentID: parentID,
		Name: n,
		Description: d,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

func (r *postgresConceptRepository) Move(ctx context.Context, userID string, id string, newParentID *string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
	UPDATE concepts
	SET parent_id = $1
	WHERE id = $2 AND user_id = $3
	`, newParentID, id, userID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
	DELETE FROM concept_paths
	WHERE descendant_id IN (
		SELECT descendant_id FROM concept_paths WHERE ancestor_id = $1
	)
	AND ancestor_id NOT IN (
		SELECT descendant_id FROM concept_paths WHERE ancestor_id = $1
	)
	`, id)
	if err != nil {
		return err
	}

	if newParentID != nil {
		_, err = tx.Exec(ctx, `
		INSERT INTO concept_paths (ancestor_id, descendant_id, depth)
		SELECT supertree.ancestor_id, subtree.descendant_id, supertree.depth + subtree.depth + 1
		FROM concept_paths subtree
		JOIN concept_paths supertree ON supertree.descendant_id = $2
		WHERE subtree.ancestor_id = $1
		`, id, *newParentID)
		if err != nil {
			return err
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (r *postgresConceptRepository) Delete(ctx context.Context, userID string, id string) error {
	_, err := r.pool.Exec(ctx, `
	DELETE FROM concepts
	WHERE id IN (
		SELECT descendant_id FROM concept_paths WHERE ancestor_id = $1
	)
	AND user_id = $2
	`, id, userID)
	if err != nil {
		return err
	}

	return nil
}
