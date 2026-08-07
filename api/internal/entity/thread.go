package entity

import (
	"time"

	"github.com/google/uuid"
)

type Thread struct {
	ID                         uuid.UUID
	UserID                     uuid.UUID
	Name                       string
	Description                *string
	HasCommitment              bool
	ColorHex                   *string
	ConfirmationTimeoutMinutes *int32
	CreatedAt                  time.Time
	UpdatedAt                  time.Time
	DeletedAt                  *time.Time
}
