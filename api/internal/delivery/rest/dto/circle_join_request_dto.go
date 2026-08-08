package dto

import (
	"time"

	"github.com/ngodingvareng/memoria/internal/entity"
)

type FollowInviteLinkRequest struct {
	Token string `json:"token" validate:"required"`
}

type CircleJoinRequestResponse struct {
	ID              string  `json:"id" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	CircleID        string  `json:"circle_id" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	UserID          string  `json:"user_id" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	InviteID        *string `json:"invite_id,omitempty"`
	Status          string  `json:"status" example:"pending"`
	DecidedByUserID *string `json:"decided_by_user_id,omitempty"`
	DecidedAt       *string `json:"decided_at,omitempty"`
	CreatedAt       string  `json:"created_at" example:"2026-07-20T10:00:00Z"`
}

// FollowInviteLinkResponse reports which of the two link outcomes
// happened — exactly one of Member/JoinRequest is set.
type FollowInviteLinkResponse struct {
	Member      *CircleMemberResponse      `json:"member,omitempty"`
	JoinRequest *CircleJoinRequestResponse `json:"join_request,omitempty"`
}

type ListCircleJoinRequestsResponse struct {
	JoinRequests []CircleJoinRequestResponse `json:"join_requests"`
}

func NewCircleJoinRequestResponse(e *entity.CircleJoinRequest) CircleJoinRequestResponse {
	var inviteID *string
	if e.InviteID != nil {
		id := e.InviteID.String()
		inviteID = &id
	}
	var decidedBy *string
	if e.DecidedByUserID != nil {
		id := e.DecidedByUserID.String()
		decidedBy = &id
	}
	var decidedAt *string
	if e.DecidedAt != nil {
		formatted := e.DecidedAt.Format(time.RFC3339)
		decidedAt = &formatted
	}

	return CircleJoinRequestResponse{
		ID:              e.ID.String(),
		CircleID:        e.CircleID.String(),
		UserID:          e.UserID.String(),
		InviteID:        inviteID,
		Status:          string(e.Status),
		DecidedByUserID: decidedBy,
		DecidedAt:       decidedAt,
		CreatedAt:       e.CreatedAt.Format(time.RFC3339),
	}
}
