// Package pagination provides offset-based and cursor-based pagination primitives
// shared across all bounded contexts.
package pagination

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// OffsetResult is the standard paginated response for offset-based queries.
type OffsetResult[T any] struct {
	Data       []T   `json:"data"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Size       int   `json:"size"`
	TotalPages int   `json:"total_pages"`
}

// NewOffsetResult builds an OffsetResult from a slice of items and total count.
func NewOffsetResult[T any](data []T, total int64, page, size int) OffsetResult[T] {
	if data == nil {
		data = []T{}
	}
	tp := 0
	if size > 0 {
		tp = int((total + int64(size) - 1) / int64(size))
	}
	return OffsetResult[T]{Data: data, Total: total, Page: page, Size: size, TotalPages: tp}
}

// CursorResult is a cursor-paginated response for large, append-only lists.
// Clients pass next_cursor back as ?cursor= to get the next page.
type CursorResult[T any] struct {
	Data       []T    `json:"data"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}

// NewCursorResult builds a CursorResult. Pass size+1 items; if len(items) > size
// the extra item is trimmed and next_cursor is set.
func NewCursorResult[T any](items []T, size int, cursorFn func(T) Cursor) CursorResult[T] {
	hasMore := len(items) > size
	if hasMore {
		items = items[:size]
	}
	result := CursorResult[T]{Data: items, HasMore: hasMore}
	if hasMore && len(items) > 0 {
		c := cursorFn(items[len(items)-1])
		result.NextCursor = c.Encode()
	}
	if result.Data == nil {
		result.Data = []T{}
	}
	return result
}

// Cursor represents the position of a cursor in a time-ordered list.
// It uses (created_at, id) as a composite key, which is stable and unique.
type Cursor struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

// ErrInvalidCursor is returned when a cursor string cannot be decoded.
var ErrInvalidCursor = errors.New("INVALID_CURSOR")

// Encode serialises the cursor to a URL-safe base64 string.
func (c Cursor) Encode() string {
	raw, _ := json.Marshal(c)
	return base64.RawURLEncoding.EncodeToString(raw)
}

// DecodeCursor parses a cursor string produced by Cursor.Encode.
func DecodeCursor(s string) (Cursor, error) {
	raw, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return Cursor{}, ErrInvalidCursor
	}
	var c Cursor
	if err := json.Unmarshal(raw, &c); err != nil {
		return Cursor{}, ErrInvalidCursor
	}
	return c, nil
}
