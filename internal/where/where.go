package where

import (
	"fmt"
	"strings"

	"sescli/internal/presets"
)

type Filter struct {
	Expression    string
	IDs           []string
	ConfigPresets map[string][]string `json:"-"` // optional: ~/.config/sescli preset lists
}

type Resolved struct {
	IDs []string
}

func Resolve(filter Filter) (Resolved, error) {
	if len(filter.IDs) > 0 {
		return Resolved{IDs: splitCSV(filter.IDs)}, nil
	}
	expr := strings.TrimSpace(filter.Expression)
	presetLookup := expr
	if presetLookup == "" {
		presetLookup = "centro"
	}
	if ids, ok := configPreset(filter.ConfigPresets, presetLookup); ok {
		return Resolved{IDs: ids}, nil
	}
	if expr == "" || strings.EqualFold(expr, "centro") || strings.EqualFold(expr, "center") || strings.EqualFold(expr, "default") {
		ids, ok := presets.UnitIDsWithUrbanMacroZone(presets.ZoneCentral)
		if !ok || len(ids) == 0 {
			return Resolved{}, fmt.Errorf("where: zonacentral geography table has no units")
		}
		return Resolved{IDs: ids}, nil
	}
	if ids, ok := presets.UnitIDsWithUrbanMacroZone(expr); ok {
		return Resolved{IDs: ids}, nil
	}
	ids, err := presets.ResolveUnitIDs([]string{expr})
	if err != nil {
		return Resolved{}, invalidExpressionErr(expr)
	}
	return Resolved{IDs: ids}, nil
}

func configPreset(m map[string][]string, name string) ([]string, bool) {
	if m == nil || name == "" {
		return nil, false
	}
	name = strings.ToLower(strings.TrimSpace(name))
	for k, ids := range m {
		if strings.ToLower(strings.TrimSpace(k)) != name {
			continue
		}
		if len(ids) == 0 {
			return nil, false
		}
		return presets.Dedupe(ids), true
	}
	return nil, false
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
