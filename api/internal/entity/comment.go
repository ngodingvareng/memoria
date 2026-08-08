package entity

import (
	"time"

	"github.com/google/uuid"
)

// Comment is flat, permanently — there is no parent/reply/depth column
// and there never will be one (FEATURES.md, Comment). CircleID names the
// audience: nil means the mention context (visible to the Moment's
// creator and its mentioned users), set means made inside that Circle.
//
// UserID going nil is what an anonymized comment looks like after its
// author deletes their account — the discussion stays coherent,
// attributed to a former member, deliberately without a name snapshot.
type Comment struct {
	ID        uuid.UUID
	MomentID  uuid.UUID
	UserID    *uuid.UUID
	CircleID  *uuid.UUID
	Body      string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
