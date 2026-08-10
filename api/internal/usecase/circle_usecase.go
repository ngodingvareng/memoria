package usecase

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/enum"
)

// --- Repository interface, defined here per the Dependency Rule ---

type CircleRepository interface {
	// Create inserts the Circle row only — the caller is responsible for
	// also seating the creator as admin via AddCreatorAsAdmin, in the
	// same transaction (see CircleUsecase.CreateCircle).
	Create(ctx context.Context, circle *entity.Circle) (*entity.Circle, error)
	AddCreatorAsAdmin(ctx context.Context, circleID, userID uuid.UUID) (*entity.CircleMember, error)
	// Update and Dissolve are admin-only, enforced in the WHERE clause —
	// a non-admin's attempt and a wrong/foreign id are indistinguishable
	// from here, both surface as errs.ErrNotFound.
	Update(ctx context.Context, circle *entity.Circle, userID uuid.UUID) (*entity.Circle, error)
	// UpdateImagePath is the same admin-only gate as Update, scoped to
	// just the profile image.
	UpdateImagePath(ctx context.Context, id, userID uuid.UUID, imagePath *string) (*entity.Circle, error)
	// Dissolve is a sqlc :exec query (no rows-affected count), so a
	// WHERE clause matching zero rows is a silent no-op here.
	Dissolve(ctx context.Context, id, userID uuid.UUID) error
	// GetByID returns errs.ErrNotFound unless userID is an active member.
	GetByID(ctx context.Context, id, userID uuid.UUID) (*entity.Circle, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Circle, error)

	ListMembers(ctx context.Context, circleID uuid.UUID) ([]*entity.CircleMember, error)
	GetActiveMember(ctx context.Context, circleID, userID uuid.UUID) (*entity.CircleMember, error)
	Leave(ctx context.Context, circleID, userID uuid.UUID) error
	RemoveMember(ctx context.Context, circleID, userID, removedByUserID uuid.UUID) error
	UpdateMemberRole(ctx context.Context, circleID, userID, changedByUserID uuid.UUID, role enum.CircleRole) (*entity.CircleMember, error)
	UpdateMemberPermissions(ctx context.Context, circleID, userID, changedByUserID uuid.UUID, canInvite, canCapture bool) (*entity.CircleMember, error)

	// ListSharedCircleIDs is unused by CircleUsecase itself — declared
	// here only because circleRepository is the single implementation
	// backing both CircleRepository and CircleShareChecker below.
	ListSharedCircleIDs(ctx context.Context, userA, userB uuid.UUID) ([]uuid.UUID, error)

	WithTransaction(ctx context.Context, fn func(CircleRepository) error) error
}

// CircleAccessChecker is the minimal capability CircleInviteUsecase and
// CircleJoinRequestUsecase need: confirming the requester is an active
// member (and, at the call site, checking their can_invite flag) before
// letting them act on a Circle's invites/join requests. Method name
// matches CircleRepository.GetActiveMember on purpose — see
// CODING_STANDARDS.md §5.
type CircleAccessChecker interface {
	GetActiveMember(ctx context.Context, circleID, userID uuid.UUID) (*entity.CircleMember, error)
}

// CircleShareChecker is the minimal capability MentionUsecase needs:
// which Circles two users both actively belong to, for the mention
// flow's optional "Share to circle too?" step (FEATURES.md, Mention).
type CircleShareChecker interface {
	ListSharedCircleIDs(ctx context.Context, userA, userB uuid.UUID) ([]uuid.UUID, error)
}

// --- Inputs / outputs ---

type CreateCircleInput struct {
	UserID      uuid.UUID
	Name        string
	Description *string
	ColorHex    *string
	ImagePath   *string
}

type UpdateCircleInput struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Name        string
	Description *string
	ColorHex    *string
	ImagePath   *string
}

type UpdateCircleMemberRoleInput struct {
	CircleID        uuid.UUID
	UserID          uuid.UUID
	ChangedByUserID uuid.UUID
	Role            enum.CircleRole
}

type UpdateCircleMemberPermissionsInput struct {
	CircleID        uuid.UUID
	UserID          uuid.UUID
	ChangedByUserID uuid.UUID
	CanInvite       bool
	CanCapture      bool
}

// --- Usecase ---

type CircleUsecase interface {
	CreateCircle(ctx context.Context, input CreateCircleInput) (*entity.Circle, error)
	UpdateCircle(ctx context.Context, input UpdateCircleInput) (*entity.Circle, error)
	DissolveCircle(ctx context.Context, id, userID uuid.UUID) error
	GetCircle(ctx context.Context, id, userID uuid.UUID) (*entity.Circle, error)
	ListMyCircles(ctx context.Context, userID uuid.UUID) ([]*entity.Circle, error)

	ListMembers(ctx context.Context, circleID, userID uuid.UUID) ([]*entity.CircleMember, error)
	LeaveCircle(ctx context.Context, circleID, userID uuid.UUID) error
	RemoveMember(ctx context.Context, circleID, targetUserID, removedByUserID uuid.UUID) error
	UpdateMemberRole(ctx context.Context, input UpdateCircleMemberRoleInput) (*entity.CircleMember, error)
	UpdateMemberPermissions(ctx context.Context, input UpdateCircleMemberPermissionsInput) (*entity.CircleMember, error)

	// UploadCircleImage replaces the Circle's profile photo, best-effort
	// deleting whichever one it replaces. Admin-only, enforced the same
	// way as UpdateCircle.
	UploadCircleImage(ctx context.Context, input UploadCircleImageInput) (*entity.Circle, error)
}

type UploadCircleImageInput struct {
	CircleID    uuid.UUID
	UserID      uuid.UUID
	FileName    string // client-supplied, used only for its extension
	ContentType string
	Size        int64
	Body        io.Reader
}

type circleUsecase struct {
	repo    CircleRepository
	storage ProfileImageStorage
}

func NewCircleUsecase(repo CircleRepository, storage ProfileImageStorage) CircleUsecase {
	return &circleUsecase{repo: repo, storage: storage}
}

// CreateCircle implements [CircleUsecase]. Seats the creator as admin in
// the same transaction as the insert (FEATURES.md, Circle Permissions:
// "The user who creates a circle becomes its admin").
func (u *circleUsecase) CreateCircle(ctx context.Context, input CreateCircleInput) (*entity.Circle, error) {
	circle := &entity.Circle{
		Name:        input.Name,
		Description: input.Description,
		ColorHex:    input.ColorHex,
		ImagePath:   input.ImagePath,
	}

	var created *entity.Circle
	err := u.repo.WithTransaction(ctx, func(tx CircleRepository) error {
		var txErr error
		created, txErr = tx.Create(ctx, circle)
		if txErr != nil {
			return txErr
		}
		if _, txErr = tx.AddCreatorAsAdmin(ctx, created.ID, input.UserID); txErr != nil {
			return txErr
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("creating circle: %w", err)
	}
	return created, nil
}

// UpdateCircle implements [CircleUsecase]. Full-representation update,
// same PUT convention as ThreadUsecase.UpdateThread.
func (u *circleUsecase) UpdateCircle(ctx context.Context, input UpdateCircleInput) (*entity.Circle, error) {
	circle := &entity.Circle{
		ID:          input.ID,
		Name:        input.Name,
		Description: input.Description,
		ColorHex:    input.ColorHex,
		ImagePath:   input.ImagePath,
	}

	updated, err := u.repo.Update(ctx, circle, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("updating circle: %w", err)
	}
	return updated, nil
}

// UploadCircleImage implements [CircleUsecase]. Admin-only, enforced by
// CircleRepository.UpdateImagePath's own WHERE clause — a non-admin's
// attempt surfaces as errs.ErrNotFound, same as UpdateCircle.
func (u *circleUsecase) UploadCircleImage(ctx context.Context, input UploadCircleImageInput) (*entity.Circle, error) {
	current, err := u.repo.GetByID(ctx, input.CircleID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("getting current circle: %w", err)
	}

	key := buildPublicImageKey("circles", input.CircleID, input.FileName)
	if err := u.storage.Put(ctx, key, input.Body, input.Size, input.ContentType); err != nil {
		return nil, fmt.Errorf("uploading circle image %s: %w", key, err)
	}

	publicURL := u.storage.PublicURL(key)
	updated, err := u.repo.UpdateImagePath(ctx, input.CircleID, input.UserID, &publicURL)
	if err != nil {
		if delErr := u.storage.Delete(context.WithoutCancel(ctx), key); delErr != nil {
			return nil, fmt.Errorf("saving circle image: %w (cleanup also failed: %v)", err, delErr)
		}
		return nil, fmt.Errorf("saving circle image: %w", err)
	}

	if current.ImagePath != nil {
		if oldKey, ok := strings.CutPrefix(*current.ImagePath, u.storage.PublicURL("")); ok {
			_ = u.storage.Delete(context.WithoutCancel(ctx), oldKey)
		}
	}

	return updated, nil
}

// DissolveCircle implements [CircleUsecase].
func (u *circleUsecase) DissolveCircle(ctx context.Context, id, userID uuid.UUID) error {
	if err := u.repo.Dissolve(ctx, id, userID); err != nil {
		return fmt.Errorf("dissolving circle: %w", err)
	}
	return nil
}

// GetCircle implements [CircleUsecase].
func (u *circleUsecase) GetCircle(ctx context.Context, id, userID uuid.UUID) (*entity.Circle, error) {
	circle, err := u.repo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("getting circle: %w", err)
	}
	return circle, nil
}

// ListMyCircles implements [CircleUsecase].
func (u *circleUsecase) ListMyCircles(ctx context.Context, userID uuid.UUID) ([]*entity.Circle, error) {
	circles, err := u.repo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("listing circles: %w", err)
	}
	return circles, nil
}

// ListMembers implements [CircleUsecase]. Requires the caller to
// already be an active member — same access rule as GetCircle.
func (u *circleUsecase) ListMembers(ctx context.Context, circleID, userID uuid.UUID) ([]*entity.CircleMember, error) {
	if _, err := u.repo.GetByID(ctx, circleID, userID); err != nil {
		return nil, fmt.Errorf("checking circle access: %w", err)
	}
	members, err := u.repo.ListMembers(ctx, circleID)
	if err != nil {
		return nil, fmt.Errorf("listing circle members: %w", err)
	}
	return members, nil
}

// LeaveCircle implements [CircleUsecase]: self-service departure.
func (u *circleUsecase) LeaveCircle(ctx context.Context, circleID, userID uuid.UUID) error {
	if err := u.repo.Leave(ctx, circleID, userID); err != nil {
		return fmt.Errorf("leaving circle: %w", err)
	}
	return nil
}

// RemoveMember implements [CircleUsecase]: admin-forced removal.
func (u *circleUsecase) RemoveMember(ctx context.Context, circleID, targetUserID, removedByUserID uuid.UUID) error {
	if err := u.repo.RemoveMember(ctx, circleID, targetUserID, removedByUserID); err != nil {
		return fmt.Errorf("removing circle member: %w", err)
	}
	return nil
}

// UpdateMemberRole implements [CircleUsecase].
func (u *circleUsecase) UpdateMemberRole(ctx context.Context, input UpdateCircleMemberRoleInput) (*entity.CircleMember, error) {
	member, err := u.repo.UpdateMemberRole(ctx, input.CircleID, input.UserID, input.ChangedByUserID, input.Role)
	if err != nil {
		return nil, fmt.Errorf("updating circle member role: %w", err)
	}
	return member, nil
}

// UpdateMemberPermissions implements [CircleUsecase].
func (u *circleUsecase) UpdateMemberPermissions(ctx context.Context, input UpdateCircleMemberPermissionsInput) (*entity.CircleMember, error) {
	member, err := u.repo.UpdateMemberPermissions(ctx, input.CircleID, input.UserID, input.ChangedByUserID, input.CanInvite, input.CanCapture)
	if err != nil {
		return nil, fmt.Errorf("updating circle member permissions: %w", err)
	}
	return member, nil
}
