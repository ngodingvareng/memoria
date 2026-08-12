package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/enum"
)

type MomentRepository interface {
	Create(ctx context.Context, moment *entity.Moment) (*entity.Moment, error)
	// Update overwrites thread_id/occurred_at/note/color_hex/place_name/
	// latitude/longitude for moment.ID, scoped to moment.UserID (the
	// WHERE ... AND user_id = ? doubles as the ownership check, same
	// pattern as ThreadRepository.Update). recorded_at, origin, and
	// client_id are never touched — see UpdateMoment's own comment.
	// Returns errs.ErrNotFound if no row matches.
	Update(ctx context.Context, moment *entity.Moment) (*entity.Moment, error)
	// SoftDelete soft-deletes (sets deleted_at), scoped to userID as the
	// ownership check. The underlying sqlc query is :exec, so a WHERE
	// clause matching zero rows is a silent no-op here rather than
	// surfacing errs.ErrNotFound.
	SoftDelete(ctx context.Context, id, userID uuid.UUID) error
	// SoftDeleteAllByUserID is account deletion's bulk counterpart to
	// SoftDelete — every Moment this user owns, in one statement (see
	// UserUsecase.DeleteAccount).
	SoftDeleteAllByUserID(ctx context.Context, userID uuid.UUID) error
	// GetByID returns errs.ErrNotFound if id/userID don't match any row.
	// Deliberately owner-only — used where that strictness is required
	// (e.g. MomentImageUsecase via MomentAccessChecker, same method).
	GetByID(ctx context.Context, id, userID uuid.UUID) (*entity.Moment, error)
	// GetWithAccess is GetMoment's read path: broader than GetByID,
	// reachable by the owner, an active mentioned user, a Circle it was
	// shared into, or the Circle that natively owns its Thread — same
	// resolution as ResponseAccessChecker (response_access.go), since a
	// viewer who can legitimately comment/react on a Moment must also
	// be able to open it.
	GetWithAccess(ctx context.Context, id, viewerID uuid.UUID) (*entity.Moment, error)
	// ListByUserID is the personal archive timeline: every Moment this
	// user owns, newest occurrence first.
	ListByUserID(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]*entity.Moment, error)
	// ListHomeFeed is the home feed: every Moment reachable by this
	// user — their own, every Moment in a Circle they're an active
	// member of, and every Moment they're actively mentioned in.
	ListHomeFeed(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]*entity.Moment, error)
	// ListByThreadID is a single Thread's browsable timeline. Access to
	// the Thread itself must be checked by the caller.
	ListByThreadID(ctx context.Context, threadID uuid.UUID, limit, offset int32) ([]*entity.Moment, error)
	// ListByCircle is the Circle's Album feed: every Moment in one of
	// the Circle's collaborative Threads, plus every personal Moment
	// shared into it. Membership must be checked by the caller.
	ListByCircle(ctx context.Context, circleID uuid.UUID, limit, offset int32) ([]*entity.Moment, error)
	// Search is the global text Search over one user's own Moments.
	Search(ctx context.Context, userID uuid.UUID, query string, limit, offset int32) ([]*entity.Moment, error)
	// SearchSuggestions returns up to limit of this user's own Moments
	// matching query, for the live-typing search popover
	// (GET /search/suggestions). query is the raw trimmed search text
	// (used for trigram ranking/fallback); prefixQuery is a caller-built
	// ':*'-suffixed tsquery string (see SearchUsecase's buildPrefixTsQuery).
	SearchSuggestions(
		ctx context.Context,
		userID uuid.UUID,
		query, prefixQuery string,
		limit int32,
	) ([]*entity.Moment, error)
	WithTransaction(ctx context.Context, fn func(MomentRepository) error) error
}

// MomentAccessChecker is the minimal capability MomentImageUsecase needs
// from the moment domain: confirming a moment actually belongs to the
// requesting user before letting them touch its images. Method name
// matches MomentRepository.GetByID on purpose — see
// CODING_STANDARDS.md §5, same pattern as ThreadAccessChecker.
type MomentAccessChecker interface {
	GetByID(ctx context.Context, id, userID uuid.UUID) (*entity.Moment, error)
}

type CreateMomentInput struct {
	UserID uuid.UUID
	// ThreadID is nil for a Moment with no Thread (e.g. photo backfill).
	// When set, its ownership/access is checked before the Moment is
	// attached to it.
	ThreadID *uuid.UUID

	// Origin defaults to enum.MomentOriginManual when empty — this
	// usecase only ever serves the manual/import capture path (FEATURES.md:
	// "A manually captured Moment has no state"). Commitment-generated
	// Moments are created from commitment_occurrences instead, elsewhere.
	Origin enum.MomentOrigin

	// OccurredAt is the instant; OccurredUTCOffsetMinutes is the local
	// UTC offset in effect where it happened. Together they derive
	// entity.Moment.OccurredLocal — see deriveOccurredLocal.
	OccurredAt               time.Time
	OccurredUTCOffsetMinutes int16

	// RecordedAt is nil for ordinary online capture (the DB defaults it
	// to NOW()). Offline capture supplies the true local record time at
	// sync — see the recorded_at column comment in the archive schema
	// migration.
	RecordedAt *time.Time

	Note      *string
	ColorHex  *string
	PlaceName *string
	Latitude  *float64
	Longitude *float64

	// ClientID enables idempotent offline sync — a Moment captured
	// offline and synced twice is silently ignored rather than erroring.
	ClientID *uuid.UUID
}

type UpdateMomentInput struct {
	ID       uuid.UUID
	UserID   uuid.UUID
	ThreadID *uuid.UUID

	OccurredAt               time.Time
	OccurredUTCOffsetMinutes int16

	Note      *string
	ColorHex  *string
	PlaceName *string
	Latitude  *float64
	Longitude *float64
}

type ListMomentsInput struct {
	UserID   uuid.UUID
	Page     int32
	PageSize int32
}

type ListThreadMomentsInput struct {
	UserID   uuid.UUID
	ThreadID uuid.UUID
	Page     int32
	PageSize int32
}

type SearchMomentsInput struct {
	UserID   uuid.UUID
	Query    string
	Page     int32
	PageSize int32
}

type ListCircleMomentsInput struct {
	UserID   uuid.UUID
	CircleID uuid.UUID
	Page     int32
	PageSize int32
}

// MomentListResult is shared by ListMoments, ListThreadMoments, and
// SearchMoments — all three are the same "page of Moments" shape, and
// none of the three underlying queries has a matching COUNT query to
// populate a Total the way SearchThreadsResult does.
type MomentListResult struct {
	Moments  []*entity.Moment
	Page     int32
	PageSize int32
}

type MomentUsecase interface {
	CreateMoment(ctx context.Context, input CreateMomentInput) (*entity.Moment, error)
	UpdateMoment(ctx context.Context, input UpdateMomentInput) (*entity.Moment, error)
	SoftDeleteMoment(ctx context.Context, id, userID uuid.UUID) error
	// GetMoment is the general-purpose read path — anyone who can
	// legitimately reach the Moment (owner, active mention, or a Circle
	// context) can fetch it, not just its owner. See
	// MomentRepository.GetWithAccess.
	GetMoment(ctx context.Context, userID, momentID uuid.UUID) (*entity.Moment, error)
	ListMoments(ctx context.Context, input ListMomentsInput) (*MomentListResult, error)
	ListThreadMoments(ctx context.Context, input ListThreadMomentsInput) (*MomentListResult, error)
	// ListCircleMoments returns circleID's Album feed. Returns
	// errs.ErrNotFound if userID isn't an active member.
	ListCircleMoments(ctx context.Context, input ListCircleMomentsInput) (*MomentListResult, error)
	SearchMoments(ctx context.Context, input SearchMomentsInput) (*MomentListResult, error)
}

type momentUsecase struct {
	repo    MomentRepository
	threads ThreadAccessChecker
	circles CircleAccessChecker
}

func NewMomentUsecase(repo MomentRepository, threads ThreadAccessChecker, circles CircleAccessChecker) MomentUsecase {
	return &momentUsecase{repo: repo, threads: threads, circles: circles}
}

// deriveOccurredLocal computes the wall-clock reading at the place a
// Moment happened, per the invariant documented on moments.occurred_local:
// occurred_local = occurred_at AT TIME ZONE 'UTC' + occurred_utc_offset_minutes.
func deriveOccurredLocal(occurredAt time.Time, offsetMinutes int16) time.Time {
	return occurredAt.UTC().Add(time.Duration(offsetMinutes) * time.Minute)
}

// CreateMoment implements [MomentUsecase].
func (u *momentUsecase) CreateMoment(ctx context.Context, input CreateMomentInput) (*entity.Moment, error) {
	if input.ThreadID != nil {
		if _, err := u.threads.GetByID(ctx, *input.ThreadID, input.UserID); err != nil {
			return nil, fmt.Errorf("checking thread ownership: %w", err)
		}
	}

	origin := input.Origin
	if origin == "" {
		origin = enum.MomentOriginManual
	}

	moment := &entity.Moment{
		UserID:                   input.UserID,
		ThreadID:                 input.ThreadID,
		Origin:                   origin,
		OccurredAt:               input.OccurredAt,
		OccurredLocal:            deriveOccurredLocal(input.OccurredAt, input.OccurredUTCOffsetMinutes),
		OccurredUTCOffsetMinutes: input.OccurredUTCOffsetMinutes,
		Note:                     input.Note,
		ColorHex:                 input.ColorHex,
		PlaceName:                input.PlaceName,
		Latitude:                 input.Latitude,
		Longitude:                input.Longitude,
		ClientID:                 input.ClientID,
	}
	if input.RecordedAt != nil {
		moment.RecordedAt = *input.RecordedAt
	}

	var created *entity.Moment
	err := u.repo.WithTransaction(ctx, func(tx MomentRepository) error {
		var txErr error
		created, txErr = tx.Create(ctx, moment)
		return txErr
	})
	if err != nil {
		return nil, fmt.Errorf("creating moment: %w", err)
	}
	return created, nil
}

// UpdateMoment implements [MomentUsecase]. Full-representation update
// (PUT semantics): every call supplies the complete desired state, same
// convention as ThreadUsecase.UpdateThread.
func (u *momentUsecase) UpdateMoment(ctx context.Context, input UpdateMomentInput) (*entity.Moment, error) {
	if input.ThreadID != nil {
		if _, err := u.threads.GetByID(ctx, *input.ThreadID, input.UserID); err != nil {
			return nil, fmt.Errorf("checking thread ownership: %w", err)
		}
	}

	moment := &entity.Moment{
		ID:                       input.ID,
		UserID:                   input.UserID,
		ThreadID:                 input.ThreadID,
		OccurredAt:               input.OccurredAt,
		OccurredLocal:            deriveOccurredLocal(input.OccurredAt, input.OccurredUTCOffsetMinutes),
		OccurredUTCOffsetMinutes: input.OccurredUTCOffsetMinutes,
		Note:                     input.Note,
		ColorHex:                 input.ColorHex,
		PlaceName:                input.PlaceName,
		Latitude:                 input.Latitude,
		Longitude:                input.Longitude,
	}

	var updated *entity.Moment
	err := u.repo.WithTransaction(ctx, func(tx MomentRepository) error {
		var txErr error
		updated, txErr = tx.Update(ctx, moment)
		return txErr
	})
	if err != nil {
		return nil, fmt.Errorf("updating moment: %w", err)
	}
	return updated, nil
}

// SoftDeleteMoment implements [MomentUsecase].
func (u *momentUsecase) SoftDeleteMoment(ctx context.Context, id, userID uuid.UUID) error {
	err := u.repo.WithTransaction(ctx, func(tx MomentRepository) error {
		return tx.SoftDelete(ctx, id, userID)
	})
	if err != nil {
		return fmt.Errorf("soft-deleting moment: %w", err)
	}
	return nil
}

// GetMoment implements [MomentUsecase].
func (u *momentUsecase) GetMoment(ctx context.Context, userID, momentID uuid.UUID) (*entity.Moment, error) {
	moment, err := u.repo.GetWithAccess(ctx, momentID, userID)
	if err != nil {
		return nil, fmt.Errorf("getting moment: %w", err)
	}
	return moment, nil
}

func normalizedPage(page, pageSize int32) (int32, int32) {
	if page < 1 {
		page = defaultPage
	}
	if pageSize < 1 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return page, pageSize
}

// ListMoments implements [MomentUsecase]: the home feed — this user's
// own Moments, every Moment in a Circle they belong to, and every
// Moment they're actively mentioned in (see ListHomeFeed).
func (u *momentUsecase) ListMoments(ctx context.Context, input ListMomentsInput) (*MomentListResult, error) {
	page, pageSize := normalizedPage(input.Page, input.PageSize)

	moments, err := u.repo.ListHomeFeed(ctx, input.UserID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, fmt.Errorf("listing moments: %w", err)
	}
	return &MomentListResult{Moments: moments, Page: page, PageSize: pageSize}, nil
}

// ListThreadMoments implements [MomentUsecase]: a single Thread's
// browsable timeline. Access to the Thread is checked here, before the
// repository is ever asked for its Moments.
func (u *momentUsecase) ListThreadMoments(
	ctx context.Context,
	input ListThreadMomentsInput,
) (*MomentListResult, error) {
	if _, err := u.threads.GetByID(ctx, input.ThreadID, input.UserID); err != nil {
		return nil, fmt.Errorf("checking thread ownership: %w", err)
	}

	page, pageSize := normalizedPage(input.Page, input.PageSize)

	moments, err := u.repo.ListByThreadID(ctx, input.ThreadID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, fmt.Errorf("listing thread moments: %w", err)
	}
	return &MomentListResult{Moments: moments, Page: page, PageSize: pageSize}, nil
}

// ListCircleMoments implements [MomentUsecase]: the Circle's Album feed.
// Membership is checked here, before the repository is ever asked for
// its Moments.
func (u *momentUsecase) ListCircleMoments(
	ctx context.Context,
	input ListCircleMomentsInput,
) (*MomentListResult, error) {
	if _, err := u.circles.GetActiveMember(ctx, input.CircleID, input.UserID); err != nil {
		return nil, fmt.Errorf("checking circle membership: %w", err)
	}

	page, pageSize := normalizedPage(input.Page, input.PageSize)

	moments, err := u.repo.ListByCircle(ctx, input.CircleID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, fmt.Errorf("listing circle moments: %w", err)
	}
	return &MomentListResult{Moments: moments, Page: page, PageSize: pageSize}, nil
}

// SearchMoments implements [MomentUsecase].
func (u *momentUsecase) SearchMoments(ctx context.Context, input SearchMomentsInput) (*MomentListResult, error) {
	page, pageSize := normalizedPage(input.Page, input.PageSize)

	moments, err := u.repo.Search(ctx, input.UserID, input.Query, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, fmt.Errorf("searching moments: %w", err)
	}
	return &MomentListResult{Moments: moments, Page: page, PageSize: pageSize}, nil
}
