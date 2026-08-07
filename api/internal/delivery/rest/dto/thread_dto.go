package dto

import (
	"time"

	"github.com/ngodingvareng/memoria/internal/entity"
)

type CreateThreadRequest struct {
	Name                       string  `json:"name" validate:"required,min=1,max=255" example:"Morning workout"`
	Description                *string `json:"description,omitempty" validate:"omitempty,max=2000" example:"Push-ups and a run every weekday morning"`
	HasCommitment              bool    `json:"has_commitment" example:"true"`
	ColorHex                   *string `json:"color_hex,omitempty" validate:"omitempty,hexcolor" example:"#FF5733"`
	ConfirmationTimeoutMinutes *int32  `json:"confirmation_timeout_minutes,omitempty" validate:"omitempty,gt=0" example:"1440"`
}

// UpdateThreadRequest has no has_commitment — see the comment on
// usecase.UpdateThreadInput for why that's a separate endpoint.
type UpdateThreadRequest struct {
	Name                       string  `json:"name" validate:"required,min=1,max=255" example:"Morning workout"`
	Description                *string `json:"description,omitempty" validate:"omitempty,max=2000" example:"Push-ups and a run every weekday morning"`
	ColorHex                   *string `json:"color_hex,omitempty" validate:"omitempty,hexcolor" example:"#FF5733"`
	ConfirmationTimeoutMinutes *int32  `json:"confirmation_timeout_minutes,omitempty" validate:"omitempty,gt=0" example:"1440"`
}

type SearchThreadsQuery struct {
	Name          *string `query:"name" validate:"omitempty,max=255"`
	HasCommitment *bool   `query:"has_commitment"`
	Page          int32   `query:"page" validate:"omitempty,gt=0"`
	PageSize      int32   `query:"page_size" validate:"omitempty,gt=0,lte=100"`
}

type ThreadResponse struct {
	ID                         string  `json:"id" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	UserID                     string  `json:"user_id" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	Name                       string  `json:"name" example:"Morning workout"`
	Description                *string `json:"description,omitempty"`
	HasCommitment              bool    `json:"has_commitment"`
	ColorHex                   *string `json:"color_hex,omitempty"`
	ConfirmationTimeoutMinutes *int32  `json:"confirmation_timeout_minutes,omitempty"`
	CreatedAt                  string  `json:"created_at" example:"2026-07-20T10:00:00Z"`
	UpdatedAt                  string  `json:"updated_at" example:"2026-07-20T10:00:00Z"`
}

type PaginationResponse struct {
	Page     int32 `json:"page" example:"1"`
	PageSize int32 `json:"page_size" example:"20"`
	Total    int64 `json:"total" example:"42"`
}

type SearchThreadsResponse struct {
	Threads    []ThreadResponse   `json:"threads"`
	Pagination PaginationResponse `json:"pagination"`
}

func NewThreadResponse(e *entity.Thread) ThreadResponse {
	return ThreadResponse{
		ID:                         e.ID.String(),
		UserID:                     e.UserID.String(),
		Name:                       e.Name,
		Description:                e.Description,
		HasCommitment:              e.HasCommitment,
		ColorHex:                   e.ColorHex,
		ConfirmationTimeoutMinutes: e.ConfirmationTimeoutMinutes,
		CreatedAt:                  e.CreatedAt.Format(time.RFC3339),
		UpdatedAt:                  e.UpdatedAt.Format(time.RFC3339),
	}

}
