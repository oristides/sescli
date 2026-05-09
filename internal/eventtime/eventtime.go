package eventtime

import (
	"fmt"
	"time"

	"sescli/internal/normalize"
)

// DropStartedBefore retains events whose upcoming session is not strictly before now.
// Rows without a parsable start timestamp are kept so the API is not over-filtered.
func DropStartedBefore(events []normalize.Event, now time.Time) []normalize.Event {
	out := make([]normalize.Event, 0, len(events))
	for _, event := range events {
		if event.DateStart == "" {
			out = append(out, event)
			continue
		}
		start, err := ParseEventStart(event.DateStart, now.Location())
		if err != nil || !start.Before(now) {
			out = append(out, event)
		}
	}
	return out
}

func ParseEventStart(value string, loc *time.Location) (time.Time, error) {
	layouts := []string{"2006-01-02T15:04", time.RFC3339, time.DateOnly}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, value, loc); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported event time %q", value)
}
