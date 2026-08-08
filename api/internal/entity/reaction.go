package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/enum"
)

// Reaction never has a counter anywhere in the domain, on purpose: who
// reacted is a row; how many is never materialized and never returned
// (FEATURES.md, Reaction). See Comment for the meaning of CircleID/UserID.
type Reaction struct {
	ID        uuid.UUID
	MomentID  uuid.UUID
	UserID    *uuid.UUID
	CircleID  *uuid.UUID
	Kind      enum.ReactionKind
	CreatedAt time.Time
}
