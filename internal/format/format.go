package format

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"sescli/internal/normalize"
)

type Meta struct {
	FetchedAt string `json:"fetched_at,omitempty"`
	Total     int    `json:"total,omitempty"`
	Source    string `json:"source,omitempty"`
	Query     string `json:"query,omitempty"`
}

type Envelope map[string]any

func Response(key string, value any, meta Meta) Envelope {
	if meta.FetchedAt == "" {
		meta.FetchedAt = time.Now().Format(time.RFC3339)
	}
	return Envelope{
		"_meta": meta,
		key:     value,
	}
}

func JSON(payload any, pretty bool) (string, error) {
	var (
		b   []byte
		err error
	)
	if pretty {
		b, err = json.MarshalIndent(payload, "", "  ")
	} else {
		b, err = json.Marshal(payload)
	}
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func WhatsApp(events []normalize.Event) string {
	if len(events) == 0 {
		return "Nenhum evento encontrado."
	}
	lines := make([]string, 0, len(events))
	for _, event := range events {
		parts := []string{event.Title}
		if event.PriceLabel != "" {
			parts = append(parts, event.PriceLabel)
		} else if event.Free {
			parts = append(parts, "Gratis")
		}
		if event.Venue != "" {
			parts = append(parts, event.Venue)
		}
		if event.URL != "" {
			parts = append(parts, event.URL)
		}
		line := strings.Join(nonEmpty(parts), " - ")
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func Table(events []normalize.Event) string {
	lines := []string{"TITLE\tWHEN\tVENUE\tPRICE\tURL"}
	for _, event := range events {
		when := strings.TrimSpace(event.DateStart + " " + event.TimeStart)
		lines = append(lines, fmt.Sprintf("%s\t%s\t%s\t%s\t%s", event.Title, when, event.Venue, event.PriceLabel, event.URL))
	}
	return strings.Join(lines, "\n")
}

func nonEmpty(values []string) []string {
	out := values[:0]
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			out = append(out, value)
		}
	}
	return out
}
