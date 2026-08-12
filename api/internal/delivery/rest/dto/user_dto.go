package dto

import (
	"time"

	"github.com/ngodingvareng/memoria/internal/entity"
)

type CheckUsernameAvailabilityResponse struct {
	Available bool `json:"available" example:"true"`
}

// PrivateUserResponse is the caller's own full profile — the privacy
// fields PublicUserResponse deliberately omits (Social Interaction +
// Data Controls, FEATURES.md Privacy & Control). Backs GET /users/me and
// settings-screen seed values.
type PrivateUserResponse struct {
	ID        string  `json:"id"                   example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	Name      string  `json:"name"                 example:"Budi Santoso"`
	Username  *string `json:"username,omitempty"   example:"budisantoso"`
	Email     string  `json:"email"                example:"budi@example.com"`
	ImagePath *string `json:"image_path,omitempty"`
	Bio       *string `json:"bio,omitempty"`

	MentionPolicy          string `json:"mention_policy"           example:"anyone"`
	CircleInvitePolicy     string `json:"circle_invite_policy"     example:"known"`
	DiscoverableByUsername bool   `json:"discoverable_by_username" example:"true"`
	StripPhotoMetadata     bool   `json:"strip_photo_metadata"     example:"false"`

	CreatedAt string `json:"created_at" example:"2026-07-20T10:00:00Z"`
}

func NewPrivateUserResponse(u *entity.User) PrivateUserResponse {
	return PrivateUserResponse{
		ID:                     u.ID.String(),
		Name:                   u.Name,
		Username:               u.Username,
		Email:                  u.Email,
		ImagePath:              u.ImagePath,
		Bio:                    u.Bio,
		MentionPolicy:          string(u.MentionPolicy),
		CircleInvitePolicy:     string(u.CircleInvitePolicy),
		DiscoverableByUsername: u.DiscoverableByUsername,
		StripPhotoMetadata:     u.StripPhotoMetadata,
		CreatedAt:              u.CreatedAt.Format(time.RFC3339),
	}
}

// UpdatePrivacySettingsRequest carries every Social Interaction + Data
// Controls toggle together — saved as one settings-screen submission.
type UpdatePrivacySettingsRequest struct {
	MentionPolicy          string `json:"mention_policy"           validate:"required,oneof=anyone known nobody" example:"known"`
	CircleInvitePolicy     string `json:"circle_invite_policy"     validate:"required,oneof=anyone known nobody" example:"known"`
	DiscoverableByUsername bool   `json:"discoverable_by_username"                                               example:"true"`
	StripPhotoMetadata     bool   `json:"strip_photo_metadata"                                                   example:"false"`
}

// BlockUserRequest / MuteUserRequest / MarkUserKnownRequest resolve the
// target the same way — by username.
type BlockUserRequest struct {
	Username string `json:"username" validate:"required,username" example:"budisantoso"`
}

type MuteUserRequest struct {
	Username string `json:"username" validate:"required,username" example:"budisantoso"`
}

type MarkUserKnownRequest struct {
	Username string `json:"username" validate:"required,username" example:"budisantoso"`
}

// DeleteAccountRequest requires the caller to type the literal word
// "DELETE" — a lightweight, auth-method-agnostic confirmation that
// works identically whether the account has a password or is
// Google-OAuth-only (FEATURES.md, Data Controls: "Delete account").
type DeleteAccountRequest struct {
	Confirmation string `json:"confirmation" validate:"required,eq=DELETE" example:"DELETE"`
}

type ListBlockedUsersResponse struct {
	Users []PublicUserResponse `json:"users"`
}

type ListMutedUsersResponse struct {
	Users []PublicUserResponse `json:"users"`
}

type ListKnownUsersResponse struct {
	Users []PublicUserResponse `json:"users"`
}

func NewListBlockedUsersResponse(users []*entity.User) ListBlockedUsersResponse {
	response := ListBlockedUsersResponse{Users: make([]PublicUserResponse, len(users))}
	for i, u := range users {
		response.Users[i] = NewPublicUserResponse(u)
	}
	return response
}

func NewListMutedUsersResponse(users []*entity.User) ListMutedUsersResponse {
	response := ListMutedUsersResponse{Users: make([]PublicUserResponse, len(users))}
	for i, u := range users {
		response.Users[i] = NewPublicUserResponse(u)
	}
	return response
}

func NewListKnownUsersResponse(users []*entity.User) ListKnownUsersResponse {
	response := ListKnownUsersResponse{Users: make([]PublicUserResponse, len(users))}
	for i, u := range users {
		response.Users[i] = NewPublicUserResponse(u)
	}
	return response
}

type SetUsernameRequest struct {
	// Matches users.chk_users_username_format — lowercase letters,
	// digits, underscore, and dot only, 3-30 characters. See the
	// "username" custom validator registered in UserHandler.
	Username string `json:"username" validate:"required,username" example:"budisantoso"`
}

// PublicUserResponse is the minimal, other-facing view of a user —
// what a Circle member list, join request, or invite is allowed to
// show about someone who isn't the caller. Deliberately excludes
// Email/EmailVerified/policy fields on UserResponse.
type PublicUserResponse struct {
	ID   string `json:"id"   example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	Name string `json:"name" example:"Budi Santoso"`
	// Username is nil until the user has claimed one.
	Username  *string `json:"username,omitempty"   example:"budisantoso"`
	ImagePath *string `json:"image_path,omitempty"`
	Bio       *string `json:"bio,omitempty"`
}

func NewPublicUserResponse(u *entity.User) PublicUserResponse {
	return PublicUserResponse{
		ID:        u.ID.String(),
		Name:      u.Name,
		Username:  u.Username,
		ImagePath: u.ImagePath,
		Bio:       u.Bio,
	}
}
