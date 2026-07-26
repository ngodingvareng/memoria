package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/enum"
)

type UserAccount struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	AccountID    string
	ProviderID   enum.AuthProvider
	PasswordHash *string // only set for ProviderID == AuthProviderCredential
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
