package dto

import (
	"time"

	"github.com/ngodingvareng/memoria/internal/entity"
)

type CreateCircleRequest struct {
	Name        string  `json:"name"                  validate:"required,min=1,max=255" example:"NgodingVareng"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=2000"`
	ColorHex    *string `json:"color_hex,omitempty"   validate:"omitempty,hexcolor"`
	ImagePath   *string `json:"image_path,omitempty"  validate:"omitempty,max=1024"`
}

type UpdateCircleRequest struct {
	Name        string  `json:"name"                  validate:"required,min=1,max=255" example:"NgodingVareng"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=2000"`
	ColorHex    *string `json:"color_hex,omitempty"   validate:"omitempty,hexcolor"`
	ImagePath   *string `json:"image_path,omitempty"  validate:"omitempty,max=1024"`
}

type UpdateCircleMemberRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=admin member" example:"admin"`
}

type UpdateCircleMemberPermissionsRequest struct {
	CanInvite  bool `json:"can_invite"`
	CanCapture bool `json:"can_capture"`
}

type CircleResponse struct {
	ID              string  `json:"id"                           example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	Name            string  `json:"name"                         example:"NgodingVareng"`
	Description     *string `json:"description,omitempty"`
	ColorHex        *string `json:"color_hex,omitempty"`
	ImagePath       *string `json:"image_path,omitempty"`
	CreatedByUserID *string `json:"created_by_user_id,omitempty"`
	DissolvedAt     *string `json:"dissolved_at,omitempty"`
	CreatedAt       string  `json:"created_at"                   example:"2026-07-20T10:00:00Z"`
	UpdatedAt       string  `json:"updated_at"                   example:"2026-07-20T10:00:00Z"`
}

type ListCirclesResponse struct {
	Circles []CircleResponse `json:"circles"`
}

type CircleMemberResponse struct {
	CircleID   string `json:"circle_id"   example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	UserID     string `json:"user_id"     example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	Role       string `json:"role"        example:"admin"`
	CanInvite  bool   `json:"can_invite"`
	CanCapture bool   `json:"can_capture"`
	JoinedAt   string `json:"joined_at"   example:"2026-07-20T10:00:00Z"`
}

type ListCircleMembersResponse struct {
	Members []CircleMemberResponse `json:"members"`
}

func NewCircleResponse(e *entity.Circle) CircleResponse {
	var createdBy *string
	if e.CreatedByUserID != nil {
		id := e.CreatedByUserID.String()
		createdBy = &id
	}

	var dissolvedAt *string
	if e.DissolvedAt != nil {
		formatted := e.DissolvedAt.Format(time.RFC3339)
		dissolvedAt = &formatted
	}

	return CircleResponse{
		ID:              e.ID.String(),
		Name:            e.Name,
		Description:     e.Description,
		ColorHex:        e.ColorHex,
		ImagePath:       e.ImagePath,
		CreatedByUserID: createdBy,
		DissolvedAt:     dissolvedAt,
		CreatedAt:       e.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       e.UpdatedAt.Format(time.RFC3339),
	}
}

func NewCircleMemberResponse(e *entity.CircleMember) CircleMemberResponse {
	return CircleMemberResponse{
		CircleID:   e.CircleID.String(),
		UserID:     e.UserID.String(),
		Role:       string(e.Role),
		CanInvite:  e.CanInvite,
		CanCapture: e.CanCapture,
		JoinedAt:   e.JoinedAt.Format(time.RFC3339),
	}
}
