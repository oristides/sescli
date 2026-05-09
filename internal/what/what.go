package what

import "strings"

type Filter struct {
	Expression string
	Audience   string
}

type Resolved struct {
	Audience      string
	ActivityTypes []string
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
	return Resolved{Audience: audience, ActivityTypes: activityTypes(expr)}
}

func activityTypes(expr string) []string {
	switch expr {
	case "all", "any", "todos", "todas":
		return nil
	case "cultural":
		return []string{
			"teatro",
			"cinema",
			"teatro-cursos-e-oficinas",
			"cinema-cursos-e-oficinas",
			"musica",
			"danca",
			"artes-visuais",
			"literatura-cursos-e-oficinas",
			"tecnologias-e-artes",
		}
	case "sports", "esportes":
		return []string{"esporte-e-atividade-fisica"}
	default:
		return splitCSV(expr)
	}
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
