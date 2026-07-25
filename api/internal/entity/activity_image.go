package entity

import (
	"time"

	"github.com/google/uuid"
)

type ActivityImage struct {
	ID         uuid.UUID
	ActivityID uuid.UUID
	ImagePath  string
	ImageAlt   *string
	CreatedAt  time.Time
}
