package dto

type CheckUsernameAvailabilityResponse struct {
	Available bool `json:"available" example:"true"`
}

type SetUsernameRequest struct {
	// Matches users.chk_users_username_format — lowercase letters,
	// digits, underscore, and dot only, 3-30 characters. See the
	// "username" custom validator registered in UserHandler.
	Username string `json:"username" validate:"required,username" example:"budisantoso"`
}
