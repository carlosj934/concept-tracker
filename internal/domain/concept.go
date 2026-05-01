package domain

import (
	"time"
)

type Concept struct {
	ID string
	UserID string
	ParentID *string
	Name string
	Description *string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ConceptPath struct {
	AncestorID string
	DescendantID string
	Depth int64
}
