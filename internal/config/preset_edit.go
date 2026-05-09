package config

import (
	"fmt"
	"sort"
	"strings"

	"sescli/internal/presets"
)

// ParseUnitIDArgs parses comma- and token-separated unit IDs from CLI args.
func ParseUnitIDArgs(args []string) []string {
	var raw []string
	for _, a := range args {
		for _, part := range strings.Split(a, ",") {
			p := strings.TrimSpace(part)
			if p != "" {
				raw = append(raw, p)
			}
		}
	}
	return presets.Dedupe(raw)
}

// NormalizePresetName trims and lowercases a preset key for storage and lookup.
// Aliases center and default map to centro so edits line up with --where centro.
func NormalizePresetName(name string) (string, error) {
	s := strings.TrimSpace(strings.ToLower(name))
	if s == "" {
		return "", fmt.Errorf("preset name is required")
	}
	switch s {
	case "center", "default":
		s = "centro"
	}
	return s, nil
}

// EffectivePreset returns the unit IDs used for --where <name>: non-empty entry
// in config first; for centro (and aliases) otherwise zonacentral geography; custom
// names only from config.
func EffectivePreset(cfg Config, name string) []string {
	key, err := NormalizePresetName(name)
	if err != nil {
		return nil
	}
	if ids := cfg.Presets[key]; len(ids) > 0 {
		return presets.Dedupe(ids)
	}
	if key == "centro" {
		return zonaCentralEffective()
	}
	return nil
}

func zonaCentralEffective() []string {
	ids, ok := presets.UnitIDsWithUrbanMacroZone(presets.ZoneCentral)
	if !ok || len(ids) == 0 {
		return nil
	}
	return append([]string(nil), ids...)
}

// PresetNamesForList returns sorted preset keys (defaults plus any extras in cfg).
func PresetNamesForList(cfg Config) []string {
	seen := map[string]bool{"centro": true}
	for k := range cfg.Presets {
		seen[k] = true
	}
	out := make([]string, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func mergeUnitIDs(base, add []string) []string {
	return presets.Dedupe(append(append([]string(nil), base...), add...))
}

func removeUnitIDs(base []string, remove []string) []string {
	rm := map[string]bool{}
	for _, id := range remove {
		rm[id] = true
	}
	out := make([]string, 0, len(base))
	for _, id := range base {
		if !rm[id] {
			out = append(out, id)
		}
	}
	return out
}

// Update loads config from path (or default path), applies mutate, and writes with overwrite.
func Update(path string, mutate func(*Config) error) error {
	if path == "" {
		var err error
		path, err = Path()
		if err != nil {
			return err
		}
	}
	cfg, err := Load(path)
	if err != nil {
		return err
	}
	if err := mutate(&cfg); err != nil {
		return err
	}
	return Write(path, cfg, true)
}

// SetPresetIDs replaces the stored list for name (deduped). cfg.Presets is allocated if needed.
func SetPresetIDs(cfg *Config, name string, ids []string) error {
	key, err := NormalizePresetName(name)
	if err != nil {
		return err
	}
	if cfg.Presets == nil {
		cfg.Presets = map[string][]string{}
	}
	ids = presets.Dedupe(ids)
	cfg.Presets[key] = append([]string(nil), ids...)
	return nil
}

// AddPresetIDs unions effective IDs with extra and stores under name.
func AddPresetIDs(cfg *Config, name string, extra []string) error {
	key, err := NormalizePresetName(name)
	if err != nil {
		return err
	}
	base := EffectivePreset(*cfg, key)
	merged := mergeUnitIDs(base, extra)
	return SetPresetIDs(cfg, key, merged)
}

// RemovePresetIDs removes IDs from the effective list and stores the result.
// If nothing remains, the preset key is removed from config (same as unset).
func RemovePresetIDs(cfg *Config, name string, remove []string) error {
	key, err := NormalizePresetName(name)
	if err != nil {
		return err
	}
	base := EffectivePreset(*cfg, key)
	out := removeUnitIDs(base, presets.Dedupe(remove))
	if len(out) == 0 {
		return UnsetPreset(cfg, key)
	}
	return SetPresetIDs(cfg, key, out)
}

// UnsetPreset removes the preset key from config so the next resolve uses zonacentral
// for centro again (no stored override).
func UnsetPreset(cfg *Config, name string) error {
	key, err := NormalizePresetName(name)
	if err != nil {
		return err
	}
	if cfg.Presets == nil {
		return nil
	}
	delete(cfg.Presets, key)
	return nil
}
