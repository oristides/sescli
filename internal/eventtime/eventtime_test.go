package eventtime

import (
	"testing"
	"time"

	"sescli/internal/normalize"
)

func TestDropStartedBeforeKeepsMissingStart(t *testing.T) {
	now := time.Date(2026, 5, 8, 20, 16, 0, 0, time.UTC)
	out := DropStartedBefore([]normalize.Event{{Title: "a"}}, now)
	if len(out) != 1 {
		t.Fatal(out)
	}
}

func TestParseEventStartLayouts(t *testing.T) {
	loc := time.UTC
	for _, tc := range []struct {
		in string
	}{
		{"2026-05-09T15:30"},
		{"2026-05-09T15:30:00Z"},
		{"2026-05-09"},
	} {
		if _, err := ParseEventStart(tc.in, loc); err != nil {
			t.Fatal(tc.in, err)
		}
	}
}
