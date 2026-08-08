package repository

import (
	"github.com/ngodingvareng/memoria/internal/db"
	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/enum"
)

func toEntityReaction(row db.Reaction) *entity.Reaction {
	return &entity.Reaction{
		ID:        row.ID,
		MomentID:  row.MomentID,
		UserID:    pgUUIDToPtr(row.UserID),
		CircleID:  pgUUIDToPtr(row.CircleID),
		Kind:      enum.ReactionKind(row.Kind),
		CreatedAt: row.CreatedAt.Time,
	}
}

func toEntityReactions(rows []db.Reaction) []*entity.Reaction {
	out := make([]*entity.Reaction, len(rows))
	for i, row := range rows {
		out[i] = toEntityReaction(row)
	}
	return out
}
