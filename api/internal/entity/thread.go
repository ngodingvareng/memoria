package entity

import (
	"time"

	"github.com/google/uuid"
)

// Thread is personal when CircleID is nil (owned by UserID) and
// collaborative when it is set (owned by the Circle; access is governed
// by CircleMember, not UserID).
//
// threads.search_document has no field here: it's a tsvector search
// index, not a domain value, and stays a repository/mapper concern.
type Thread struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	CircleID    *uuid.UUID
	Name        string
	Description *string
	ColorHex    *string
	SortOrder   int32
	ArchivedAt  *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}
