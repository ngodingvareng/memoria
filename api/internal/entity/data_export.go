package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/enum"
)

// DataExport is "export all my data" (FEATURES.md, Data Controls): a
// complete, readable archive generated asynchronously, with FilePath
// pointing at object storage once ready.
type DataExport struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	Status       enum.DataExportStatus
	FilePath     *string
	ByteSize     *int64
	ErrorMessage *string
	RequestedAt  time.Time
	CompletedAt  *time.Time
	ExpiresAt    *time.Time
}
