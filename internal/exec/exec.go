package exec

import (
	"strings"
	"time"

	"sescli/internal/query"
	"sescli/internal/what"
	"sescli/internal/when"
	"sescli/internal/where"
)

// QueryInput matches the semantic flags used by the CLI and MCP tools before
// resolution into canonical API bounds.
type QueryInput struct {
	When         string
	Where        string
	Preset       string
	Profile      string
	What         string
	Audience     string
	From         string
	To           string
	FromNow      bool
	Units        []string
	UnitNames    []string
	PerPage      int
	Page         int
	Format       string
	IncludeRaw   bool
	SummaryChars int
	// PresetUnitIDs is optional config.json "presets"; keys that match the
	// where expression override zonacentral for preset centro before resolving.
	PresetUnitIDs map[string][]string
}

func splitCSV(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			if trimmed := strings.TrimSpace(part); trimmed != "" {
				out = append(out, trimmed)
			}
		}
	}
	return out
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

// BuildQuery maps when/where/what flags into the canonical Query used for events.
func BuildQuery(in QueryInput, base time.Time) (query.Query, error) {
	if strings.TrimSpace(in.When) == "" && strings.TrimSpace(in.From) == "" && strings.TrimSpace(in.To) == "" {
		in.When = "today"
	}
	whatExpr := strings.TrimSpace(in.What)
	if whatExpr == "" {
		whatExpr = strings.TrimSpace(in.Profile)
	}
	if err := what.Validate(whatExpr); err != nil {
		return query.Query{}, err
	}
	resolvedWhat := what.Resolve(what.Filter{Expression: whatExpr, Audience: in.Audience})

	whereExpr := in.Where
	if whereExpr == "" {
		whereExpr = in.Preset
	}
	whereFilter := where.Filter{
		Expression:    whereExpr,
		IDs:           splitCSV(in.Units),
		ConfigPresets: in.PresetUnitIDs,
	}
	if len(in.UnitNames) > 0 {
		whereFilter.Expression = strings.Join(in.UnitNames, ",")
		whereFilter.IDs = nil
	}
	resolvedWhere, err := where.Resolve(whereFilter)
	if err != nil {
		return query.Query{}, err
	}

	whenFilter := when.Filter{From: in.From, To: in.To, FromNow: in.FromNow}
	forcedFromNow := in.FromNow
	if in.From != "" || in.To != "" {
		parsed, err := when.ParseRange(defaultString(in.From, "today"), defaultString(in.To, in.From), base)
		if err != nil {
			return query.Query{}, err
		}
		whenFilter = parsed
	} else if in.When != "" {
		parsed, err := when.Parse(in.When, base)
		if err != nil {
			return query.Query{}, err
		}
		whenFilter = parsed
	}
	whenFilter.FromNow = whenFilter.FromNow || forcedFromNow

	return query.Query{
		When:  whenFilter,
		Where: resolvedWhere,
		What:  resolvedWhat,
		Page:  query.PageOptions{Limit: in.PerPage, Page: in.Page},
		Out: query.OutputOptions{
			Format:       in.Format,
			IncludeRaw:   in.IncludeRaw,
			SummaryChars: in.SummaryChars,
		},
	}, nil
}
