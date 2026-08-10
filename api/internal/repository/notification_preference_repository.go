package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ngodingvareng/memoria/internal/db"
	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/usecase"
)

var _ usecase.NotificationPreferenceRepository = (*notificationPreferenceRepository)(nil)

type notificationPreferenceRepository struct {
	q *db.Queries
}

func NewNotificationPreferenceRepository(pool *pgxpool.Pool) *notificationPreferenceRepository {
	return &notificationPreferenceRepository{q: db.New(pool)}
}

// Get implements [usecase.NotificationPreferenceRepository]. Lazily
// inserts the all-defaults row the first time a user has none, so
// "missing row" never has to mean "off" above this layer.
func (r *notificationPreferenceRepository) Get(ctx context.Context, userID uuid.UUID) (*entity.NotificationPreference, error) {
	row, err := r.q.GetNotificationPreferences(ctx, userID)
	if err == nil {
		return toEntityNotificationPreference(row), nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("get notification preferences: %w", err)
	}

	created, err := r.q.CreateDefaultNotificationPreferences(ctx, userID)
	if err == nil {
		return toEntityNotificationPreference(created), nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("create default notification preferences: %w", err)
	}

	// ON CONFLICT DO NOTHING returned zero rows — lost a create race to a
	// concurrent caller, so the row exists now.
	row, err = r.q.GetNotificationPreferences(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get notification preferences after create race: %w", err)
	}
	return toEntityNotificationPreference(row), nil
}

// Update implements [usecase.NotificationPreferenceRepository]. Callers
// must ensure a row already exists (e.g. via Get) — this is a plain
// UPDATE, not an upsert.
func (r *notificationPreferenceRepository) Update(ctx context.Context, pref *entity.NotificationPreference) (*entity.NotificationPreference, error) {
	row, err := r.q.UpdateNotificationPreferences(ctx, db.UpdateNotificationPreferencesParams{
		CommitmentUpcomingEnabled:        pref.CommitmentUpcomingEnabled,
		MomentDueEnabled:                 pref.MomentDueEnabled,
		MomentMissedEnabled:              pref.MomentMissedEnabled,
		EchoReadyEnabled:                 pref.EchoReadyEnabled,
		RecapGeneratedEnabled:            pref.RecapGeneratedEnabled,
		CircleInviteReceivedEnabled:      pref.CircleInviteReceivedEnabled,
		CircleJoinRequestReceivedEnabled: pref.CircleJoinRequestReceivedEnabled,
		MentionedInMomentEnabled:         pref.MentionedInMomentEnabled,
		AddedToCircleEnabled:             pref.AddedToCircleEnabled,
		CircleJoinRequestApprovedEnabled: pref.CircleJoinRequestApprovedEnabled,
		ResponseDigestEnabled:            pref.ResponseDigestEnabled,
		ResponseDigestHour:               pref.ResponseDigestHour,
		QuietHoursStart:                  ptrToPgTime(pref.QuietHoursStart),
		QuietHoursEnd:                    ptrToPgTime(pref.QuietHoursEnd),
		UserID:                           pref.UserID,
	})
	if err != nil {
		return nil, fmt.Errorf("update notification preferences: %w", err)
	}
	return toEntityNotificationPreference(row), nil
}
