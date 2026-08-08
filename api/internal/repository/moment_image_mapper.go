package repository

import (
	"github.com/ngodingvareng/memoria/internal/db"
	"github.com/ngodingvareng/memoria/internal/entity"
)

func toEntityMomentImage(row db.MomentImage) *entity.MomentImage {
	return &entity.MomentImage{
		ID:               row.ID,
		MomentID:         row.MomentID,
		ImagePath:        row.ImagePath,
		ImageAlt:         pgTextToPtr(row.ImageAlt),
		ContentType:      pgTextToPtr(row.ContentType),
		ByteSize:         pgInt8ToPtr(row.ByteSize),
		Width:            pgInt4ToPtr(row.Width),
		Height:           pgInt4ToPtr(row.Height),
		MetadataStripped: row.MetadataStripped,
		SortOrder:        row.SortOrder,
		CreatedAt:        row.CreatedAt.Time,
	}
}

func toEntityMomentImages(rows []db.MomentImage) []*entity.MomentImage {
	out := make([]*entity.MomentImage, len(rows))
	for i, row := range rows {
		out[i] = toEntityMomentImage(row)
	}
	return out
}
