package entity

import (
	"time"

	"github.com/google/uuid"
)

type Circle struct {
	ID              uuid.UUID
	Name            string
	Description     *string
	ColorHex        *string
	ImagePath       *string
	CreatedByUserID *uuid.UUID
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DissolvedAt     *time.Time
}
