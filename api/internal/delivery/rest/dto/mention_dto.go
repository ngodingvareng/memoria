package dto

import (
	"time"

	"github.com/ngodingvareng/memoria/internal/entity"
)

type CreateMentionRequest struct {
	Username string `json:"username" validate:"required,min=3,max=30"`
}

type MentionResponse struct {
	ID              string  `json:"id" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	MomentID        string  `json:"moment_id" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	MentionedUserID *string `json:"mentioned_user_id,omitempty"`
	DisplayName     string  `json:"display_name" example:"Gede"`
	CreatedAt       string  `json:"created_at" example:"2026-07-20T10:00:00Z"`
}

type ListMentionsResponse struct {
	Mentions []MentionResponse `json:"mentions"`
}

type ShareMomentToCircleRequest struct {
	CircleID string `json:"circle_id" validate:"required,uuid"`
}

type MomentCircleResponse struct {
	MomentID       string  `json:"moment_id" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	CircleID       string  `json:"circle_id" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	SharedByUserID *string `json:"shared_by_user_id,omitempty"`
	SharedAt       string  `json:"shared_at" example:"2026-07-20T10:00:00Z"`
}

func NewMentionResponse(e *entity.MomentMention) MentionResponse {
	var mentionedUserID *string
	if e.MentionedUserID != nil {
		id := e.MentionedUserID.String()
		mentionedUserID = &id
	}

	return MentionResponse{
		ID:              e.ID.String(),
		MomentID:        e.MomentID.String(),
		MentionedUserID: mentionedUserID,
		DisplayName:     e.DisplayName,
		CreatedAt:       e.CreatedAt.Format(time.RFC3339),
	}
}

func NewMomentCircleResponse(e *entity.MomentCircle) MomentCircleResponse {
	var sharedBy *string
	if e.SharedByUserID != nil {
		id := e.SharedByUserID.String()
		sharedBy = &id
	}

	return MomentCircleResponse{
		MomentID:       e.MomentID.String(),
		CircleID:       e.CircleID.String(),
		SharedByUserID: sharedBy,
		SharedAt:       e.SharedAt.Format(time.RFC3339),
	}
}
