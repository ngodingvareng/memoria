package usecase

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"

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
	// MarkUserKnown records that knowerUserID knows username (FEATURES.md,
	// Privacy & Control's "known" audience tier) — one-directional,
	// silent, and idempotent (see UserKnownRepository). Returns
	// errs.ErrNotFound if username doesn't resolve to a user.
	MarkUserKnown(ctx context.Context, knowerUserID uuid.UUID, username string) error
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

func (u *userUsecase) CheckUsernameAvailability(ctx context.Context, username string) (bool, error) {
	_, err := u.users.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return true, nil
		}
		return false, fmt.Errorf("checking username availability: %w", err)
	}
	return false, nil
}

func (u *userUsecase) SetUsername(ctx context.Context, userID uuid.UUID, username string) (*entity.User, error) {
	user, err := u.users.SetUsername(ctx, userID, username)
	if err != nil {
		return nil, fmt.Errorf("setting username: %w", err)
	}
	return user, nil
}

func (u *userUsecase) GetPublicProfile(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	user, err := u.users.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("getting public profile: %w", err)
	}
	return user, nil
}

func (u *userUsecase) GetPublicProfileByUsername(ctx context.Context, username string) (*entity.User, error) {
	user, err := u.users.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("getting public profile by username: %w", err)
	}
	return user, nil
}

func (u *userUsecase) MarkUserKnown(ctx context.Context, knowerUserID uuid.UUID, username string) error {
	target, err := u.users.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return errs.ErrNotFound
		}
		return fmt.Errorf("resolving username: %w", err)
	}
	if err := u.knowns.MarkKnown(ctx, knowerUserID, target.ID); err != nil {
		return fmt.Errorf("marking user known: %w", err)
	}
	return nil
}

func (u *userUsecase) UploadProfileImage(ctx context.Context, input UploadProfileImageInput) (*entity.User, error) {
	current, err := u.users.GetByID(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("getting current user: %w", err)
	}

	key := buildPublicImageKey("users", input.UserID, input.FileName)
	if err := u.storage.Put(ctx, key, input.Body, input.Size, input.ContentType); err != nil {
		return nil, fmt.Errorf("uploading profile image %s: %w", key, err)
	}

	publicURL := u.storage.PublicURL(key)
	updated, err := u.users.UpdateImagePath(ctx, input.UserID, &publicURL)
	if err != nil {
		// Mirrors thread_image_usecase.go's orphan-cleanup: the upload
		// already succeeded before this DB write failed, so best-effort
		// delete it rather than leave it dangling in storage.
		if delErr := u.storage.Delete(context.WithoutCancel(ctx), key); delErr != nil {
			return nil, fmt.Errorf("saving profile image: %w (cleanup also failed: %v)", err, delErr)
		}
		return nil, fmt.Errorf("saving profile image: %w", err)
	}

	if current.ImagePath != nil {
		if oldKey, ok := strings.CutPrefix(*current.ImagePath, u.storage.PublicURL("")); ok {
			_ = u.storage.Delete(context.WithoutCancel(ctx), oldKey)
		}
	}

	return updated, nil
}

// GetOwnProfile implements [UserUsecase].
func (u *userUsecase) GetOwnProfile(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	user, err := u.users.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting own profile: %w", err)
	}
	return user, nil
}

// UpdatePrivacySettings implements [UserUsecase].
func (u *userUsecase) UpdatePrivacySettings(ctx context.Context, input UpdatePrivacySettingsInput) (*entity.User, error) {
	updated, err := u.users.UpdatePrivacySettings(ctx, &entity.User{
		ID:                     input.UserID,
		MentionPolicy:          input.MentionPolicy,
		CircleInvitePolicy:     input.CircleInvitePolicy,
		DiscoverableByUsername: input.DiscoverableByUsername,
		StripPhotoMetadata:     input.StripPhotoMetadata,
	})
	if err != nil {
		return nil, fmt.Errorf("updating privacy settings: %w", err)
	}
	return updated, nil
}

// BlockUser implements [UserUsecase].
func (u *userUsecase) BlockUser(ctx context.Context, blockerUserID uuid.UUID, username string) error {
	target, err := u.users.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return errs.ErrNotFound
		}
		return fmt.Errorf("resolving username to block: %w", err)
	}
	if err := u.blocks.Block(ctx, blockerUserID, target.ID); err != nil {
		return fmt.Errorf("blocking user: %w", err)
	}
	return nil
}

// UnblockUser implements [UserUsecase].
func (u *userUsecase) UnblockUser(ctx context.Context, blockerUserID uuid.UUID, username string) error {
	target, err := u.users.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return errs.ErrNotFound
		}
		return fmt.Errorf("resolving username to unblock: %w", err)
	}
	if err := u.blocks.Unblock(ctx, blockerUserID, target.ID); err != nil {
		return fmt.Errorf("unblocking user: %w", err)
	}
	return nil
}

// ListBlockedUsers implements [UserUsecase].
func (u *userUsecase) ListBlockedUsers(ctx context.Context, blockerUserID uuid.UUID) ([]*entity.User, error) {
	ids, err := u.blocks.ListBlockedUserIDs(ctx, blockerUserID)
	if err != nil {
		return nil, fmt.Errorf("listing blocked users: %w", err)
	}
	users, err := u.resolveUsers(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("resolving blocked users: %w", err)
	}
	return users, nil
}

// MuteUser implements [UserUsecase].
func (u *userUsecase) MuteUser(ctx context.Context, muterUserID uuid.UUID, username string) error {
	target, err := u.users.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return errs.ErrNotFound
		}
		return fmt.Errorf("resolving username to mute: %w", err)
	}
	if err := u.mutes.Mute(ctx, muterUserID, target.ID); err != nil {
		return fmt.Errorf("muting user: %w", err)
	}
	return nil
}

// UnmuteUser implements [UserUsecase].
func (u *userUsecase) UnmuteUser(ctx context.Context, muterUserID uuid.UUID, username string) error {
	target, err := u.users.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return errs.ErrNotFound
		}
		return fmt.Errorf("resolving username to unmute: %w", err)
	}
	if err := u.mutes.Unmute(ctx, muterUserID, target.ID); err != nil {
		return fmt.Errorf("unmuting user: %w", err)
	}
	return nil
}

// ListMutedUsers implements [UserUsecase].
func (u *userUsecase) ListMutedUsers(ctx context.Context, muterUserID uuid.UUID) ([]*entity.User, error) {
	ids, err := u.mutes.ListMutedUserIDs(ctx, muterUserID)
	if err != nil {
		return nil, fmt.Errorf("listing muted users: %w", err)
	}
	users, err := u.resolveUsers(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("resolving muted users: %w", err)
	}
	return users, nil
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
