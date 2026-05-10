package presets

import "strings"

// Defaults captures defaults: adult audience, municipal São Paulo geography
// (capital union), cultural profile.
type DefaultSet struct {
	Audience      string
	Profile       string
	Preset        string
	UnitIDs       []string
	ActivityTypes []string
	PerPage       int
	Page          int
}

// DefaultCapitalUnitIDs is the builtin municipal union (all zona-* + metropolitana).
func DefaultCapitalUnitIDs() []string {
	ids, ok := UnitIDsWithUrbanMacroZone("capital")
	if !ok || len(ids) == 0 {
		return nil
	}
	return append([]string(nil), ids...)
}

// DefaultInstallPresets returns preset keys and unit IDs seeded into config.json on init.
// Built only from heuristic geography — no hand-maintained centro/capital CSV lists.
func DefaultInstallPresets() map[string][]string {
	ids, ok := UnitIDsWithUrbanMacroZone(ZoneCentral)
	if !ok || len(ids) == 0 {
		return map[string][]string{"centro": {}}
	}
	return map[string][]string{
		"centro": append([]string(nil), ids...),
	}
}

// ClonePresetMap returns a shallow copy with copied ID slices (safe defaults for config).
func ClonePresetMap(src map[string][]string) map[string][]string {
	if src == nil {
		return nil
	}
	out := make(map[string][]string, len(src))
	for k, v := range src {
		out[k] = append([]string(nil), v...)
	}
	return out
}

// DefaultCentroPresetUnitIDs is the zonacentral unit list — same IDs written to config presets.centro.
func DefaultCentroPresetUnitIDs() []string {
	m := DefaultInstallPresets()
	return append([]string(nil), m["centro"]...)
}

// UnitIDsForPreset resolves `--preset NAME` dinamico unit lists outside of config.json.
func UnitIDsForPreset(preset string) []string {
	p := strings.ToLower(strings.TrimSpace(preset))
	switch p {
	case "", "centro", "center", "default":
		return DefaultCentroPresetUnitIDs()
	default:
		if ids, ok := UnitIDsWithUrbanMacroZone(preset); ok {
			return append([]string(nil), ids...)
		}
		return Dedupe(strings.Split(preset, ","))
	}
}

func Defaults() DefaultSet {
	return DefaultSet{
		Audience: "adulto",
		Profile:  "cultural",
		Preset:   "capital",
		UnitIDs:  DefaultCapitalUnitIDs(),
		// Match atividades/filter taxonomy: theater is under shows-espetaculos-e-performances, not atividade=teatro.
		ActivityTypes: []string{"shows-espetaculos-e-performances", "cinema", "cursos-e-oficinas"},
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

// Dedupe removes duplicates and empty tokens while keeping first-seen order.
func Dedupe(values []string) []string {
	return dedupe(values)
}
