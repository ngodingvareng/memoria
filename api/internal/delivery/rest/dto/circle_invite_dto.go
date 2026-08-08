package dto

import (
	"time"

	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/usecase"
)

type InviteDirectRequest struct {
	Usernames []string `json:"usernames" validate:"required,min=1,max=50,dive,required,min=3,max=30"`
}

type CreateOrRotateInviteLinkRequest struct {
	RequiresApproval bool `json:"requires_approval"`
}

type SetInviteLinkRequiresApprovalRequest struct {
	RequiresApproval bool `json:"requires_approval"`
}

type CircleInviteResponse struct {
	ID               string  `json:"id" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	CircleID         string  `json:"circle_id" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	InvitedByUserID  *string `json:"invited_by_user_id,omitempty"`
	Kind             string  `json:"kind" example:"username"`
	InviteeUserID    *string `json:"invitee_user_id,omitempty"`
	RequiresApproval *bool   `json:"requires_approval,omitempty"`
	Status           string  `json:"status" example:"pending"`
	ExpiresAt        *string `json:"expires_at,omitempty"`
	RespondedAt      *string `json:"responded_at,omitempty"`
	CreatedAt        string  `json:"created_at" example:"2026-07-20T10:00:00Z"`
}

// CircleInviteLinkResponse is returned only from the create/rotate
// endpoint — it is the one time the raw token is ever exposed.
type CircleInviteLinkResponse struct {
	Invite CircleInviteResponse `json:"invite"`
	Token  string               `json:"token" example:"a1b2c3..."`
}

type InviteDirectResponse struct {
	AddedMembers     []CircleMemberResponse `json:"added_members"`
	PendingInvites   []CircleInviteResponse `json:"pending_invites"`
	SkippedUsernames []string               `json:"skipped_usernames"`
}

type ListCircleInvitesResponse struct {
	Invites []CircleInviteResponse `json:"invites"`
}

func NewCircleInviteResponse(e *entity.CircleInvite) CircleInviteResponse {
	var invitedBy *string
	if e.InvitedByUserID != nil {
		id := e.InvitedByUserID.String()
		invitedBy = &id
	}
	var inviteeID *string
	if e.InviteeUserID != nil {
		id := e.InviteeUserID.String()
		inviteeID = &id
	}
	var expiresAt *string
	if e.ExpiresAt != nil {
		formatted := e.ExpiresAt.Format(time.RFC3339)
		expiresAt = &formatted
	}
	var respondedAt *string
	if e.RespondedAt != nil {
		formatted := e.RespondedAt.Format(time.RFC3339)
		respondedAt = &formatted
	}

	return CircleInviteResponse{
		ID:               e.ID.String(),
		CircleID:         e.CircleID.String(),
		InvitedByUserID:  invitedBy,
		Kind:             string(e.Kind),
		InviteeUserID:    inviteeID,
		RequiresApproval: e.RequiresApproval,
		Status:           string(e.Status),
		ExpiresAt:        expiresAt,
		RespondedAt:      respondedAt,
		CreatedAt:        e.CreatedAt.Format(time.RFC3339),
	}
}

func NewInviteDirectResponse(r *usecase.InviteDirectResult) InviteDirectResponse {
	added := make([]CircleMemberResponse, len(r.AddedMembers))
	for i, m := range r.AddedMembers {
		added[i] = NewCircleMemberResponse(m)
	}
	pending := make([]CircleInviteResponse, len(r.PendingInvites))
	for i, inv := range r.PendingInvites {
		pending[i] = NewCircleInviteResponse(inv)
	}
	skipped := r.SkippedUsernames
	if skipped == nil {
		skipped = []string{}
	}
	return InviteDirectResponse{AddedMembers: added, PendingInvites: pending, SkippedUsernames: skipped}
}
