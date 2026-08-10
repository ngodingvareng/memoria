package repository

import (
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// This file holds pgtype <-> plain-Go-pointer conversions shared by
// every entity mapper in this package (thread_mapper.go,
// user_mapper.go, etc. as they're added). Keeping pgtype awareness
// confined to internal/repository is deliberate — entity and usecase
// never see pgtype directly, only plain Go types (string, *string,
// int32, time.Time, ...).

func pgTextToPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return new(t.String)
}

func ptrToPgText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func pgInt4ToPtr(i pgtype.Int4) *int32 {
	if !i.Valid {
		return nil
	}
	return new(i.Int32)
}

func ptrToPgInt4(i *int32) pgtype.Int4 {
	if i == nil {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: *i, Valid: true}
}

func pgTimestamptzToPtr(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return new(t.Time)
}

func timeToPgTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func ptrToPgBool(b *bool) pgtype.Bool {
	if b == nil {
		return pgtype.Bool{}
	}
	return pgtype.Bool{Bool: *b, Valid: true}
}

func pgUUIDToPtr(u pgtype.UUID) *uuid.UUID {
	if !u.Valid {
		return nil
	}
	id := uuid.UUID(u.Bytes)
	return &id
}

func ptrToPgUUID(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: *id, Valid: true}
}

func pgInt8ToPtr(i pgtype.Int8) *int64 {
	if !i.Valid {
		return nil
	}
	return new(i.Int64)
}

func ptrToPgInt8(i *int64) pgtype.Int8 {
	if i == nil {
		return pgtype.Int8{}
	}
	return pgtype.Int8{Int64: *i, Valid: true}
}

// pgNumericToPtr converts a NUMERIC column (moments.latitude/longitude)
// into a plain *float64. The error Float64Value can return only occurs
// on malformed digit data, which can't happen for a value Postgres has
// already round-tripped through a column the CHECK constraints on
// moments already bound — so it's discarded here rather than threaded
// through every mapper caller.
func pgNumericToPtr(n pgtype.Numeric) *float64 {
	if !n.Valid {
		return nil
	}
	f, _ := n.Float64Value()
	return &f.Float64
}

// ptrToPgNumeric is the inverse of pgNumericToPtr. It goes through
// Scan(string) because pgtype.Numeric has no direct float64 setter; the
// Scan error is discarded because strconv.FormatFloat always produces
// text Scan can parse.
func ptrToPgNumeric(f *float64) pgtype.Numeric {
	if f == nil {
		return pgtype.Numeric{}
	}
	var n pgtype.Numeric
	_ = n.Scan(strconv.FormatFloat(*f, 'f', -1, 64))
	return n
}

// pgTimeToPtr converts a TIME column (notification_preferences.quiet_hours_*)
// into a *time.Time anchored to the zero date — only the wall-clock
// component (via pgtype.Time's microseconds-since-midnight) is meaningful,
// the date is never read by callers.
func pgTimeToPtr(t pgtype.Time) *time.Time {
	if !t.Valid {
		return nil
	}
	wallClock := time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC).
		Add(time.Duration(t.Microseconds) * time.Microsecond)
	return &wallClock
}

// ptrToPgTime is the inverse of pgTimeToPtr: only the hour/minute/second/
// nanosecond components of t are read.
func ptrToPgTime(t *time.Time) pgtype.Time {
	if t == nil {
		return pgtype.Time{}
	}
	microseconds := int64(t.Hour())*3600e6 +
		int64(t.Minute())*60e6 +
		int64(t.Second())*1e6 +
		int64(t.Nanosecond())/1000
	return pgtype.Time{Microseconds: microseconds, Valid: true}
}

// pgIntervalToDuration converts moments.settling_time. That column is
// GENERATED ALWAYS AS (recorded_at - occurred_at) — a subtraction of two
// timestamptz values — so Postgres never populates the Months component
// for it; it's still handled here for a faithful Interval conversion.
func pgIntervalToDuration(i pgtype.Interval) time.Duration {
	if !i.Valid {
		return 0
	}
	return time.Duration(i.Months)*30*24*time.Hour +
		time.Duration(i.Days)*24*time.Hour +
		time.Duration(i.Microseconds)*time.Microsecond
}
