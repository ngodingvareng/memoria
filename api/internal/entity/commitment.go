package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/enum"
)

// Commitment is the opt-in mode where a Thread takes on a schedule
// (FEATURES.md). A Thread may carry more than one, each fully
// independent. PausedAt/ArchivedAt together are the exit ramp and
// replace the old threads.has_commitment flag: "does this Thread have
// an active Commitment" is just "any row where both are nil."
type Commitment struct {
	ID       uuid.UUID
	ThreadID uuid.UUID
	Name     string

	CronExpression string
	Timezone       string

	Strictness                enum.CommitmentStrictness
	ConfirmationWindowMinutes int32

	// The opt-in contract: cannot exist without a recorded, explicit
	// consent event. ConsentVersion pins which wording was agreed to.
	ConsentedAt    time.Time
	ConsentVersion int16

	// Scoped to this Commitment, part of what the user opted into.
	// NotifyMissed is off by default per FEATURES.md, Notification.
	NotifyUpcoming bool
	NotifyDue      bool
	NotifyMissed   bool

	PausedAt   *time.Time
	ArchivedAt *time.Time

	// Scheduler bookmark: the last occurrence instant already generated.
	LastGeneratedAt *time.Time

	CreatedAt time.Time
	UpdatedAt time.Time

	// Optimistic locking for concurrent edits.
	Version int32
}
