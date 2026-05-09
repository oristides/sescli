package normalize

import (
	"fmt"
	"html"
	"regexp"
	"strconv"
	"strings"
)

// Event is the compact, stable event shape emitted by sescli.
type Event struct {
	ID         string         `json:"id,omitempty"`
	Title      string         `json:"title,omitempty"`
	URL        string         `json:"url,omitempty"`
	Venue      string         `json:"venue,omitempty"`
	DateStart  string         `json:"date_start,omitempty"`
	DateEnd    string         `json:"date_end,omitempty"`
	TimeStart  string         `json:"time_start,omitempty"`
	TimeEnd    string         `json:"time_end,omitempty"`
	Free       bool           `json:"is_free,omitempty"`
	Online     bool           `json:"is_online,omitempty"`
	PriceLabel string         `json:"price,omitempty"`
	Categories []string       `json:"categories,omitempty"`
	Activities []string       `json:"activities,omitempty"`
	Summary    string         `json:"summary,omitempty"`
	Raw        map[string]any `json:"raw,omitempty"`
}

// Unit is the stable shape for SESC physical spaces.
type Unit struct {
	ID   string         `json:"id,omitempty"`
	Name string         `json:"name,omitempty"`
	Slug string         `json:"slug,omitempty"`
	URL  string         `json:"url,omitempty"`
	Raw  map[string]any `json:"raw,omitempty"`
}

var tags = regexp.MustCompile(`<[^>]+>`)

func EventFromRaw(raw map[string]any, includeRaw bool) Event {
	event := Event{
		ID:         firstString(raw, "ID", "id", "post_id"),
		Title:      firstString(raw, "post_title", "titulo", "title", "nome", "name"),
		URL:        absoluteURL(firstString(raw, "permalink", "link", "url")),
		Venue:      firstVenue(raw),
		DateStart:  firstString(raw, "data_inicio", "date_start", "dataPrimeiraSessao", "dataProxSessao", "inicio"),
		DateEnd:    firstString(raw, "data_fim", "date_end", "fim"),
		TimeStart:  firstString(raw, "hora_inicio", "time_start"),
		TimeEnd:    firstString(raw, "hora_fim", "time_end"),
		Free:       truthy(first(raw, "gratuito", "is_free", "free")),
		Online:     truthy(first(raw, "online", "is_online")),
		PriceLabel: firstString(raw, "preco", "price", "valor", "gratuito"),
		Categories: firstCategories(raw),
		Activities: stringList(first(raw, "tipo_atividade", "tipos_atividades", "atividade")),
		Summary:    cleanText(firstString(raw, "post_excerpt", "resumo", "post_content", "description"), 220),
	}
	if event.Free && event.PriceLabel == "" {
		event.PriceLabel = "Gratis"
	}
	if event.Free && strings.Contains(strings.ToLower(event.PriceLabel), "gratuit") {
		event.PriceLabel = "Gratis"
	}
	if !event.Free && strings.Contains(strings.ToLower(event.PriceLabel), "paga") {
		event.PriceLabel = "Pago"
	}
	if includeRaw {
		event.Raw = raw
	}
	return event
}

func EventsFromRaw(raw any, includeRaw bool) []Event {
	items := eventItems(raw)
	if len(items) == 0 {
		return nil
	}
	events := make([]Event, 0, len(items))
	for _, item := range items {
		event := EventFromRaw(item, includeRaw)
		if event.Title == "" && event.ID == "" {
			continue
		}
		events = append(events, event)
	}
	return events
}

func eventItems(raw any) []map[string]any {
	items, ok := raw.([]any)
	if !ok {
		if obj, ok := raw.(map[string]any); ok {
			return eventItems(obj["atividade"])
		}
		return nil
	}
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}
		out = append(out, obj)
	}
	return out
}

func UnitFromRaw(raw map[string]any, includeRaw bool) Unit {
	unit := Unit{
		ID:   firstString(raw, "term_id", "groupID", "ID", "id"),
		Name: firstString(raw, "name", "groupName", "post_title", "title"),
		Slug: firstString(raw, "slug", "groupLink", "post_name"),
		URL:  firstString(raw, "link", "permalink", "url"),
	}
	if includeRaw {
		unit.Raw = raw
	}
	return unit
}

func UnitsFromRaw(raw any, includeRaw bool) []Unit {
	items := unitItems(raw)
	units := make([]Unit, 0, len(items))
	for _, item := range items {
		unit := UnitFromRaw(item, includeRaw)
		if unit.ID == "" && unit.Name == "" {
			continue
		}
		units = append(units, unit)
	}
	return units
}

func unitItems(raw any) []map[string]any {
	items, ok := raw.([]any)
	if ok {
		out := make([]map[string]any, 0, len(items))
		for _, item := range items {
			if obj, ok := item.(map[string]any); ok {
				out = append(out, obj)
			}
		}
		return out
	}
	obj, ok := raw.(map[string]any)
	if !ok {
		return nil
	}
	if unidades, ok := obj["unidades"].(map[string]any); ok {
		out := []map[string]any{}
		for _, grouped := range unidades {
			out = append(out, unitItems(grouped)...)
		}
		return out
	}
	return nil
}

func first(raw map[string]any, keys ...string) any {
	for _, key := range keys {
		if value, ok := raw[key]; ok && value != nil {
			return value
		}
	}
	return nil
}

func firstString(raw map[string]any, keys ...string) string {
	return toString(first(raw, keys...))
}

func toString(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(v)
	case float64:
		if v == float64(int64(v)) {
			return strconv.FormatInt(int64(v), 10)
		}
		return strconv.FormatFloat(v, 'f', -1, 64)
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	default:
		return strings.TrimSpace(fmt.Sprint(v))
	}
}

func truthy(value any) bool {
	switch v := value.(type) {
	case bool:
		return v
	case string:
		v = strings.ToLower(strings.TrimSpace(v))
		return v == "1" || v == "true" || v == "sim" || v == "gratuito" || v == "gratis" || strings.Contains(v, "gratuit")
	case float64:
		return v != 0
	case int:
		return v != 0
	default:
		return false
	}
}

func stringList(value any) []string {
	switch v := value.(type) {
	case nil:
		return nil
	case string:
		if strings.TrimSpace(v) == "" {
			return nil
		}
		parts := strings.Split(v, ",")
		out := make([]string, 0, len(parts))
		for _, part := range parts {
			if trimmed := strings.TrimSpace(part); trimmed != "" {
				out = append(out, trimmed)
			}
		}
		return out
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s := toString(item); s != "" {
				out = append(out, s)
			}
		}
		return out
	case []string:
		return v
	default:
		if s := toString(v); s != "" {
			return []string{s}
		}
		return nil
	}
}

func cleanText(input string, limit int) string {
	input = html.UnescapeString(tags.ReplaceAllString(input, " "))
	input = strings.Join(strings.Fields(input), " ")
	if limit > 0 && len(input) > limit {
		return input[:limit] + "..."
	}
	return input
}

func absoluteURL(value string) string {
	if value == "" || strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") {
		return value
	}
	if strings.HasPrefix(value, "/") {
		return "https://www.sescsp.org.br" + value
	}
	return value
}

func firstVenue(raw map[string]any) string {
	if value := firstString(raw, "unidade_nome", "venue", "local"); value != "" {
		return value
	}
	if units, ok := raw["unidade"].([]any); ok && len(units) > 0 {
		if obj, ok := units[0].(map[string]any); ok {
			return firstString(obj, "name", "nome", "title")
		}
	}
	return firstString(raw, "unidade")
}

func firstCategories(raw map[string]any) []string {
	if titles := titlesFromList(first(raw, "categorias", "categoria", "linguagens")); len(titles) > 0 {
		return titles
	}
	categories := stringList(first(raw, "categorias", "categoria", "linguagens"))
	if len(categories) > 0 {
		return categories
	}
	return titlesFromList(first(raw, "tipos_linguagens", "conjunto"))
}

func titlesFromList(value any) []string {
	items, ok := value.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if title := firstString(obj, "titulo", "name", "title"); title != "" {
			out = append(out, title)
		}
	}
	return out
}
