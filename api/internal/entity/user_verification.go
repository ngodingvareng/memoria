package entity

import (
	"time"

	"github.com/google/uuid"
)

type UserVerification struct {
	ID         uuid.UUID
	Identifier string
	Value      string
	ExpiresAt  time.Time
	ConsumedAt *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
