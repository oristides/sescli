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
func UnitIDsWithUrbanMacroZone(zoneLabel string) ([]string, bool) {
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

func init() {
	for i := range units {
		if z := UrbanMacro(units[i].ID); z != "" {
			units[i].Zone = z
		}
	}
}
