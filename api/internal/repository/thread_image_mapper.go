package repository

import (
	"github.com/ngodingvareng/memoria/internal/db"
	"github.com/ngodingvareng/memoria/internal/entity"
)

func toEntityThreadImage(row db.ThreadImage) *entity.ThreadImage {
	return &entity.ThreadImage{
		ID:        row.ID,
		ThreadID:  row.ThreadID,
		ImagePath: row.ImagePath,
		ImageAlt:  pgTextToPtr(row.ImageAlt),
		SortOrder: row.SortOrder,
		CreatedAt: row.CreatedAt.Time,
	}
}

func toEntityThreadImages(rows []db.ThreadImage) []*entity.ThreadImage {
	out := make([]*entity.ThreadImage, len(rows))
	for i, row := range rows {
		out[i] = toEntityThreadImage(row)
	}
	return out
}
