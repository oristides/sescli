package where

import (
	"fmt"
	"strings"

	"sescli/internal/presets"
)

type Filter struct {
	Expression string
	IDs        []string
}

type Resolved struct {
	IDs []string
}

func Resolve(filter Filter) (Resolved, error) {
	if len(filter.IDs) > 0 {
		return Resolved{IDs: splitCSV(filter.IDs)}, nil
	}
	expr := strings.TrimSpace(filter.Expression)
	if expr == "" || strings.EqualFold(expr, "centro") {
		return Resolved{IDs: presets.CentroUnitIDs()}, nil
	}
	if strings.EqualFold(expr, "capital") {
		return Resolved{IDs: presets.CapitalUnitIDs()}, nil
	}
	ids, err := presets.ResolveUnitIDs([]string{expr})
	if err != nil {
		return Resolved{}, fmt.Errorf("resolve where: %w", err)
	}
	return Resolved{IDs: ids}, nil
}

func splitCSV(values []string) []string {
	out := []string{}
	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			if trimmed := strings.TrimSpace(part); trimmed != "" {
				out = append(out, trimmed)
			}
		}
	}
	return out
}
