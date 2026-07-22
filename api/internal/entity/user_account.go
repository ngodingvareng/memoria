package entity

import "github.com/google/uuid"

type UserAccount struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	AccountID    string
	ProviderID   string
	PasswordHash *string
}
