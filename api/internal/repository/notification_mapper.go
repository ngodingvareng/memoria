package repository

import (
	"github.com/ngodingvareng/memoria/internal/db"
	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/enum"
)

func toEntityNotification(row db.Notification) *entity.Notification {
	return &entity.Notification{
		ID:                     row.ID,
		UserID:                 row.UserID,
		Kind:                   enum.NotificationKind(row.Kind),
		ActorUserID:            pgUUIDToPtr(row.ActorUserID),
		MomentID:               pgUUIDToPtr(row.MomentID),
		ThreadID:               pgUUIDToPtr(row.ThreadID),
		CommitmentID:           pgUUIDToPtr(row.CommitmentID),
		CommitmentOccurrenceID: pgUUIDToPtr(row.CommitmentOccurrenceID),
		CircleID:               pgUUIDToPtr(row.CircleID),
		CircleInviteID:         pgUUIDToPtr(row.CircleInviteID),
		CircleJoinRequestID:    pgUUIDToPtr(row.CircleJoinRequestID),
		RecapID:                pgUUIDToPtr(row.RecapID),
		ResurfacingID:          pgUUIDToPtr(row.ResurfacingID),
		Payload:                row.Payload,
		DeliverAfter:           row.DeliverAfter.Time,
		DeliveredAt:            pgTimestamptzToPtr(row.DeliveredAt),
		ReadAt:                 pgTimestamptzToPtr(row.ReadAt),
		CreatedAt:              row.CreatedAt.Time,
	}
}

func toEntityNotificationPreference(row db.NotificationPreference) *entity.NotificationPreference {
	return &entity.NotificationPreference{
		UserID:                           row.UserID,
		CommitmentUpcomingEnabled:        row.CommitmentUpcomingEnabled,
		MomentDueEnabled:                 row.MomentDueEnabled,
		MomentMissedEnabled:              row.MomentMissedEnabled,
		EchoReadyEnabled:                 row.EchoReadyEnabled,
		RecapGeneratedEnabled:            row.RecapGeneratedEnabled,
		CircleInviteReceivedEnabled:      row.CircleInviteReceivedEnabled,
		CircleJoinRequestReceivedEnabled: row.CircleJoinRequestReceivedEnabled,
		MentionedInMomentEnabled:         row.MentionedInMomentEnabled,
		AddedToCircleEnabled:             row.AddedToCircleEnabled,
		CircleJoinRequestApprovedEnabled: row.CircleJoinRequestApprovedEnabled,
		ResponseDigestEnabled:            row.ResponseDigestEnabled,
		ResponseDigestHour:               row.ResponseDigestHour,
		QuietHoursStart:                  pgTimeToPtr(row.QuietHoursStart),
		QuietHoursEnd:                    pgTimeToPtr(row.QuietHoursEnd),
		CreatedAt:                        row.CreatedAt.Time,
		UpdatedAt:                        row.UpdatedAt.Time,
	}
}
