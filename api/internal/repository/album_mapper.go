package repository

import (
	"github.com/ngodingvareng/memoria/internal/db"
	"github.com/ngodingvareng/memoria/internal/entity"
)

func toEntityAlbumImage(row db.ListAlbumImagesRow) *entity.AlbumImage {
	return &entity.AlbumImage{
		ID:                       row.ID,
		MomentID:                 row.MomentID,
		ImagePath:                row.ImagePath,
		ImageAlt:                 pgTextToPtr(row.ImageAlt),
		Width:                    pgInt4ToPtr(row.Width),
		Height:                   pgInt4ToPtr(row.Height),
		OccurredLocal:            row.OccurredLocal.Time,
		OccurredUTCOffsetMinutes: row.OccurredUtcOffsetMinutes,
		ThreadName:               pgTextToPtr(row.ThreadName),
	}
}

func toEntityAlbumImages(rows []db.ListAlbumImagesRow) []*entity.AlbumImage {
	out := make([]*entity.AlbumImage, len(rows))
	for i, row := range rows {
		out[i] = toEntityAlbumImage(row)
	}
	return out
}

func toEntityCircleAlbumImage(row db.ListCircleAlbumImagesRow) *entity.AlbumImage {
	return &entity.AlbumImage{
		ID:                       row.ID,
		MomentID:                 row.MomentID,
		ImagePath:                row.ImagePath,
		ImageAlt:                 pgTextToPtr(row.ImageAlt),
		Width:                    pgInt4ToPtr(row.Width),
		Height:                   pgInt4ToPtr(row.Height),
		OccurredLocal:            row.OccurredLocal.Time,
		OccurredUTCOffsetMinutes: row.OccurredUtcOffsetMinutes,
		IsShared:                 row.IsShared,
		SharedByUserID:           pgUUIDToPtr(row.SharedByUserID),
		SharedByUsername:         pgTextToPtr(row.SharedByUsername),
		ThreadName:               pgTextToPtr(row.ThreadName),
	}
}

func toEntityCircleAlbumImages(rows []db.ListCircleAlbumImagesRow) []*entity.AlbumImage {
	out := make([]*entity.AlbumImage, len(rows))
	for i, row := range rows {
		out[i] = toEntityCircleAlbumImage(row)
	}
	return out
}
