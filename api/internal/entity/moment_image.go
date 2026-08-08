package entity

import (
	"time"

	"github.com/google/uuid"
)

type MomentImage struct {
	ID               uuid.UUID
	MomentID         uuid.UUID
	ImagePath        string
	ImageAlt         *string
	ContentType      *string
	ByteSize         *int64
	Width            *int32
	Height           *int32
	MetadataStripped bool
	SortOrder        int32
	CreatedAt        time.Time
}
