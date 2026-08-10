package dto

import (
	"time"

	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/usecase"
)

type ListNotificationsQuery struct {
	Page     int32 `query:"page" validate:"omitempty,gt=0"`
	PageSize int32 `query:"page_size" validate:"omitempty,gt=0,lte=100"`
}

// NotificationResponse exposes whichever subject id is non-nil for its
// Kind — see entity.Notification's own field comments for which subject
// goes with which kind. Payload is rendering data only, never
// authoritative.
type NotificationResponse struct {
	ID          string  `json:"id" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	Kind        string  `json:"kind" example:"mentioned_in_moment"`
	ActorUserID *string `json:"actor_user_id,omitempty"`

	MomentID               *string `json:"moment_id,omitempty"`
	ThreadID               *string `json:"thread_id,omitempty"`
	CommitmentID           *string `json:"commitment_id,omitempty"`
	CommitmentOccurrenceID *string `json:"commitment_occurrence_id,omitempty"`
	CircleID               *string `json:"circle_id,omitempty"`
	CircleInviteID         *string `json:"circle_invite_id,omitempty"`
	CircleJoinRequestID    *string `json:"circle_join_request_id,omitempty"`
	RecapID                *string `json:"recap_id,omitempty"`
	ResurfacingID          *string `json:"resurfacing_id,omitempty"`

	Payload   string  `json:"payload" example:"{}"`
	CreatedAt string  `json:"created_at" example:"2026-08-04T14:35:00Z"`
	ReadAt    *string `json:"read_at,omitempty"`
}

type ListNotificationsResponse struct {
	Notifications []NotificationResponse `json:"notifications"`
	Pagination    PaginationResponse     `json:"pagination"`
	UnreadCount   int64                  `json:"unread_count" example:"3"`
}

// NotificationPreferencesResponse mirrors every column on
// notification_preferences — every kind is togglable independently, plus
// the response digest hour and quiet hours (FEATURES.md, Notification
// Preferences).
type NotificationPreferencesResponse struct {
	CommitmentUpcomingEnabled bool `json:"commitment_upcoming_enabled"`
	MomentDueEnabled          bool `json:"moment_due_enabled"`
	MomentMissedEnabled       bool `json:"moment_missed_enabled"`

	EchoReadyEnabled      bool `json:"echo_ready_enabled"`
	RecapGeneratedEnabled bool `json:"recap_generated_enabled"`

	CircleInviteReceivedEnabled      bool `json:"circle_invite_received_enabled"`
	CircleJoinRequestReceivedEnabled bool `json:"circle_join_request_received_enabled"`
	MentionedInMomentEnabled         bool `json:"mentioned_in_moment_enabled"`

	AddedToCircleEnabled             bool `json:"added_to_circle_enabled"`
	CircleJoinRequestApprovedEnabled bool `json:"circle_join_request_approved_enabled"`

	ResponseDigestEnabled bool  `json:"response_digest_enabled"`
	ResponseDigestHour    int16 `json:"response_digest_hour" example:"20"`

	// QuietHoursStart/End are "HH:MM:SS" wall-clock strings, both set or
	// both nil — may wrap past midnight (e.g. 22:00:00 -> 07:00:00).
	QuietHoursStart *string `json:"quiet_hours_start,omitempty" example:"22:00:00"`
	QuietHoursEnd   *string `json:"quiet_hours_end,omitempty" example:"07:00:00"`
}

// UpdateNotificationPreferencesRequest carries every field together —
// preferences are saved as one settings-screen submission, not per-field.
type UpdateNotificationPreferencesRequest struct {
	CommitmentUpcomingEnabled bool `json:"commitment_upcoming_enabled"`
	MomentDueEnabled          bool `json:"moment_due_enabled"`
	MomentMissedEnabled       bool `json:"moment_missed_enabled"`

	EchoReadyEnabled      bool `json:"echo_ready_enabled"`
	RecapGeneratedEnabled bool `json:"recap_generated_enabled"`

	CircleInviteReceivedEnabled      bool `json:"circle_invite_received_enabled"`
	CircleJoinRequestReceivedEnabled bool `json:"circle_join_request_received_enabled"`
	MentionedInMomentEnabled         bool `json:"mentioned_in_moment_enabled"`

	AddedToCircleEnabled             bool `json:"added_to_circle_enabled"`
	CircleJoinRequestApprovedEnabled bool `json:"circle_join_request_approved_enabled"`

	ResponseDigestEnabled bool  `json:"response_digest_enabled"`
	ResponseDigestHour    int16 `json:"response_digest_hour" validate:"gte=0,lte=23" example:"20"`

	QuietHoursStart *string `json:"quiet_hours_start,omitempty" validate:"omitempty,datetime=15:04:05" example:"22:00:00"`
	QuietHoursEnd   *string `json:"quiet_hours_end,omitempty" validate:"omitempty,datetime=15:04:05" example:"07:00:00"`
}

func NewNotificationResponse(n *entity.Notification) NotificationResponse {
	response := NotificationResponse{
		ID:        n.ID.String(),
		Kind:      string(n.Kind),
		Payload:   string(n.Payload),
		CreatedAt: n.CreatedAt.Format(time.RFC3339),
	}
	if n.ActorUserID != nil {
		id := n.ActorUserID.String()
		response.ActorUserID = &id
	}
	if n.MomentID != nil {
		id := n.MomentID.String()
		response.MomentID = &id
	}
	if n.ThreadID != nil {
		id := n.ThreadID.String()
		response.ThreadID = &id
	}
	if n.CommitmentID != nil {
		id := n.CommitmentID.String()
		response.CommitmentID = &id
	}
	if n.CommitmentOccurrenceID != nil {
		id := n.CommitmentOccurrenceID.String()
		response.CommitmentOccurrenceID = &id
	}
	if n.CircleID != nil {
		id := n.CircleID.String()
		response.CircleID = &id
	}
	if n.CircleInviteID != nil {
		id := n.CircleInviteID.String()
		response.CircleInviteID = &id
	}
	if n.CircleJoinRequestID != nil {
		id := n.CircleJoinRequestID.String()
		response.CircleJoinRequestID = &id
	}
	if n.RecapID != nil {
		id := n.RecapID.String()
		response.RecapID = &id
	}
	if n.ResurfacingID != nil {
		id := n.ResurfacingID.String()
		response.ResurfacingID = &id
	}
	if n.ReadAt != nil {
		readAt := n.ReadAt.Format(time.RFC3339)
		response.ReadAt = &readAt
	}
	return response
}

func NewListNotificationsResponse(r *usecase.NotificationListResult) ListNotificationsResponse {
	notifications := make([]NotificationResponse, len(r.Notifications))
	for i, n := range r.Notifications {
		notifications[i] = NewNotificationResponse(n)
	}
	return ListNotificationsResponse{
		Notifications: notifications,
		Pagination:    PaginationResponse{Page: r.Page, PageSize: r.PageSize, Total: r.Total},
		UnreadCount:   r.UnreadCount,
	}
}

func NewNotificationPreferencesResponse(p *entity.NotificationPreference) NotificationPreferencesResponse {
	response := NotificationPreferencesResponse{
		CommitmentUpcomingEnabled:        p.CommitmentUpcomingEnabled,
		MomentDueEnabled:                 p.MomentDueEnabled,
		MomentMissedEnabled:              p.MomentMissedEnabled,
		EchoReadyEnabled:                 p.EchoReadyEnabled,
		RecapGeneratedEnabled:            p.RecapGeneratedEnabled,
		CircleInviteReceivedEnabled:      p.CircleInviteReceivedEnabled,
		CircleJoinRequestReceivedEnabled: p.CircleJoinRequestReceivedEnabled,
		MentionedInMomentEnabled:         p.MentionedInMomentEnabled,
		AddedToCircleEnabled:             p.AddedToCircleEnabled,
		CircleJoinRequestApprovedEnabled: p.CircleJoinRequestApprovedEnabled,
		ResponseDigestEnabled:            p.ResponseDigestEnabled,
		ResponseDigestHour:               p.ResponseDigestHour,
	}
	if p.QuietHoursStart != nil {
		start := p.QuietHoursStart.Format("15:04:05")
		response.QuietHoursStart = &start
	}
	if p.QuietHoursEnd != nil {
		end := p.QuietHoursEnd.Format("15:04:05")
		response.QuietHoursEnd = &end
	}
	return response
}
