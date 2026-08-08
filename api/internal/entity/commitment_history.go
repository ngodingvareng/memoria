package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/enum"
)

// CommitmentHistory is an append-only snapshot of each superseded
// version of a Commitment, answering "what were the rules in March?"
// for adherence recomputation. It does not freeze an individual
// occurrence's verdict — CommitmentOccurrence carries its own frozen
// copy of the rule it was judged under.
type CommitmentHistory struct {
	ID           uuid.UUID
	ThreadID     uuid.UUID
	CommitmentID *uuid.UUID

	CronExpression            string
	Timezone                  string
	Strictness                enum.CommitmentStrictness
	ConfirmationWindowMinutes int32

	ActiveFrom  time.Time
	ActiveUntil time.Time

	CreatedAt time.Time
}
