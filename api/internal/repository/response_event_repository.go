package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ngodingvareng/memoria/internal/db"
	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/usecase"
)

var _ usecase.ResponseEventRepository = (*responseEventRepository)(nil)

// responseEventRepository writes to the batched daily digest queue
// (FEATURES.md, Notification). It only ever writes — nothing reads or
// digests response_events yet, since that's the Notification domain's
// worker, not built (see the response_events.sql query comment).
type responseEventRepository struct {
	q *db.Queries
}

func NewResponseEventRepository(pool *pgxpool.Pool) *responseEventRepository {
	return &responseEventRepository{q: db.New(pool)}
}

// Create implements [usecase.ResponseEventRepository].
func (r *responseEventRepository) Create(ctx context.Context, event *entity.ResponseEvent) error {
	if err := r.q.CreateResponseEvent(ctx, db.CreateResponseEventParams{
		RecipientUserID: event.RecipientUserID,
		ActorUserID:     ptrToPgUUID(event.ActorUserID),
		MomentID:        event.MomentID,
		Kind:            db.ResponseEventKind(event.Kind),
		CommentID:       ptrToPgUUID(event.CommentID),
		ReactionID:      ptrToPgUUID(event.ReactionID),
	}); err != nil {
		return fmt.Errorf("create response event: %w", err)
	}
	return nil
}
