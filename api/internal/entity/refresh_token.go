package entity

import (
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	FamilyID     uuid.UUID
	TokenHash    string
	ExpiresAt    time.Time
	RevokedAt    *time.Time
	ReplacedByID *uuid.UUID
	IPAddress    *string
	UserAgent    *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
