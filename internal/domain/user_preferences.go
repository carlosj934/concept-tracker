package domain

import (
	"time"
)

type UserPreferences struct {
	UserID    string
	Timezone  string
	UpdatedAt time.Time
}
