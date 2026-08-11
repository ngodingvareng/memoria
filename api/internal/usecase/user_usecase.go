package usecase

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path"

	"github.com/google/uuid"

	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/enum"
	"github.com/ngodingvareng/memoria/internal/errs"
)

// ProfileImageStorage is the minimal capability UserUsecase and
// CircleUsecase need for profile-photo uploads — declared once since
// both need the identical signature (CODING_STANDARDS.md §5). Public
// photos only, unlike thread/moment images' Storage: Put + PublicURL,
// no PresignGet — see storage.Storage's PublicPrefixes doc for why
// these specific uploads never need signing.
type ProfileImageStorage interface {
	Put(ctx context.Context, key string, body io.Reader, size int64, contentType string) error
	PublicURL(key string) string
	Delete(ctx context.Context, key string) error
}

// buildPublicImageKey deliberately ignores the client-supplied filename
// except for its extension — using it directly would be a
// path-traversal risk (mirrors thread_image_usecase.go's buildImageKey).
func buildPublicImageKey(resource string, id uuid.UUID, clientFileName string) string {
	return fmt.Sprintf("%s/%s/%s%s", resource, id, uuid.NewString(), path.Ext(clientFileName))
}

type UploadProfileImageInput struct {
	UserID      uuid.UUID
	FileName    string // client-supplied, used only for its extension
	ContentType string
	Size        int64
	Body        io.Reader
}

// UpdatePrivacySettingsInput carries every Social Interaction + Data
// Controls toggle (FEATURES.md, Privacy & Control) — saved together as
// one settings-screen submission, not individually.
type UpdatePrivacySettingsInput struct {
	UserID                 uuid.UUID
	MentionPolicy          enum.AudiencePolicy
	CircleInvitePolicy     enum.AudiencePolicy
	DiscoverableByUsername bool
	StripPhotoMetadata     bool
}

// UserUsecase backs the post-register onboarding step, where a newly
// created account (no username yet) claims one.
type UserUsecase interface {
	CheckUsernameAvailability(ctx context.Context, username string) (bool, error)
	// SetUsername returns errs.ErrUsernameAlreadyExists if the username
	// was taken between the last availability check and this call.
	SetUsername(ctx context.Context, userID uuid.UUID, username string) (*entity.User, error)
	// GetPublicProfile resolves a user_id to the minimal public-facing
	// fields (name, username, image) other users are shown — e.g.
	// rendering a Circle's member list. Returns errs.ErrNotFound if no
	// such user exists.
	GetPublicProfile(ctx context.Context, id uuid.UUID) (*entity.User, error)
	// GetPublicProfileByUsername is GetPublicProfile's counterpart for
	// the @username profile page, which only has the username to go on.
	GetPublicProfileByUsername(ctx context.Context, username string) (*entity.User, error)
	// GetOwnProfile is the caller's own full profile, including the
	// privacy fields GetPublicProfile/GetPublicProfileByUsername
	// deliberately omit — backs GET /users/me and settings-screen seed
	// values.
	GetOwnProfile(ctx context.Context, id uuid.UUID) (*entity.User, error)
	// UpdatePrivacySettings implements the Social Interaction + Data
	// Controls settings screen.
	UpdatePrivacySettings(ctx context.Context, input UpdatePrivacySettingsInput) (*entity.User, error)
	// MarkUserKnown/UnmarkUserKnown/ListKnownUsers back the "Known
	// people" settings surface (FEATURES.md, Privacy & Control's
	// "known" audience tier). Marking is one-directional and silent
	// toward the known user — they're never notified and there's no
	// reverse-direction query (who marked *me*). Mark/Unmark resolve
	// username the same way BlockUser does, and are idempotent (see
	// UserKnownRepository).
	MarkUserKnown(ctx context.Context, knowerUserID uuid.UUID, username string) error
	UnmarkUserKnown(ctx context.Context, knowerUserID uuid.UUID, username string) error
	ListKnownUsers(ctx context.Context, knowerUserID uuid.UUID) ([]*entity.User, error)
	// UploadProfileImage replaces the caller's own profile photo,
	// best-effort deleting whichever one it replaces.
	UploadProfileImage(ctx context.Context, input UploadProfileImageInput) (*entity.User, error)

	// BlockUser/UnblockUser/ListBlockedUsers back the "Blocked users"
	// settings surface (FEATURES.md, Privacy & Control). Block/Unblock
	// resolve username the same way MarkUserKnown does, and are
	// idempotent (see UserBlockRepository).
	BlockUser(ctx context.Context, blockerUserID uuid.UUID, username string) error
	UnblockUser(ctx context.Context, blockerUserID uuid.UUID, username string) error
	ListBlockedUsers(ctx context.Context, blockerUserID uuid.UUID) ([]*entity.User, error)

	// MuteUser/UnmuteUser/ListMutedUsers back the "Muted users" settings
	// surface — a view filter only, never an access gate.
	MuteUser(ctx context.Context, muterUserID uuid.UUID, username string) error
	UnmuteUser(ctx context.Context, muterUserID uuid.UUID, username string) error
	ListMutedUsers(ctx context.Context, muterUserID uuid.UUID) ([]*entity.User, error)
}

type userUsecase struct {
	users   UserRepository
	knowns  UserKnownRepository
	blocks  UserBlockRepository
	mutes   UserMuteRepository
	storage ProfileImageStorage
}

func NewUserUsecase(users UserRepository, knowns UserKnownRepository, blocks UserBlockRepository, mutes UserMuteRepository, storage ProfileImageStorage) UserUsecase {
	return &userUsecase{users: users, knowns: knowns, blocks: blocks, mutes: mutes, storage: storage}
}

// resolveUsers looks up each id's profile, silently skipping any that no
// longer resolve (e.g. a since-deleted account) rather than failing the
// whole list over one stale row.
func (u *userUsecase) resolveUsers(ctx context.Context, ids []uuid.UUID) ([]*entity.User, error) {
	users := make([]*entity.User, 0, len(ids))
	for _, id := range ids {
		user, err := u.users.GetByID(ctx, id)
		if err != nil {
			if errors.Is(err, errs.ErrNotFound) {
				continue
			}
			return nil, fmt.Errorf("resolving user %s: %w", id, err)
		}
		users = append(users, user)
	}
	return users, nil
}
