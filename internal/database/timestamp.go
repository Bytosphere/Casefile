package database

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// sqliteTimeLayout matches the RFC3339 format written by the created_at/etc.
// column defaults (see migration's strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
// and by Go-side inserts/seed data.
const sqliteTimeLayout = "2006-01-02T15:04:05Z07:00" // time.RFC3339

// Timestamp wraps time.Time so it can Scan modernc.org/sqlite's TEXT-typed
// datetime columns directly, and Value back into the same format on write.
type Timestamp struct {
	time.Time
}

func (t *Timestamp) Scan(src any) error {
	if src == nil {
		return nil
	}
	switch v := src.(type) {
	case string:
		parsed, err := time.Parse(sqliteTimeLayout, v)
		if err != nil {
			return fmt.Errorf("timestamp: parse %q: %w", v, err)
		}
		t.Time = parsed
	case time.Time:
		t.Time = v
	default:
		return fmt.Errorf("timestamp: unsupported scan type %T", src)
	}
	return nil
}

func (t *Timestamp) Value() (driver.Value, error) {
	if t.Time.IsZero() {
		return nil, nil
	}
	return t.Time.Format(sqliteTimeLayout), nil
}
