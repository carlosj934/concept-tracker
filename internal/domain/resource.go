package domain

import (
	"time"
)

type ConceptResource struct {
	ID         string
	ConceptID  string
	UserID     string
	Provider   string
	ExternalID string
	URL        string
	Title      string
	Meta       []byte
	CreatedAt  time.Time
}
