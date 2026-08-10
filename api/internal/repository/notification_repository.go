package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ngodingvareng/memoria/internal/db"
	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/usecase"
)

var _ usecase.NotificationRepository = (*notificationRepository)(nil)

type notificationRepository struct {
	q *db.Queries
}

func NewNotificationRepository(pool *pgxpool.Pool) *notificationRepository {
	return &notificationRepository{q: db.New(pool)}
}

// Create implements [usecase.NotificationRepository].
func (r *notificationRepository) Create(ctx context.Context, n *entity.Notification) (*entity.Notification, error) {
	payload := n.Payload
	if len(payload) == 0 {
		payload = []byte("{}")
	}

	row, err := r.q.CreateNotification(ctx, db.CreateNotificationParams{
		UserID:                 n.UserID,
		Kind:                   db.NotificationKind(n.Kind),
		ActorUserID:            ptrToPgUUID(n.ActorUserID),
		MomentID:               ptrToPgUUID(n.MomentID),
		ThreadID:               ptrToPgUUID(n.ThreadID),
		CommitmentID:           ptrToPgUUID(n.CommitmentID),
		CommitmentOccurrenceID: ptrToPgUUID(n.CommitmentOccurrenceID),
		CircleID:               ptrToPgUUID(n.CircleID),
		CircleInviteID:         ptrToPgUUID(n.CircleInviteID),
		CircleJoinRequestID:    ptrToPgUUID(n.CircleJoinRequestID),
		RecapID:                ptrToPgUUID(n.RecapID),
		ResurfacingID:          ptrToPgUUID(n.ResurfacingID),
		Payload:                payload,
		DeliverAfter:           timeToPgTimestamptz(n.DeliverAfter),
	})
	if err != nil {
		return nil, fmt.Errorf("create notification: %w", err)
	}
	return toEntityNotification(row), nil
}

// ListByUserID implements [usecase.NotificationRepository].
func (r *notificationRepository) ListByUserID(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]*entity.Notification, error) {
	rows, err := r.q.ListNotificationsByUserID(ctx, db.ListNotificationsByUserIDParams{
		UserID:     userID,
		PageLimit:  limit,
		PageOffset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list notifications by user id: %w", err)
	}
	notifications := make([]*entity.Notification, len(rows))
	for i, row := range rows {
		notifications[i] = toEntityNotification(row)
	}
	return notifications, nil
}

// Count implements [usecase.NotificationRepository].
func (r *notificationRepository) Count(ctx context.Context, userID uuid.UUID) (int64, error) {
	count, err := r.q.CountNotifications(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("count notifications: %w", err)
	}
	return count, nil
}

// CountUnread implements [usecase.NotificationRepository].
func (r *notificationRepository) CountUnread(ctx context.Context, userID uuid.UUID) (int64, error) {
	count, err := r.q.CountUnreadNotifications(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("count unread notifications: %w", err)
	}
	return count, nil
}

// MarkRead implements [usecase.NotificationRepository]. Silent no-op
// (see notifications.sql's MarkNotificationRead) if id/userID don't match
// any unread row.
func (r *notificationRepository) MarkRead(ctx context.Context, id, userID uuid.UUID) error {
	if err := r.q.MarkNotificationRead(ctx, db.MarkNotificationReadParams{ID: id, UserID: userID}); err != nil {
		return fmt.Errorf("mark notification read: %w", err)
	}
	return nil
}

// MarkAllRead implements [usecase.NotificationRepository].
func (r *notificationRepository) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	if err := r.q.MarkAllNotificationsRead(ctx, userID); err != nil {
		return fmt.Errorf("mark all notifications read: %w", err)
	}
	return nil
}
