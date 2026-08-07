package entity

import (
	"time"

	"github.com/google/uuid"
)

type ThreadImage struct {
	ID        uuid.UUID
	ThreadID  uuid.UUID
	ImagePath string
	ImageAlt  *string
	CreatedAt time.Time
}
