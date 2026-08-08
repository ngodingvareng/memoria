package repository

import (
	"github.com/ngodingvareng/memoria/internal/db"
	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/enum"
)

func toEntityCircleInvite(row db.CircleInvite) *entity.CircleInvite {
	var requiresApproval *bool
	if row.RequiresApproval.Valid {
		requiresApproval = &row.RequiresApproval.Bool
	}

	return &entity.CircleInvite{
		ID:               row.ID,
		CircleID:         row.CircleID,
		InvitedByUserID:  pgUUIDToPtr(row.InvitedByUserID),
		Kind:             enum.CircleInviteKind(row.Kind),
		InviteeUserID:    pgUUIDToPtr(row.InviteeUserID),
		TokenHash:        pgTextToPtr(row.TokenHash),
		RequiresApproval: requiresApproval,
		Status:           enum.CircleInviteStatus(row.Status),
		ExpiresAt:        pgTimestamptzToPtr(row.ExpiresAt),
		RespondedAt:      pgTimestamptzToPtr(row.RespondedAt),
		CreatedAt:        row.CreatedAt.Time,
	}
}

func toEntityCircleInvites(rows []db.CircleInvite) []*entity.CircleInvite {
	out := make([]*entity.CircleInvite, len(rows))
	for i, row := range rows {
		out[i] = toEntityCircleInvite(row)
	}
	return out
}
