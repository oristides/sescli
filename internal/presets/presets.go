package presets

import "strings"

// Defaults captures the v1 operator persona: adult, central units, cultural focus.
type DefaultSet struct {
	Audience      string
	Profile       string
	Preset        string
	UnitIDs       []string
	ActivityTypes []string
	PerPage       int
	Page          int
}

// CapitalUnitIDs returns the central SESC SP unit IDs captured from the Postman
// collection, deduplicated while preserving order.
func CapitalUnitIDs() []string {
	return dedupe(strings.Split(capitalUnitsCSV, ","))
}

// CentroUnitIDs is the v1 operator default: central venues only, not the full
// SESC "capital" API group that includes Interlagos, Sao Caetano, etc.
func CentroUnitIDs() []string {
	return dedupe(strings.Split(centroUnitsCSV, ","))
}

func UnitIDsForPreset(preset string) []string {
	switch strings.ToLower(strings.TrimSpace(preset)) {
	case "", "centro", "center", "default":
		return CentroUnitIDs()
	case "capital":
		return CapitalUnitIDs()
	default:
		return dedupe(strings.Split(preset, ","))
	}
}

func Defaults() DefaultSet {
	return DefaultSet{
		Audience:      "adulto",
		Profile:       "cultural",
		Preset:        "centro",
		UnitIDs:       CentroUnitIDs(),
		ActivityTypes: []string{"teatro", "cinema", "cursos-e-oficinas"},
		PerPage:       40,
		Page:          1,
	}
}

func ActivityTypesForProfile(profile string) []string {
	switch profile {
	case "", "cultural":
		return append([]string(nil), Defaults().ActivityTypes...)
	case "all", "any", "todos", "todas":
		return nil
	case "sports", "esportes":
		return []string{"esportes"}
	default:
		return []string{profile}
	}
}

const capitalUnitsCSV = "761,2,43,47,48,49,730,51,52,53,71,55,56,57,80,58,60,61,62,63,64,54,65,66,761,2,43,48,730,51,52,53,60,61,66"
const centroUnitsCSV = "2,43,51,52,53,60,61,66,761"

func dedupe(values []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}
