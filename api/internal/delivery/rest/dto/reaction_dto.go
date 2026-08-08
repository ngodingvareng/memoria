package dto

import (
	"time"

	"github.com/ngodingvareng/memoria/internal/entity"
)

type SetReactionRequest struct {
	// CircleID omitted means the mention context — see CreateCommentRequest.
	CircleID *string `json:"circle_id,omitempty" validate:"omitempty,uuid"`
	Kind     string  `json:"kind" validate:"required,oneof=heart joy tender" example:"heart"`
}

type RemoveReactionQuery struct {
	CircleID *string `query:"circle_id" validate:"omitempty,uuid"`
}

type ReactionResponse struct {
	ID        string  `json:"id" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	MomentID  string  `json:"moment_id" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	UserID    *string `json:"user_id,omitempty"`
	CircleID  *string `json:"circle_id,omitempty"`
	Kind      string  `json:"kind" example:"heart"`
	CreatedAt string  `json:"created_at" example:"2026-07-20T10:00:00Z"`
}

// ListReactionsResponse deliberately carries only the rows — who
// reacted is visible, how many is never returned as a number
// (FEATURES.md, Response).
type ListReactionsResponse struct {
	Reactions []ReactionResponse `json:"reactions"`
}

func NewReactionResponse(e *entity.Reaction) ReactionResponse {
	var userID *string
	if e.UserID != nil {
		id := e.UserID.String()
		userID = &id
	}
	var circleID *string
	if e.CircleID != nil {
		id := e.CircleID.String()
		circleID = &id
	}

	return ReactionResponse{
		ID:        e.ID.String(),
		MomentID:  e.MomentID.String(),
		UserID:    userID,
		CircleID:  circleID,
		Kind:      string(e.Kind),
		CreatedAt: e.CreatedAt.Format(time.RFC3339),
	}
}
