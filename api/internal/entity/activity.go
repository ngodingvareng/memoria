package entity

import (
	"time"

	"github.com/google/uuid"
)

type Activity struct {
	ID                         uuid.UUID
	UserID                     uuid.UUID
	Name                       string
	Description                *string
	IsFixedSchedule            bool
	ColorPalette               *string
	ConfirmationTimeoutMinutes *int32
	CreatedAt                  time.Time
	UpdatedAt                  time.Time
	DeletedAt                  *time.Time
}
