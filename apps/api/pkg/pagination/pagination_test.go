package pagination_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/pkg/pagination"
)

func TestOffsetResult_Empty(t *testing.T) {
	t.Parallel()
	r := pagination.NewOffsetResult([]string{}, 0, 1, 20)
	if r.Data == nil {
		t.Error("Data should be empty slice, not nil")
	}
	if r.TotalPages != 0 {
		t.Errorf("want TotalPages=0, got %d", r.TotalPages)
	}
}

func TestOffsetResult_TotalPages(t *testing.T) {
	t.Parallel()
	tests := []struct {
		total     int64
		size      int
		wantPages int
	}{
		{0, 20, 0},
		{1, 20, 1},
		{20, 20, 1},
		{21, 20, 2},
		{100, 20, 5},
		{101, 20, 6},
	}
	for _, tt := range tests {
		r := pagination.NewOffsetResult([]string{}, tt.total, 1, tt.size)
		if r.TotalPages != tt.wantPages {
			t.Errorf("total=%d size=%d: want %d pages, got %d", tt.total, tt.size, tt.wantPages, r.TotalPages)
		}
	}
}

func TestCursorResult_NoMore(t *testing.T) {
	t.Parallel()
	items := []string{"a", "b", "c"}
	r := pagination.NewCursorResult(items, 20, func(s string) pagination.Cursor {
		return pagination.Cursor{ID: uuid.New(), CreatedAt: time.Now()}
	})
	if r.HasMore {
		t.Error("expected HasMore=false when fewer items than page size")
	}
	if r.NextCursor != "" {
		t.Error("expected no next_cursor")
	}
	if len(r.Data) != 3 {
		t.Errorf("want 3 items, got %d", len(r.Data))
	}
}

func TestCursorResult_HasMore(t *testing.T) {
	t.Parallel()
	// 6 items fetched but page size is 5 → has more
	items := []string{"a", "b", "c", "d", "e", "f"}
	r := pagination.NewCursorResult(items, 5, func(s string) pagination.Cursor {
		return pagination.Cursor{ID: uuid.New(), CreatedAt: time.Now()}
	})
	if !r.HasMore {
		t.Error("expected HasMore=true")
	}
	if r.NextCursor == "" {
		t.Error("expected a next_cursor")
	}
	if len(r.Data) != 5 {
		t.Errorf("want 5 items after trim, got %d", len(r.Data))
	}
}

func TestCursorEncodeDecodeRoundtrip(t *testing.T) {
	t.Parallel()
	id := uuid.New()
	ts := time.Now().UTC().Truncate(time.Millisecond)
	c := pagination.Cursor{ID: id, CreatedAt: ts}
	encoded := c.Encode()
	decoded, err := pagination.DecodeCursor(encoded)
	if err != nil {
		t.Fatalf("DecodeCursor: %v", err)
	}
	if decoded.ID != id {
		t.Errorf("ID mismatch: want %v, got %v", id, decoded.ID)
	}
	if !decoded.CreatedAt.Equal(ts) {
		t.Errorf("CreatedAt mismatch: want %v, got %v", ts, decoded.CreatedAt)
	}
}

func TestDecodeCursor_Invalid(t *testing.T) {
	t.Parallel()
	_, err := pagination.DecodeCursor("not-valid-base64!!!")
	if err == nil {
		t.Fatal("expected error for invalid cursor")
	}
}
