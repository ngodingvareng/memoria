package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/entity"
)

// defaultConfirmationTimeoutMinutes mirrors the DB column's own DEFAULT
// 1440. It has to be applied here explicitly (not just left to Postgres)
// because CreateActivity's sqlc query lists confirmation_timeout_minutes
// in its explicit column list — inserting NULL there sets the column to
// NULL, it does NOT fall back to the column's DEFAULT. Only omitting the
// column entirely triggers the DEFAULT, which our fixed sqlc query can't
// do conditionally.
const defaultConfirmationTimeoutMinutes int32 = 1440

// defaultColorHex is taken from https://tailwindcss.com/docs/colors (Gray 700)
// We recommended using using dark gray and similar tones from that
// source so it matches  seamlessly with the frontend.
const defaultColorHex string = "#374151"

type ActivityRepository interface {
	Create(ctx context.Context, activity *entity.Activity) (*entity.Activity, error)
	// Update overwrites name/description/color_hex/confirmation_timeout_minutes
	// for activity.ID, scoped to activity.UserID (the WHERE ... AND user_id = ?
	// in UpdateActivity doubles as the ownership check — same pattern as
	// GetActivityByID). Returns errs.ErrNotFound if no row matches (wrong
	// owner, wrong id, or already soft-deleted).
	Update(ctx context.Context, activity *entity.Activity) (*entity.Activity, error)
	// SoftDelete soft-deletes (sets deleted_at), scoped to userID as the
	// ownership check. Mirrors ActivityImageRepository.Delete: the
	// underlying sqlc query is :exec, so a WHERE clause matching zero
	// rows (wrong owner/id, or already deleted) is a silent no-op here
	// rather than surfacing errs.ErrNotFound.
	SoftDelete(ctx context.Context, id, userID uuid.UUID) error
	WithTransaction(ctx context.Context, fn func(ActivityRepository) error) error
}

type CreateActivityInput struct {
	UserID                     uuid.UUID
	Name                       string
	Description                *string
	IsFixedSchedule            bool
	ColorHex                   *string
	ConfirmationTimeoutMinutes *int32
}

// UpdateActivityInput deliberately has no IsFixedSchedule — flipping that
// flag is its own endpoint (SetActivityFixedSchedule query already
// exists, usecase/handler not implemented yet; see TODO.md) since
// toggling it has side effects on activity_schedules that a plain field
// update shouldn't trigger implicitly.
type UpdateActivityInput struct {
	ID                         uuid.UUID
	UserID                     uuid.UUID
	Name                       string
	Description                *string
	ColorHex                   *string
	ConfirmationTimeoutMinutes *int32
}

type ActivityUsecase interface {
	CreateActivity(ctx context.Context, input CreateActivityInput) (*entity.Activity, error)
	UpdateActivity(ctx context.Context, input UpdateActivityInput) (*entity.Activity, error)
	SoftDeleteActivity(ctx context.Context, id, userID uuid.UUID) error
}

type activityUsecase struct {
	repo ActivityRepository
}

func NewActivityUsecase(repo ActivityRepository) ActivityUsecase {
	return &activityUsecase{repo: repo}
}

// CreateActivity implements [ActivityUsecase].
func (u *activityUsecase) CreateActivity(ctx context.Context, input CreateActivityInput) (*entity.Activity, error) {
	timeout := defaultConfirmationTimeoutMinutes
	if input.ConfirmationTimeoutMinutes != nil {
		timeout = *input.ConfirmationTimeoutMinutes
	}

	colorHex := defaultColorHex
	if input.ColorHex != nil {
		colorHex = *input.ColorHex
	}

	activity := &entity.Activity{
		UserID:                     input.UserID,
		Name:                       input.Name,
		Description:                input.Description,
		IsFixedSchedule:            input.IsFixedSchedule,
		ColorHex:                   &colorHex,
		ConfirmationTimeoutMinutes: &timeout,
	}

	var created *entity.Activity
	err := u.repo.WithTransaction(ctx, func(tx ActivityRepository) error {
		var txErr error
		created, txErr = tx.Create(ctx, activity)
		return txErr
	})
	if err != nil {
		return nil, fmt.Errorf("creating activity: %w", err)
	}

	return created, nil
}

// UpdateActivity implements [ActivityUsecase]. Same defaulting rules as
// CreateActivity: a nil ColorHex/ConfirmationTimeoutMinutes falls back
// to the same defaults rather than being sent through as NULL — this is
// a full-representation update (PUT semantics), not a partial patch, so
// every call is expected to supply the complete desired state.
func (u *activityUsecase) UpdateActivity(ctx context.Context, input UpdateActivityInput) (*entity.Activity, error) {
	timeout := defaultConfirmationTimeoutMinutes
	if input.ConfirmationTimeoutMinutes != nil {
		timeout = *input.ConfirmationTimeoutMinutes
	}

	colorHex := defaultColorHex
	if input.ColorHex != nil {
		colorHex = *input.ColorHex
	}

	activity := &entity.Activity{
		ID:                         input.ID,
		UserID:                     input.UserID,
		Name:                       input.Name,
		Description:                input.Description,
		ColorHex:                   &colorHex,
		ConfirmationTimeoutMinutes: &timeout,
	}

	var updated *entity.Activity
	err := u.repo.WithTransaction(ctx, func(tx ActivityRepository) error {
		var txErr error
		updated, txErr = tx.Update(ctx, activity)
		return txErr
	})
	if err != nil {
		return nil, fmt.Errorf("updating activity: %w", err)
	}

	return updated, nil
}

// SoftDeleteActivity implements [ActivityUsecase].
func (u *activityUsecase) SoftDeleteActivity(ctx context.Context, id, userID uuid.UUID) error {
	err := u.repo.WithTransaction(ctx, func(tx ActivityRepository) error {
		return tx.SoftDelete(ctx, id, userID)
	})
	if err != nil {
		return fmt.Errorf("soft-deleting activity: %w", err)
	}
	return nil
}
