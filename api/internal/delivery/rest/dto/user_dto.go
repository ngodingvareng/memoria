package dto

import "github.com/ngodingvareng/memoria/internal/entity"

type CheckUsernameAvailabilityResponse struct {
	Available bool `json:"available" example:"true"`
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
	ID   string `json:"id" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	Name string `json:"name" example:"Budi Santoso"`
	// Username is nil until the user has claimed one.
	Username  *string `json:"username,omitempty" example:"budisantoso"`
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
