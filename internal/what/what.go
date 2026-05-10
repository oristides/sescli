package what

import "strings"

// Slug the SESC site uses in atividades/filter for the parent bucket that
// contains theater as a child "linguagem" (atividade=teatro matches nothing).
type Filter struct {
	Expression string
	Audience   string
}

type Resolved struct {
	Audience      string
	ActivityTypes []string
	// Language is sent as linguagem= on atividades/filter (e.g. teatro under shows).
	Language string
}

func Resolve(filter Filter) Resolved {
	audience := strings.TrimSpace(filter.Audience)
	if audience == "" {
		audience = "adulto"
	}
	expr := strings.ToLower(strings.TrimSpace(filter.Expression))
	if expr == "" {
		expr = "cultural"
	}
	if expr == "teatro" {
		return Resolved{
			Audience:      audience,
			ActivityTypes: []string{ShowsPerformancesActivity},
			Language:      "teatro",
		}
	}
	return Resolved{Audience: audience, ActivityTypes: activityTypes(expr)}
}

func activityTypes(expr string) []string {
	switch expr {
	case "all", "any", "todos", "todas":
		return nil
	case "cultural":
		return mapTeatroToShowsParent(append([]string(nil), CulturalBundleSlugs...))
	case "sports", "esportes":
		return []string{"esporte-e-atividade-fisica"}
	default:
		raw := splitCSV(expr)
		parts := make([]string, len(raw))
		for i, p := range raw {
			parts[i] = normalizeExprToken(p)
		}
		return mapTeatroToShowsParent(parts)
	}
}

func mapTeatroToShowsParent(slugs []string) []string {
	out := make([]string, 0, len(slugs))
	for _, s := range slugs {
		if s == "teatro" {
			out = append(out, ShowsPerformancesActivity)
			continue
		}
		out = append(out, s)
	}
	return out
}

func splitCSV(value string) []string {
	out := []string{}
	for _, part := range strings.Split(value, ",") {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
