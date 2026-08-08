package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/enum"
)

// ResponseEvent is the queue behind the daily response digest —
// reactions and comments are never delivered individually. The digest
// job collects everything with DigestedAt nil for a recipient, emits
// one notification, and stamps the rows (FEATURES.md, Notification).
type ResponseEvent struct {
	ID uuid.UUID

	RecipientUserID uuid.UUID // the Moment's owner, i.e. who gets told
	ActorUserID     *uuid.UUID
	MomentID        uuid.UUID

	Kind       enum.ResponseEventKind
	CommentID  *uuid.UUID
	ReactionID *uuid.UUID

	CreatedAt  time.Time
	DigestedAt *time.Time
}
