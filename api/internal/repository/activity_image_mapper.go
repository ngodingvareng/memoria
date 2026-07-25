package repository

import (
	"github.com/ngodingvareng/memoria/internal/db"
	"github.com/ngodingvareng/memoria/internal/entity"
)

func toEntityActivityImage(row db.ActivityImage) *entity.ActivityImage {
	return &entity.ActivityImage{
		ID:         row.ID,
		ActivityID: row.ActivityID,
		ImagePath:  row.ImagePath,
		ImageAlt:   pgTextToPtr(row.ImageAlt),
		CreatedAt:  row.CreatedAt.Time,
	}
}

func toEntityActivityImages(rows []db.ActivityImage) []*entity.ActivityImage {
	out := make([]*entity.ActivityImage, len(rows))
	for i, row := range rows {
		out[i] = toEntityActivityImage(row)
	}
	return out
}
