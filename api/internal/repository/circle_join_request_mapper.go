package repository

import (
	"github.com/ngodingvareng/memoria/internal/db"
	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/enum"
)

func toEntityCircleJoinRequest(row db.CircleJoinRequest) *entity.CircleJoinRequest {
	return &entity.CircleJoinRequest{
		ID:              row.ID,
		CircleID:        row.CircleID,
		UserID:          row.UserID,
		InviteID:        pgUUIDToPtr(row.InviteID),
		Status:          enum.CircleJoinRequestStatus(row.Status),
		DecidedByUserID: pgUUIDToPtr(row.DecidedByUserID),
		DecidedAt:       pgTimestamptzToPtr(row.DecidedAt),
		CreatedAt:       row.CreatedAt.Time,
	}
}

func toEntityCircleJoinRequests(rows []db.CircleJoinRequest) []*entity.CircleJoinRequest {
	out := make([]*entity.CircleJoinRequest, len(rows))
	for i, row := range rows {
		out[i] = toEntityCircleJoinRequest(row)
	}
	return out
}
