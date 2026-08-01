package dto

import (
	"time"

	"github.com/ngodingvareng/memoria/internal/entity"
)

type CreateActivityRequest struct {
	Name                       string  `json:"name" validate:"required,min=1,max=255" example:"Olahraga pagi"`
	Description                *string `json:"description,omitempty" validate:"omitempty,max=2000" example:"Push-up dan lari tiap pagi kerja"`
	IsFixedSchedule            bool    `json:"is_fixed_schedule" example:"true"`
	ColorHex                   *string `json:"color_hex,omitempty" validate:"omitempty,hexcolor" example:"#FF5733"`
	ConfirmationTimeoutMinutes *int32  `json:"confirmation_timeout_minutes,omitempty" validate:"omitempty,gt=0" example:"1440"`
}

// UpdateActivityRequest has no is_fixed_schedule — see the comment on
// usecase.UpdateActivityInput for why that's a separate endpoint.
type UpdateActivityRequest struct {
	Name                       string  `json:"name" validate:"required,min=1,max=255" example:"Olahraga pagi"`
	Description                *string `json:"description,omitempty" validate:"omitempty,max=2000" example:"Push-up dan lari tiap pagi kerja"`
	ColorHex                   *string `json:"color_hex,omitempty" validate:"omitempty,hexcolor" example:"#FF5733"`
	ConfirmationTimeoutMinutes *int32  `json:"confirmation_timeout_minutes,omitempty" validate:"omitempty,gt=0" example:"1440"`
}

type SearchActivitiesQuery struct {
	Name            *string `query:"name" validate:"omitempty,max=255"`
	IsFixedSchedule *bool   `query:"is_fixed_schedule"`
	Page            int32   `query:"page" validate:"omitempty,gt=0"`
	PageSize        int32   `query:"page_size" validate:"omitempty,gt=0,lte=100"`
}

type ActivityResponse struct {
	ID                         string  `json:"id" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	UserID                     string  `json:"user_id" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	Name                       string  `json:"name" example:"Olahraga pagi"`
	Description                *string `json:"description,omitempty"`
	IsFixedSchedule            bool    `json:"is_fixed_schedule"`
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

type SearchActivitiesResponse struct {
	Activities []ActivityResponse `json:"activities"`
	Pagination PaginationResponse `json:"pagination"`
}

func NewActivityResponse(e *entity.Activity) ActivityResponse {
	return ActivityResponse{
		ID:                         e.ID.String(),
		UserID:                     e.UserID.String(),
		Name:                       e.Name,
		Description:                e.Description,
		IsFixedSchedule:            e.IsFixedSchedule,
		ColorHex:                   e.ColorHex,
		ConfirmationTimeoutMinutes: e.ConfirmationTimeoutMinutes,
		CreatedAt:                  e.CreatedAt.Format(time.RFC3339),
		UpdatedAt:                  e.UpdatedAt.Format(time.RFC3339),
	}

}
