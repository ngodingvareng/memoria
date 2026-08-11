package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/entity"
)

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
	Archived      *bool
	SortByRecency bool
	Limit         int32
	Offset        int32
}

type ThreadRepository interface {
	Create(ctx context.Context, thread *entity.Thread) (*entity.Thread, error)
	// Update overwrites name/description/color_hex for thread.ID, scoped
	// to thread.UserID (the WHERE ... AND user_id = ? in UpdateThread
	// doubles as the ownership check — same pattern as GetByID). Returns
	// errs.ErrNotFound if no row matches (wrong owner, wrong id, or
	// already soft-deleted).
	Update(ctx context.Context, thread *entity.Thread) (*entity.Thread, error)
	// SoftDelete soft-deletes (sets deleted_at), scoped to userID as the
	// ownership check. Mirrors ThreadImageRepository.Delete: the
	// underlying sqlc query is :exec, so a WHERE clause matching zero
	// rows (wrong owner/id, or already deleted) is a silent no-op here
	// rather than surfacing errs.ErrNotFound.
	SoftDelete(ctx context.Context, id, userID uuid.UUID) error
	// SoftDeleteAllByUserID is account deletion's bulk counterpart to
	// SoftDelete — every Thread this user owns (see
	// UserUsecase.DeleteAccount). Moments other Circle members captured
	// into one of these Threads are untouched.
	SoftDeleteAllByUserID(ctx context.Context, userID uuid.UUID) error
	Search(ctx context.Context, params SearchThreadsParams) ([]*entity.Thread, int64, error)
	// SearchSuggestions returns up to limit reachable Threads matching
	// query, for the live-typing search popover (GET /search/suggestions).
	// No pagination/count — a fixed-size top-N, not a full page. query is
	// the raw trimmed search text (used for trigram ranking/fallback);
	// prefixQuery is a caller-built ':*'-suffixed tsquery string (see
	// SearchUsecase's buildPrefixTsQuery).
	SearchSuggestions(ctx context.Context, userID uuid.UUID, query, prefixQuery string, limit int32) ([]*entity.Thread, error)
	// GetByID also grants access to a collaborative Thread's Circle
	// members, not just its creator — see threads.GetThreadWithAccess.
	GetByID(ctx context.Context, id, userID uuid.UUID) (*entity.Thread, error)
	// ListByCircle returns a Circle's collaborative Threads. Callers are
	// responsible for the Circle-membership check (see ListCircleThreads)
	// — this query has no user-scoping of its own.
	ListByCircle(ctx context.Context, circleID uuid.UUID) ([]*entity.Thread, error)
	WithTransaction(ctx context.Context, fn func(ThreadRepository) error) error
}

// CreateThreadInput's CircleID is nil for a personal Thread (owned by
// UserID) and set to create a collaborative Thread owned by that Circle
// instead — CreateThread checks the caller is an active member of it
// first (FEATURES.md, Circle: "Users can create collaborative Threads
// that any member can contribute to" — no admin/permission gate beyond
// active membership).
type CreateThreadInput struct {
	UserID      uuid.UUID
	CircleID    *uuid.UUID
	Name        string
	Description *string
	ColorHex    *string
}

type UpdateThreadInput struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Name        string
	Description *string
	ColorHex    *string
}

type SearchThreadsInput struct {
	UserID        uuid.UUID
	Name          *string
	Archived      *bool
	SortByRecency bool
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
	// ListCircleThreads returns circleID's collaborative Threads.
	// Returns errs.ErrNotFound if userID isn't an active member.
	ListCircleThreads(ctx context.Context, circleID, userID uuid.UUID) ([]*entity.Thread, error)
}

type threadUsecase struct {
	repo    ThreadRepository
	circles CircleAccessChecker
}

func NewThreadUsecase(repo ThreadRepository, circles CircleAccessChecker) ThreadUsecase {
	return &threadUsecase{repo: repo, circles: circles}
}

// CreateThread implements [ThreadUsecase].
func (u *threadUsecase) CreateThread(ctx context.Context, input CreateThreadInput) (*entity.Thread, error) {
	colorHex := defaultColorHex
	if input.ColorHex != nil {
		colorHex = *input.ColorHex
	}

	if input.CircleID != nil {
		if _, err := u.circles.GetActiveMember(ctx, *input.CircleID, input.UserID); err != nil {
			return nil, fmt.Errorf("checking circle membership: %w", err)
		}
	}

	thread := &entity.Thread{
		UserID:      input.UserID,
		CircleID:    input.CircleID,
		Name:        input.Name,
		Description: input.Description,
		ColorHex:    &colorHex,
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

// UpdateThread implements [ThreadUsecase]. Same defaulting rule as
// CreateThread: a nil ColorHex falls back to the same default rather
// than being sent through as NULL — this is a full-representation
// update (PUT semantics), not a partial patch, so every call is
// expected to supply the complete desired state.
func (u *threadUsecase) UpdateThread(ctx context.Context, input UpdateThreadInput) (*entity.Thread, error) {
	colorHex := defaultColorHex
	if input.ColorHex != nil {
		colorHex = *input.ColorHex
	}

	thread := &entity.Thread{
		ID:          input.ID,
		UserID:      input.UserID,
		Name:        input.Name,
		Description: input.Description,
		ColorHex:    &colorHex,
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
		Archived:      input.Archived,
		SortByRecency: input.SortByRecency,
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

// ListCircleThreads implements [ThreadUsecase].
func (u *threadUsecase) ListCircleThreads(ctx context.Context, circleID, userID uuid.UUID) ([]*entity.Thread, error) {
	if _, err := u.circles.GetActiveMember(ctx, circleID, userID); err != nil {
		return nil, fmt.Errorf("checking circle membership: %w", err)
	}

	threads, err := u.repo.ListByCircle(ctx, circleID)
	if err != nil {
		return nil, fmt.Errorf("listing circle threads: %w", err)
	}
	return threads, nil
}
