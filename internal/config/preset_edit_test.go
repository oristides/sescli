package config

import (
	"slices"
	"testing"

	"sescli/internal/presets"
)

func TestParseUnitIDArgs(t *testing.T) {
	got := ParseUnitIDArgs([]string{"1, 2,", " 3", "4,4"})
	want := []string{"1", "2", "3", "4"}
	if !slices.Equal(got, want) {
		t.Fatalf("ParseUnitIDArgs got %v want %v", got, want)
	}
}

func TestNormalizePresetNameAliases(t *testing.T) {
	for _, tc := range []struct{ raw, want string }{
		{"Center", "centro"},
		{"DEFAULT", "centro"},
		{"Foo ", "foo"},
		{"perto-do-trabalho", "perto-do-trabalho"},
	} {
		got, err := NormalizePresetName(tc.raw)
		if err != nil {
			t.Fatalf("%q: %v", tc.raw, err)
		}
		if got != tc.want {
			t.Fatalf("%q: got %q want %q", tc.raw, got, tc.want)
		}
	}
}

func TestEffectivePresetConfigOverridesCentroBuiltin(t *testing.T) {
	cfg := Default()
	cfg.Presets["centro"] = []string{"999"}
	got := EffectivePreset(cfg, "centro")
	want := []string{"999"}
	if !slices.Equal(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestEffectivePresetBuiltinCentroWhenUnset(t *testing.T) {
	cfg := Config{Presets: map[string][]string{}}
	got := EffectivePreset(cfg, "centro")
	want, ok := presets.UnitIDsWithUrbanMacroZone(presets.ZoneCentral)
	if !ok || !slices.Equal(got, want) {
		t.Fatalf("effective centro got %v ok=%v vs zone want %v", got, ok, want)
	}
}

func TestAddPresetIDsStartsFromEffectiveBuiltinCentro(t *testing.T) {
	cfg := Default()
	delete(cfg.Presets, "centro")
	if err := AddPresetIDs(&cfg, "centro", []string{"000"}); err != nil {
		t.Fatal(err)
	}
	base := presets.DefaultCentroPresetUnitIDs()
	merged := presets.Dedupe(append(append([]string(nil), base...), "000"))
	if !slices.Equal(cfg.Presets["centro"], merged) {
		t.Fatalf("unexpected merge: %#v", cfg.Presets["centro"])
	}
}

func TestUnsetPreset(t *testing.T) {
	cfg := Default()
	cfg.Presets["centro"] = []string{"1"}
	if err := UnsetPreset(&cfg, "centro"); err != nil {
		t.Fatal(err)
	}
	if _, ok := cfg.Presets["centro"]; ok {
		t.Fatalf("expected centro key cleared")
	}
}

func TestRemovePresetIDsDropsKeyWhenNothingLeft(t *testing.T) {
	cfg := Default()
	cfg.Presets["centro"] = []string{"2"}
	if err := RemovePresetIDs(&cfg, "centro", []string{"2"}); err != nil {
		t.Fatal(err)
	}
	if _, ok := cfg.Presets["centro"]; ok {
		t.Fatal("expected centro removed when last id stripped")
	}
}
