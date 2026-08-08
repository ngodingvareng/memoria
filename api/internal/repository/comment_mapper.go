package repository

import (
	"github.com/ngodingvareng/memoria/internal/db"
	"github.com/ngodingvareng/memoria/internal/entity"
)

func toEntityComment(row db.Comment) *entity.Comment {
	return &entity.Comment{
		ID:        row.ID,
		MomentID:  row.MomentID,
		UserID:    pgUUIDToPtr(row.UserID),
		CircleID:  pgUUIDToPtr(row.CircleID),
		Body:      row.Body,
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
		DeletedAt: pgTimestamptzToPtr(row.DeletedAt),
	}
}

func toEntityComments(rows []db.Comment) []*entity.Comment {
	out := make([]*entity.Comment, len(rows))
	for i, row := range rows {
		out[i] = toEntityComment(row)
	}
	return out
}
