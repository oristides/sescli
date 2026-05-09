package presets

import "testing"

func TestCapitalPresetKeepsBroadAPIGroupWithoutDuplicates(t *testing.T) {
	ids := CapitalUnitIDs()

	if len(ids) < 20 {
		t.Fatalf("capital preset should include the broad API capital group, got %d ids", len(ids))
	}
	if ids[0] != "761" || ids[1] != "2" {
		t.Fatalf("capital preset should preserve the known leading IDs, got %#v", ids[:2])
	}

	seen := map[string]bool{}
	for _, id := range ids {
		if seen[id] {
			t.Fatalf("capital preset should deduplicate repeated Postman IDs, duplicated %s", id)
		}
		seen[id] = true
	}
}

func TestCentroPresetIsDefaultAndExcludesFarCapitalUnits(t *testing.T) {
	ids := CentroUnitIDs()
	joined := map[string]bool{}
	for _, id := range ids {
		joined[id] = true
	}

	for _, far := range []string{"55", "57", "58", "64", "65", "71"} {
		if joined[far] {
			t.Fatalf("centro default should not include far capital unit %s in %#v", far, ids)
		}
	}
	for _, central := range []string{"2", "43", "51", "52", "53"} {
		if !joined[central] {
			t.Fatalf("centro default should include %s in %#v", central, ids)
		}
	}
	if Defaults().Preset != "centro" {
		t.Fatalf("default preset should be centro, got %q", Defaults().Preset)
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
	if len(defaults.UnitIDs) == 0 || defaults.UnitIDs[0] != "2" {
		t.Fatalf("expected default centro units, got %#v", defaults.UnitIDs)
	}
}

func TestAllProfileRemovesActivityTypes(t *testing.T) {
	if got := ActivityTypesForProfile("all"); len(got) != 0 {
		t.Fatalf("expected all profile to remove filters, got %#v", got)
	}
}
