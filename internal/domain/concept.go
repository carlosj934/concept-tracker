package domain

import (
	"errors"
	"time"
)

type Concept struct {
	ID          string
	UserID      string
	ParentID    *string
	Name        string
	Description *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ConceptPath struct {
	AncestorID   string
	DescendantID string
	Depth        int64
}

type ConceptWithChildren struct {
	Concept
	Children []Concept
}

var ErrNotFound = errors.New("not found")
