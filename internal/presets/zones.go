package presets

import (
	"sort"
	"strings"
)

// UrbanMacro maps each unit ID to an approximate geographic bucket.
// Heuristic buckets:
//   - zona-* — São Paulo municipality macro-areas
//   - metropolitana, interior, litoral — RM / state regions
//
// Preset "centro" is seeded at install from zona-central unit IDs; config may
// override or add other named preset lists.
const (
	ZoneCentral       = "zona-central"
	ZoneNorte         = "zona-norte"
	ZoneSul           = "zona-sul"
	ZoneLeste         = "zona-leste"
	ZoneOeste         = "zona-oeste"
	ZoneMetropolitana = "metropolitana"
	ZoneInterior      = "interior"
	ZoneLitoral       = "litoral"
)

// UrbanMacro returns a macro-zone for unit IDs observed on the production API.
func UrbanMacro(unitID string) string {
	if zone, ok := urbanMacroByID[unitID]; ok {
		return zone
	}
	return ""
}

// CanonicalUrbanMacroZone normalizes user input (e.g. from --where) to a known
// bucket string, or empty if it is not one of the heuristic urban labels.
func CanonicalUrbanMacroZone(input string) string {
	k := strings.ToLower(strings.TrimSpace(input))
	k = strings.ReplaceAll(k, " ", "-")
	k = strings.ReplaceAll(k, "_", "-")
	stripHyphens := func(s string) string { return strings.ReplaceAll(s, "-", "") }
	kCompact := stripHyphens(k)

	for _, v := range urbanMacroBuckets {
		lv := strings.ToLower(v)
		if k == lv || kCompact == stripHyphens(lv) {
			return v
		}
	}
	return ""
}

var urbanMacroBuckets = []string{
	ZoneCentral,
	ZoneNorte,
	ZoneSul,
	ZoneLeste,
	ZoneOeste,
	ZoneMetropolitana,
	ZoneInterior,
	ZoneLitoral,
}

// UnitIDsWithUrbanMacroZone returns sorted, deduplicated unit IDs for a
// heuristic geography bucket (excluding unknown IDs). Second return is false if
// input is not a known bucket.
//
// Aliases capital / cidade / municipio / grande-sp union all *municipal* macro
// zones (zona-central, norte, sul, leste, oeste, metropolitana) — i.e. Greater
// São Paulo city coverage without interior or litoral. This is usually what
// people mean by "SESC SP na capital"; the label metropolitana alone is only a
// small periphery subset in this table, not the full metro.
func UnitIDsWithUrbanMacroZone(zoneLabel string) ([]string, bool) {
	if ids, ok := unitIDsCapitalMunicipalityUnion(zoneLabel); ok {
		return ids, true
	}
	z := CanonicalUrbanMacroZone(zoneLabel)
	if z == "" {
		return nil, false
	}
	var out []string
	for id, got := range urbanMacroByID {
		if got == z {
			out = append(out, id)
		}
	}
	if len(out) == 0 {
		return nil, false
	}
	sort.Strings(out)
	return dedupe(out), true
}

var capitalMunicipalityMacroZones = map[string]bool{
	ZoneCentral:       true,
	ZoneNorte:         true,
	ZoneSul:           true,
	ZoneLeste:         true,
	ZoneOeste:         true,
	ZoneMetropolitana: true,
}

func unitIDsCapitalMunicipalityUnion(label string) ([]string, bool) {
	if !isCapitalMunicipalityAlias(label) {
		return nil, false
	}
	var out []string
	for id, zone := range urbanMacroByID {
		if capitalMunicipalityMacroZones[zone] {
			out = append(out, id)
		}
	}
	if len(out) == 0 {
		return nil, false
	}
	sort.Strings(out)
	return dedupe(out), true
}

func isCapitalMunicipalityAlias(label string) bool {
	k := strings.ToLower(strings.TrimSpace(label))
	k = strings.ReplaceAll(k, " ", "-")
	k = strings.ReplaceAll(k, "_", "-")
	for strings.Contains(k, "--") {
		k = strings.ReplaceAll(k, "--", "-")
	}
	compact := strings.ReplaceAll(k, "-", "")
	switch compact {
	case "capital", "capitalsp", "capitalsaopaulo", "spcapital", "grandesp", "cidade", "cidadesp", "municipio":
		return true
	default:
		return false
	}
}

// IsBuiltinWhereGeography reports whether label is a heuristic zone or the
// capital/cidade/… composite (pass through to where.Resolve without treating
// it as a venue slug in the CLI).
func IsBuiltinWhereGeography(label string) bool {
	if isCapitalMunicipalityAlias(label) {
		return true
	}
	return CanonicalUrbanMacroZone(label) != ""
}

// urbanMacroByID keys match normalize.Unit.ID / presets.Unit.ID.
var urbanMacroByID = map[string]string{
	"761":  ZoneCentral,
	"2":    ZoneCentral,
	"43":   ZoneCentral,
	"48":   ZoneCentral,
	"51":   ZoneCentral,
	"52":   ZoneCentral,
	"53":   ZoneCentral,
	"54":   ZoneCentral,
	"47":   ZoneLeste,
	"50":   ZoneLeste,
	"57":   ZoneLeste,
	"730":  ZoneNorte,
	"62":   ZoneNorte,
	"49":   ZoneSul,
	"55":   ZoneSul,
	"56":   ZoneSul,
	"63":   ZoneSul,
	"66":   ZoneSul,
	"60":   ZoneOeste,
	"61":   ZoneOeste,
	"58":   ZoneMetropolitana,
	"64":   ZoneMetropolitana,
	"65":   ZoneMetropolitana,
	"71":   ZoneMetropolitana,
	"80":   ZoneMetropolitana,
	"27":   ZoneLitoral,
	"37":   ZoneLitoral,
	"25":   ZoneInterior,
	"26":   ZoneInterior,
	"28":   ZoneInterior,
	"29":   ZoneInterior,
	"30":   ZoneInterior,
	"31":   ZoneInterior,
	"32":   ZoneInterior,
	"33":   ZoneInterior,
	"34":   ZoneInterior,
	"35":   ZoneInterior,
	"36":   ZoneInterior,
	"38":   ZoneInterior,
	"40":   ZoneInterior,
	"41":   ZoneInterior,
	"42":   ZoneInterior,
	"1005": ZoneInterior,
}

// BuiltinWhereGeographyExamples lists tokens users may pass for macro geography
// or special aliases (for help text; CanonicalUrbanMacroZone accepts more flexibly).
func BuiltinWhereGeographyExamples() []string {
	out := append([]string(nil), urbanMacroBuckets...)
	out = append(out,
		"centro", "center", "default",
		"capital", "cidade", "municipio", "grande-sp",
	)
	sort.Strings(out)
	return dedupe(out)
}

func init() {
	for i := range units {
		if z := UrbanMacro(units[i].ID); z != "" {
			units[i].Zone = z
		}
	}
}
