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

type ActivityUsecase interface {
	CreateActivity(ctx context.Context, input CreateActivityInput) (*entity.Activity, error)
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
