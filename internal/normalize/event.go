package normalize

import (
	"fmt"
	"html"
	"regexp"
	"strconv"
	"strings"
)

// NormalizeOpts tweaks how raw API rows are turned into Event fields.
type NormalizeOpts struct {
	// SummaryMax is the max rune length for Summary after cleaning. 0 disables
	// truncation; values < 0 use the default (220).
	SummaryMax int
}

const defaultSummaryMax = 220

// DefaultNormalizeOpts matches historical sescli behavior.
func DefaultNormalizeOpts() NormalizeOpts {
	return NormalizeOpts{SummaryMax: defaultSummaryMax}
}

func (o NormalizeOpts) summaryLimit() int {
	if o.SummaryMax < 0 {
		return defaultSummaryMax
	}
	return o.SummaryMax
}

// EventPricing is ticket price data from the Java bilheteria API (via WordPress admin-ajax proxy).
type EventPricing struct {
	Gratuito          bool   `json:"gratuito,omitempty"`
	ValorInteira      string `json:"valor_inteira,omitempty"`
	ValorMeia         string `json:"valor_meia,omitempty"`
	ValorComerciario  string `json:"valor_comerciario,omitempty"`
	StatusIngresso    string `json:"status_ingresso,omitempty"`
	QtdeIngressosWeb  int    `json:"qtde_ingressos_web,omitempty"`
	QtdeIngressosRede int    `json:"qtde_ingressos_rede,omitempty"`
}

// Event is the compact, stable event shape emitted by sescli.
type Event struct {
	ID         string         `json:"id,omitempty"`
	JavaID     string         `json:"id_java,omitempty"`
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
	Pricing    *EventPricing  `json:"pricing,omitempty"`
	Categories []string       `json:"categories,omitempty"`
	Activities []string       `json:"activities,omitempty"`
	Summary    string         `json:"summary,omitempty"`
	Synopsis   string         `json:"synopsis,omitempty"`
	Raw        map[string]any `json:"raw,omitempty"`
}

// Unit is the stable shape for SESC physical spaces.
type Unit struct {
	ID         string         `json:"id,omitempty"`
	Name       string         `json:"name,omitempty"`
	Slug       string         `json:"slug,omitempty"`
	URL        string         `json:"url,omitempty"`
	APISegment string         `json:"api_segment,omitempty"`
	Zone       string         `json:"zone,omitempty"`
	Raw        map[string]any `json:"raw,omitempty"`
}

var tags = regexp.MustCompile(`<[^>]+>`)

func EventFromRaw(raw map[string]any, includeRaw bool) Event {
	return EventFromRawOpts(raw, includeRaw, DefaultNormalizeOpts())
}

// EventFromRawOpts is like EventFromRaw with custom normalization options.
func EventFromRawOpts(raw map[string]any, includeRaw bool, opts NormalizeOpts) Event {
	limit := opts.summaryLimit()
	event := Event{
		ID:         firstString(raw, "ID", "id", "post_id"),
		JavaID:     firstString(raw, "id_java", "idJava"),
		Title:      firstString(raw, "post_title", "titulo", "title", "nome", "name"),
		URL:        absoluteURL(firstString(raw, "permalink", "link", "url")),
		Venue:      firstVenue(raw),
		DateStart:  firstString(raw, "data_inicio", "date_start", "dataPrimeiraSessao", "dataProxSessao", "inicio"),
		DateEnd:    firstString(raw, "data_fim", "date_end", "dataUltimaSessao", "fim"),
		TimeStart:  firstString(raw, "hora_inicio", "time_start"),
		TimeEnd:    firstString(raw, "hora_fim", "time_end"),
		Free:       truthy(first(raw, "gratuito", "is_free", "free")),
		Online:     truthy(first(raw, "online", "is_online")),
		PriceLabel: firstString(raw, "preco", "price", "valor", "gratuito"),
		Categories: firstCategories(raw),
		Activities: stringList(first(raw, "tipo_atividade", "tipos_atividades", "atividade")),
		Summary: cleanText(firstString(raw,
			"post_excerpt", "resumo", "complemento", "sinopse",
			"post_content", "description",
		), limit),
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
	return EventsFromRawOpts(raw, includeRaw, DefaultNormalizeOpts())
}

// EventsFromRawOpts is like EventsFromRaw with custom normalization options.
func EventsFromRawOpts(raw any, includeRaw bool, opts NormalizeOpts) []Event {
	items := eventItems(raw)
	if len(items) == 0 {
		return nil
	}
	events := make([]Event, 0, len(items))
	for _, item := range items {
		event := EventFromRawOpts(item, includeRaw, opts)
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
		ID:         firstString(raw, "group_id", "term_id", "groupID", "ID", "id"),
		Name:       firstString(raw, "name", "groupName", "post_title", "title"),
		Slug:       firstString(raw, "group_slug", "slug", "groupLink", "post_name"),
		URL:        firstString(raw, "link", "permalink", "url"),
		APISegment: firstString(raw, "description", "groupType"),
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

// ApplyBilheteriaPricing attaches portal ticket data and fills aggregate price/free when the list row was incomplete.
func ApplyBilheteriaPricing(e *Event, p *EventPricing) {
	if e == nil || p == nil {
		return
	}
	e.Pricing = p
	if p.Gratuito {
		e.Free = true
		if e.PriceLabel == "" {
			e.PriceLabel = "Gratis"
		}
		return
	}
	var parts []string
	if s := strings.TrimSpace(p.ValorInteira); s != "" && !looksZeroBRL(s) {
		parts = append(parts, "Inteira "+s)
	}
	if s := strings.TrimSpace(p.ValorMeia); s != "" && !looksZeroBRL(s) {
		parts = append(parts, "Meia "+s)
	}
	if s := strings.TrimSpace(p.ValorComerciario); s != "" && !looksZeroBRL(s) {
		parts = append(parts, "Comerciário "+s)
	}
	if len(parts) > 0 && e.PriceLabel == "" {
		e.PriceLabel = strings.Join(parts, "; ")
	}
}

func looksZeroBRL(s string) bool {
	ls := strings.ToLower(strings.TrimSpace(strings.ReplaceAll(s, " ", "")))
	return ls == "$0.00" || ls == "r$0,00" || ls == "r$0.00" || ls == "0,00" || ls == "0.00"
}
