package dto

import (
	"time"

	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/usecase"
)

type CreateCommentRequest struct {
	// CircleID omitted means the mention context; set, it names which
	// Circle audience the comment is posted in (FEATURES.md, Response).
	CircleID *string `json:"circle_id,omitempty" validate:"omitempty,uuid"`
	Body     string  `json:"body" validate:"required,min=1,max=10000"`
}

type UpdateCommentRequest struct {
	Body string `json:"body" validate:"required,min=1,max=10000"`
}

type CommentResponse struct {
	ID        string  `json:"id" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	MomentID  string  `json:"moment_id" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	UserID    *string `json:"user_id,omitempty"`
	CircleID  *string `json:"circle_id,omitempty"`
	Body      string  `json:"body" example:"Oh man, I remember this!"`
	CreatedAt string  `json:"created_at" example:"2026-07-20T10:00:00Z"`
	UpdatedAt string  `json:"updated_at" example:"2026-07-20T10:00:00Z"`
}

type ListCommentsResponse struct {
	Comments []CommentResponse `json:"comments"`
}

// MomentAudienceResponse is which Response contexts (FEATURES.md) the
// caller may act/read in for a Moment — CircleIDs deliberately carries
// only ids, not names; the frontend already resolves Circle names
// elsewhere (GET /circles/{id}).
type MomentAudienceResponse struct {
	MentionAllowed bool     `json:"mention_allowed"`
	CircleIDs      []string `json:"circle_ids"`
}

func NewMomentAudienceResponse(a *usecase.ResponseAudience) MomentAudienceResponse {
	circleIDs := make([]string, len(a.CircleIDs))
	for i, id := range a.CircleIDs {
		circleIDs[i] = id.String()
	}
	return MomentAudienceResponse{
		MentionAllowed: a.MentionAllowed,
		CircleIDs:      circleIDs,
	}
}

func NewCommentResponse(e *entity.Comment) CommentResponse {
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

	return CommentResponse{
		ID:        e.ID.String(),
		MomentID:  e.MomentID.String(),
		UserID:    userID,
		CircleID:  circleID,
		Body:      e.Body,
		CreatedAt: e.CreatedAt.Format(time.RFC3339),
		UpdatedAt: e.UpdatedAt.Format(time.RFC3339),
	}
}
