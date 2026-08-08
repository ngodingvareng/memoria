package repository

import (
	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/db"
	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/enum"
)

func toEntityCircle(row db.Circle) *entity.Circle {
	return &entity.Circle{
		ID:              row.ID,
		Name:            row.Name,
		Description:     pgTextToPtr(row.Description),
		ColorHex:        pgTextToPtr(row.ColorHex),
		ImagePath:       pgTextToPtr(row.ImagePath),
		CreatedByUserID: pgUUIDToPtr(row.CreatedByUserID),
		CreatedAt:       row.CreatedAt.Time,
		UpdatedAt:       row.UpdatedAt.Time,
		DissolvedAt:     pgTimestamptzToPtr(row.DissolvedAt),
	}
}

func toEntityCircles(rows []db.Circle) []*entity.Circle {
	out := make([]*entity.Circle, len(rows))
	for i, row := range rows {
		out[i] = toEntityCircle(row)
	}
	return out
}

func toCreateCircleParams(c *entity.Circle) db.CreateCircleParams {
	return db.CreateCircleParams{
		Name:        c.Name,
		Description: ptrToPgText(c.Description),
		ColorHex:    ptrToPgText(c.ColorHex),
		ImagePath:   ptrToPgText(c.ImagePath),
	}
}

func toUpdateCircleParams(c *entity.Circle, userID uuid.UUID) db.UpdateCircleParams {
	return db.UpdateCircleParams{
		ID:          c.ID,
		UserID:      userID,
		Name:        c.Name,
		Description: ptrToPgText(c.Description),
		ColorHex:    ptrToPgText(c.ColorHex),
		ImagePath:   ptrToPgText(c.ImagePath),
	}
}

func toEntityCircleMember(row db.CircleMember) *entity.CircleMember {
	return &entity.CircleMember{
		CircleID:   row.CircleID,
		UserID:     row.UserID,
		Role:       enum.CircleRole(row.Role),
		CanInvite:  row.CanInvite,
		CanCapture: row.CanCapture,
		JoinedAt:   row.JoinedAt.Time,
		LeftAt:     pgTimestamptzToPtr(row.LeftAt),
	}
}

func toEntityCircleMembers(rows []db.CircleMember) []*entity.CircleMember {
	out := make([]*entity.CircleMember, len(rows))
	for i, row := range rows {
		out[i] = toEntityCircleMember(row)
	}
	return out
}
