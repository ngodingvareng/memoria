package entity

import (
	"time"

	"github.com/google/uuid"
)

// MomentCircle is a personal Moment shared into a Circle as an optional
// byproduct of mentioning someone — there is no standalone "share to
// circle" action, only this one, reached via the mention flow.
type MomentCircle struct {
	MomentID       uuid.UUID
	CircleID       uuid.UUID
	SharedByUserID *uuid.UUID
	SharedAt       time.Time
}
