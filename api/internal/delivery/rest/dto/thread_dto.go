package dto

import (
	"time"

	"github.com/ngodingvareng/memoria/internal/entity"
)

type CreateThreadRequest struct {
	Name        string  `json:"name"                  validate:"required,min=1,max=255" example:"Morning workout"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=2000"     example:"Push-ups and a run every weekday morning"`
	ColorHex    *string `json:"color_hex,omitempty"   validate:"omitempty,hexcolor"     example:"#FF5733"`
	// CircleID creates a collaborative Thread owned by that Circle
	// instead of a personal one; the caller must be an active member.
	CircleID *string `json:"circle_id,omitempty" validate:"omitempty,uuid" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
}

type UpdateThreadRequest struct {
	Name        string  `json:"name"                  validate:"required,min=1,max=255" example:"Morning workout"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=2000"     example:"Push-ups and a run every weekday morning"`
	ColorHex    *string `json:"color_hex,omitempty"   validate:"omitempty,hexcolor"     example:"#FF5733"`
}

type SearchThreadsQuery struct {
	Name     *string `query:"name"     validate:"omitempty,max=255"`
	Archived *bool   `query:"archived"`
	// Sort is optional; the only supported value switches the default
	// sort_order/created_at ordering to updated_at DESC — backs the
	// moment-creation Thread picker's "most recently updated" default
	// suggestions.
	Sort     *string `query:"sort"      validate:"omitempty,oneof=updated_at"`
	Page     int32   `query:"page"      validate:"omitempty,gt=0"`
	PageSize int32   `query:"page_size" validate:"omitempty,gt=0,lte=100"`
}

type ThreadResponse struct {
	ID          string  `json:"id"                    example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	UserID      string  `json:"user_id"               example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	CircleID    *string `json:"circle_id,omitempty"   example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	Name        string  `json:"name"                  example:"Morning workout"`
	Description *string `json:"description,omitempty"`
	ColorHex    *string `json:"color_hex,omitempty"`
	SortOrder   int32   `json:"sort_order"            example:"0"`
	ArchivedAt  *string `json:"archived_at,omitempty" example:"2026-07-20T10:00:00Z"`
	CreatedAt   string  `json:"created_at"            example:"2026-07-20T10:00:00Z"`
	UpdatedAt   string  `json:"updated_at"            example:"2026-07-20T10:00:00Z"`
}

type PaginationResponse struct {
	Page     int32 `json:"page"      example:"1"`
	PageSize int32 `json:"page_size" example:"20"`
	Total    int64 `json:"total"     example:"42"`
}

type SearchThreadsResponse struct {
	Threads    []ThreadResponse   `json:"threads"`
	Pagination PaginationResponse `json:"pagination"`
}

type ListThreadsResponse struct {
	Threads []ThreadResponse `json:"threads"`
}

func NewThreadResponse(e *entity.Thread) ThreadResponse {
	var circleID *string
	if e.CircleID != nil {
		id := e.CircleID.String()
		circleID = &id
	}

	var archivedAt *string
	if e.ArchivedAt != nil {
		formatted := e.ArchivedAt.Format(time.RFC3339)
		archivedAt = &formatted
	}

	return ThreadResponse{
		ID:          e.ID.String(),
		UserID:      e.UserID.String(),
		CircleID:    circleID,
		Name:        e.Name,
		Description: e.Description,
		ColorHex:    e.ColorHex,
		SortOrder:   e.SortOrder,
		ArchivedAt:  archivedAt,
		CreatedAt:   e.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   e.UpdatedAt.Format(time.RFC3339),
	}
}
