package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/entity"
)

// defaultConfirmationTimeoutMinutes mirrors the DB column's own DEFAULT
// 1440. It has to be applied here explicitly (not just left to Postgres)
// because CreateThread's sqlc query lists confirmation_timeout_minutes
// in its explicit column list — inserting NULL there sets the column to
// NULL, it does NOT fall back to the column's DEFAULT. Only omitting the
// column entirely triggers the DEFAULT, which our fixed sqlc query can't
// do conditionally.
const defaultConfirmationTimeoutMinutes int32 = 1440

// defaultColorHex is taken from https://tailwindcss.com/docs/colors (Gray 700)
// We recommended using using dark gray and similar tones from that
// source so it matches  seamlessly with the frontend.
const defaultColorHex string = "#374151"

const (
	defaultPage     int32 = 1
	defaultPageSize int32 = 20
	maxPageSize     int32 = 100
)

type SearchThreadsParams struct {
	UserID        uuid.UUID
	Name          *string
	HasCommitment *bool
	Limit         int32
	Offset        int32
}

type ThreadRepository interface {
	Create(ctx context.Context, thread *entity.Thread) (*entity.Thread, error)
	// Update overwrites name/description/color_hex/confirmation_timeout_minutes
	// for thread.ID, scoped to thread.UserID (the WHERE ... AND user_id = ?
	// in UpdateThread doubles as the ownership check — same pattern as
	// GetThreadByID). Returns errs.ErrNotFound if no row matches (wrong
	// owner, wrong id, or already soft-deleted).
	Update(ctx context.Context, thread *entity.Thread) (*entity.Thread, error)
	// SoftDelete soft-deletes (sets deleted_at), scoped to userID as the
	// ownership check. Mirrors ThreadImageRepository.Delete: the
	// underlying sqlc query is :exec, so a WHERE clause matching zero
	// rows (wrong owner/id, or already deleted) is a silent no-op here
	// rather than surfacing errs.ErrNotFound.
	SoftDelete(ctx context.Context, id, userID uuid.UUID) error
	Search(ctx context.Context, params SearchThreadsParams) ([]*entity.Thread, int64, error)
	GetByID(ctx context.Context, id, userID uuid.UUID) (*entity.Thread, error)
	WithTransaction(ctx context.Context, fn func(ThreadRepository) error) error
}

type CreateThreadInput struct {
	UserID                     uuid.UUID
	Name                       string
	Description                *string
	HasCommitment              bool
	ColorHex                   *string
	ConfirmationTimeoutMinutes *int32
}

// UpdateThreadInput deliberately has no HasCommitment — flipping that
// flag is its own endpoint (SetThreadHasCommitment query already
// exists, usecase/handler not implemented yet; see TODO.md) since
// toggling it has side effects on commitments that a plain field
// update shouldn't trigger implicitly.
type UpdateThreadInput struct {
	ID                         uuid.UUID
	UserID                     uuid.UUID
	Name                       string
	Description                *string
	ColorHex                   *string
	ConfirmationTimeoutMinutes *int32
}

type SearchThreadsInput struct {
	UserID        uuid.UUID
	Name          *string
	HasCommitment *bool
	Page          int32
	PageSize      int32
}

type SearchThreadsResult struct {
	Threads  []*entity.Thread
	Total    int64
	Page     int32
	PageSize int32
}

type ThreadUsecase interface {
	CreateThread(ctx context.Context, input CreateThreadInput) (*entity.Thread, error)
	UpdateThread(ctx context.Context, input UpdateThreadInput) (*entity.Thread, error)
	SoftDeleteThread(ctx context.Context, id, userID uuid.UUID) error
	GetThread(ctx context.Context, userID, threadID uuid.UUID) (*entity.Thread, error)
	SearchThreads(ctx context.Context, input SearchThreadsInput) (*SearchThreadsResult, error)
}

type threadUsecase struct {
	repo ThreadRepository
}

func NewThreadUsecase(repo ThreadRepository) ThreadUsecase {
	return &threadUsecase{repo: repo}
}

// CreateThread implements [ThreadUsecase].
func (u *threadUsecase) CreateThread(ctx context.Context, input CreateThreadInput) (*entity.Thread, error) {
	timeout := defaultConfirmationTimeoutMinutes
	if input.ConfirmationTimeoutMinutes != nil {
		timeout = *input.ConfirmationTimeoutMinutes
	}

	colorHex := defaultColorHex
	if input.ColorHex != nil {
		colorHex = *input.ColorHex
	}

	thread := &entity.Thread{
		UserID:                     input.UserID,
		Name:                       input.Name,
		Description:                input.Description,
		HasCommitment:              input.HasCommitment,
		ColorHex:                   &colorHex,
		ConfirmationTimeoutMinutes: &timeout,
	}

	var created *entity.Thread
	err := u.repo.WithTransaction(ctx, func(tx ThreadRepository) error {
		var txErr error
		created, txErr = tx.Create(ctx, thread)
		return txErr
	})
	if err != nil {
		return nil, fmt.Errorf("creating thread: %w", err)
	}

	return created, nil
}

// UpdateThread implements [ThreadUsecase]. Same defaulting rules as
// CreateThread: a nil ColorHex/ConfirmationTimeoutMinutes falls back
// to the same defaults rather than being sent through as NULL — this is
// a full-representation update (PUT semantics), not a partial patch, so
// every call is expected to supply the complete desired state.
func (u *threadUsecase) UpdateThread(ctx context.Context, input UpdateThreadInput) (*entity.Thread, error) {
	timeout := defaultConfirmationTimeoutMinutes
	if input.ConfirmationTimeoutMinutes != nil {
		timeout = *input.ConfirmationTimeoutMinutes
	}

	colorHex := defaultColorHex
	if input.ColorHex != nil {
		colorHex = *input.ColorHex
	}

	thread := &entity.Thread{
		ID:                         input.ID,
		UserID:                     input.UserID,
		Name:                       input.Name,
		Description:                input.Description,
		ColorHex:                   &colorHex,
		ConfirmationTimeoutMinutes: &timeout,
	}

	var updated *entity.Thread
	err := u.repo.WithTransaction(ctx, func(tx ThreadRepository) error {
		var txErr error
		updated, txErr = tx.Update(ctx, thread)
		return txErr
	})
	if err != nil {
		return nil, fmt.Errorf("updating thread: %w", err)
	}

	return updated, nil
}

// SoftDeleteThread implements [ThreadUsecase].
func (u *threadUsecase) SoftDeleteThread(ctx context.Context, id, userID uuid.UUID) error {
	err := u.repo.WithTransaction(ctx, func(tx ThreadRepository) error {
		return tx.SoftDelete(ctx, id, userID)
	})
	if err != nil {
		return fmt.Errorf("soft-deleting thread: %w", err)
	}
	return nil
}

// GetThread implements [ThreadUsecase]. Ownership is enforced at the
// repository/query level (threads.user_id in the WHERE clause), so a
// mismatched userID simply surfaces as errs.ErrNotFound.
func (u *threadUsecase) GetThread(ctx context.Context, userID, threadID uuid.UUID) (*entity.Thread, error) {
	thread, err := u.repo.GetByID(ctx, threadID, userID)
	if err != nil {
		return nil, fmt.Errorf("getting thread: %w", err)
	}
	return thread, nil
}

// SearchThreads implements [ThreadUsecase].
func (u *threadUsecase) SearchThreads(ctx context.Context, input SearchThreadsInput) (*SearchThreadsResult, error) {
	page := input.Page
	if page < 1 {
		page = defaultPage
	}
	pageSize := input.PageSize
	if pageSize < 1 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	threads, total, err := u.repo.Search(ctx, SearchThreadsParams{
		UserID:        input.UserID,
		Name:          input.Name,
		HasCommitment: input.HasCommitment,
		Limit:         pageSize,
		Offset:        (page - 1) * pageSize,
	})
	if err != nil {
		return nil, fmt.Errorf("searching threads: %w", err)
	}

	return &SearchThreadsResult{
		Threads:  threads,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}
