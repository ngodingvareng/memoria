package repository

import (
	"github.com/ngodingvareng/memoria/internal/db"
	"github.com/ngodingvareng/memoria/internal/entity"
)

func toEntityMomentMention(row db.MomentMention) *entity.MomentMention {
	return &entity.MomentMention{
		ID:              row.ID,
		MomentID:        row.MomentID,
		MentionedUserID: pgUUIDToPtr(row.MentionedUserID),
		DisplayName:     row.DisplayName,
		RemovedAt:       pgTimestamptzToPtr(row.RemovedAt),
		CreatedAt:       row.CreatedAt.Time,
	}
}

func toEntityMomentMentions(rows []db.MomentMention) []*entity.MomentMention {
	out := make([]*entity.MomentMention, len(rows))
	for i, row := range rows {
		out[i] = toEntityMomentMention(row)
	}
	return out
}

func toEntityMomentCircle(row db.MomentCircle) *entity.MomentCircle {
	return &entity.MomentCircle{
		MomentID:       row.MomentID,
		CircleID:       row.CircleID,
		SharedByUserID: pgUUIDToPtr(row.SharedByUserID),
		SharedAt:       row.SharedAt.Time,
	}
}
