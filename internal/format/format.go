package format

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"sescli/internal/normalize"
)

type Meta struct {
	FetchedAt string `json:"fetched_at,omitempty"`
	// DateFrom / DateTo are the API window (YYYY-MM-DD), for hints and JSON _meta.
	DateFrom string `json:"date_from,omitempty"`
	DateTo   string `json:"date_to,omitempty"`
	// Total is the number of events in this response (after normalization).
	Total int `json:"total,omitempty"`
	// TotalReported is the API's total.value for this filter (may exceed len(events) when ppp limits rows).
	TotalReported *int   `json:"total_reported,omitempty"`
	HasMore       bool   `json:"has_more,omitempty"`
	Page          int    `json:"page,omitempty"`
	PerPage       int    `json:"per_page,omitempty"`
	Source        string `json:"source,omitempty"`
	Query         string `json:"query,omitempty"`
	// Min-results fill (optional --min-results): widened query until enough distinct events.
	MinResultsTarget       int  `json:"min_results_target,omitempty"`
	MinResultsWidenedDate  bool `json:"min_results_widened_date,omitempty"`
	MinResultsWidenedWhere bool `json:"min_results_widened_where,omitempty"`
}

// EventListMeta builds _meta for atividades/filter list calls.
func EventListMeta(fetched, page, perPage int, reported *int, source, query string) Meta {
	if page <= 0 {
		page = 1
	}
	m := Meta{
		Total:         fetched,
		TotalReported: reported,
		Page:          page,
		PerPage:       perPage,
		Source:        source,
		Query:         query,
	}
	if reported != nil {
		shown := (page-1)*perPage + fetched
		if *reported > shown {
			m.HasMore = true
		}
	} else if perPage > 0 && fetched >= perPage {
		m.HasMore = true
	}
	return m
}

// WhatsAppPaginationHint appends a short hint when more pages may exist (Portuguese, same as empty copy).
func WhatsAppPaginationHint(m Meta) string {
	if !m.HasMore {
		return ""
	}
	next := m.Page + 1
	if m.TotalReported != nil {
		return "\n— Há mais na API (total " + strconv.Itoa(*m.TotalReported) + "; nesta página " + strconv.Itoa(m.Total) + "). Tente --page " + strconv.Itoa(next) + " ou aumente --limit."
	}
	return "\n— Pode haver mais resultados (--page " + strconv.Itoa(next) + " ou --limit maior)."
}

// WhatsAppQueryFooter explains date range and API total so users do not confuse
// "few lines" with a broken limit when the filter is narrow (e.g. one day only).
func WhatsAppQueryFooter(m Meta) string {
	var b strings.Builder
	if m.DateFrom != "" && m.DateTo != "" {
		if m.DateFrom == m.DateTo {
			b.WriteString("\n— Período: ")
			b.WriteString(m.DateFrom)
			b.WriteString(" (só este dia). Para mais itens: use --from/--to com intervalo, --when weekend, --where mais amplo (ex.: capital), ou --what all.")
		} else {
			b.WriteString("\n— Período: ")
			b.WriteString(m.DateFrom)
			b.WriteString(" … ")
			b.WriteString(m.DateTo)
			b.WriteString(".")
		}
	}
	if m.TotalReported != nil {
		b.WriteString("\n— Total no filtro (API): ")
		b.WriteString(strconv.Itoa(*m.TotalReported))
		b.WriteString("; nesta página ")
		b.WriteString(strconv.Itoa(m.Total))
		b.WriteString(" (até ")
		b.WriteString(strconv.Itoa(m.PerPage))
		b.WriteString(" por página; --page 2 … se houver mais).")
	} else if m.PerPage > 0 && m.Total > 0 {
		b.WriteString("\n— Nesta resposta: ")
		b.WriteString(strconv.Itoa(m.Total))
		b.WriteString(" evento(s) (até ")
		b.WriteString(strconv.Itoa(m.PerPage))
		b.WriteString(" por página com --limit).")
	}
	if m.MinResultsTarget > 0 && (m.MinResultsWidenedDate || m.MinResultsWidenedWhere) {
		b.WriteString("\n— --min-results ")
		b.WriteString(strconv.Itoa(m.MinResultsTarget))
		b.WriteString(": janela ampliada")
		if m.MinResultsWidenedDate && m.DateTo != "" {
			b.WriteString(" até ")
			b.WriteString(m.DateTo)
		}
		if m.MinResultsWidenedWhere {
			b.WriteString("; unidades expandidas para a união municipal (capital)")
		}
		b.WriteString(".")
	}
	return b.String()
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
		if strings.TrimSpace(event.Summary) != "" {
			lines = append(lines, "  "+strings.TrimSpace(event.Summary))
		}
		if syn := strings.TrimSpace(event.Synopsis); syn != "" {
			lines = append(lines, "")
			lines = append(lines, syn)
		}
	}
	return strings.Join(lines, "\n")
}

func Table(events []normalize.Event) string {
	lines := []string{"TITLE\tSUMMARY\tSYNOPSIS\tWHEN\tVENUE\tPRICE\tURL"}
	for _, event := range events {
		when := strings.TrimSpace(event.DateStart + " " + event.TimeStart)
		syn := truncateRunes(strings.TrimSpace(event.Synopsis), 120)
		lines = append(lines, fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s", event.Title, event.Summary, syn, when, event.Venue, event.PriceLabel, event.URL))
	}
	return strings.Join(lines, "\n")
}

func truncateRunes(s string, max int) string {
	if max <= 0 {
		return ""
	}
	n := 0
	for i := range s {
		if n == max {
			return s[:i] + "..."
		}
		n++
	}
	return s
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
