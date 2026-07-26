package dto

import (
	"time"

	"github.com/ngodingvareng/memoria/internal/entity"
)

type RegisterRequest struct {
	Name  string `json:"name" validate:"required,min=1,max=255" example:"Budi Santoso"`
	Email string `json:"email" validate:"required,email,max=255" example:"budi@example.com"`
	// max=72: bcrypt silently truncates anything past 72 bytes rather
	// than erroring — validating here avoids a confusing situation where
	// a very long password "works" but only its first 72 bytes actually
	// matter for the hash.
	Password string `json:"password" validate:"required,min=8,max=72" example:"correct horse battery staple"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email" example:"budi@example.com"`
	Password string `json:"password" validate:"required" example:"correct horse battery staple"`
}

type UserResponse struct {
	ID            string `json:"id" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	Name          string `json:"name" example:"Budi Santoso"`
	Email         string `json:"email" example:"budi@example.com"`
	EmailVerified bool   `json:"email_verified" example:"false"`
	CreatedAt     string `json:"created_at" example:"2026-07-20T10:00:00Z"`
}

func NewUserResponse(u *entity.User) UserResponse {
	return UserResponse{
		ID:            u.ID.String(),
		Name:          u.Name,
		Email:         u.Email,
		EmailVerified: u.EmailVerified,
		CreatedAt:     u.CreatedAt.Format(time.RFC3339),
	}
}
