package presets

import (
	"slices"
	"testing"
)

func TestDefaultInstallCentroPresetMatchesZonaCentral(t *testing.T) {
	zoneCentral, ok := UnitIDsWithUrbanMacroZone(ZoneCentral)
	install := DefaultInstallPresets()
	cent := install["centro"]
	defaultsUnits := Defaults().UnitIDs

	if !ok || len(zoneCentral) == 0 {
		t.Fatalf("expected zonacentral ids")
	}
	if len(cent) == 0 {
		t.Fatalf("expected seeded centro preset: %#v", install)
	}
	if !slices.Equal(cent, zoneCentral) {
		t.Fatalf("install centro != zonacentral: %#v vs %#v", cent, zoneCentral)
	}
	capitalIDs, ok := UnitIDsWithUrbanMacroZone("capital")
	if !ok || len(capitalIDs) == 0 {
		t.Fatalf("expected capital union ids")
	}
	if !slices.Equal(defaultsUnits, capitalIDs) {
		t.Fatalf("defaults unit ids %#v vs capital %#v", defaultsUnits, capitalIDs)
	}
}

func TestDefaultCentroUsesGeographicFilterNotRetrofits(t *testing.T) {
	ids := DefaultCentroPresetUnitIDs()
	joined := map[string]bool{}
	for _, id := range ids {
		joined[id] = true
	}
	for id := range joined {
		if UrbanMacro(id) != ZoneCentral {
			t.Fatalf("id %s in default centro preset is not zonacentral (got macro %q)", id, UrbanMacro(id))
		}
	}
}

func TestDefaultsPresetName(t *testing.T) {
	if Defaults().Preset != "capital" {
		t.Fatalf("expected default preset name capital")
	}
}

func TestDefaultProfileMatchesAdultCulturalTaste(t *testing.T) {
	defaults := Defaults()

	if defaults.Audience != "adulto" {
		t.Fatalf("expected adulto audience, got %q", defaults.Audience)
	}
	if defaults.Profile != "cultural" {
		t.Fatalf("expected cultural profile, got %q", defaults.Profile)
	}
	if defaults.Page != 1 || defaults.PerPage < 20 {
		t.Fatalf("pagination defaults are not useful: %#v", defaults)
	}
	if got := defaults.ActivityTypes; len(got) < 2 {
		t.Fatalf("expected theater/cinema/workshop activity types, got %#v", got)
	}
	if n := len(DefaultCentroPresetUnitIDs()); len(defaults.UnitIDs) <= n {
		t.Fatalf("expected default capital union wider than zona-central (centro %d ids, got %d)", n, len(defaults.UnitIDs))
	}
}

func TestClonePresetMapIsIndependent(t *testing.T) {
	a := DefaultInstallPresets()
	b := ClonePresetMap(a)
	b["centro"] = append(b["centro"], "999")
	if slices.Equal(a["centro"], b["centro"]) {
		t.Fatal("slice should have been copied")
	}
}

func TestAllProfileRemovesActivityTypes(t *testing.T) {
	if got := ActivityTypesForProfile("all"); len(got) != 0 {
		t.Fatalf("expected all profile to remove filters, got %#v", got)
	}
}
