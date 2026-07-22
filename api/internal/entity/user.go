package entity

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID            uuid.UUID
	Name          string
	Email         string
	EmailVerified bool
	Image         *string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
