package presets

import "testing"

func TestUrbanMacroApproximateQuadrants(t *testing.T) {
	tests := map[string]string{
		"48": ZoneCentral,
		"60": ZoneOeste,
		"66": ZoneSul,
		"62": ZoneNorte,
		"47": ZoneLeste,
		"71": ZoneMetropolitana,
		"29": ZoneInterior,
		"37": ZoneLitoral,
	}
	for id, want := range tests {
		if got := UrbanMacro(id); got != want {
			t.Fatalf("UrbanMacro(%s)=%q want %q", id, got, want)
		}
	}
}

func TestEmbeddedUnitsCarryUrbanMacroZone(t *testing.T) {
	for _, u := range AllUnits() {
		if u.ID == "48" && u.Name == "Bom Retiro" {
			if u.Zone != ZoneCentral {
				t.Fatalf("Bom Retiro expected %s (geographic), got %q", ZoneCentral, u.Zone)
			}
			return
		}
	}
	t.Fatal("bom retiro unit missing")
}

func TestCapitalCompositeUnionCoversWestAndSouthUnits(t *testing.T) {
	ids, ok := UnitIDsWithUrbanMacroZone("capital")
	if !ok || len(ids) < 10 {
		t.Fatalf("capital union: ok=%v len=%d %#v", ok, len(ids), ids)
	}
	seen := map[string]bool{}
	for _, id := range ids {
		seen[id] = true
	}
	for _, need := range []string{"60", "56", "48"} {
		if !seen[need] {
			t.Fatalf("capital union missing id %s (need Pinheiros/sul/central coverage)", need)
		}
	}
}

func TestCanonicalUrbanMacroZoneFlexibleInput(t *testing.T) {
	for _, tt := range []struct{ in, want string }{
		{"zona-sul", ZoneSul},
		{"ZONA SUL", ZoneSul},
		{"zonasul", ZoneSul},
		{"zonacentral", ZoneCentral},
		{"interior", ZoneInterior},
	} {
		if got := CanonicalUrbanMacroZone(tt.in); got != tt.want {
			t.Fatalf("%q → %q, want %q", tt.in, got, tt.want)
		}
	}
	if CanonicalUrbanMacroZone("centro") != "" {
		t.Fatal("preset name centro must not parse as heuristic zone")
	}
	// "capital" is a composite resolved by UnitIDsWithUrbanMacroZone, not a single bucket.
	if CanonicalUrbanMacroZone("capital") != "" {
		t.Fatal("capital must not be a single CanonicalUrbanMacroZone bucket")
	}
	if !IsBuiltinWhereGeography("capital") || !IsBuiltinWhereGeography("zona-sul") || IsBuiltinWhereGeography("ipiranga") {
		t.Fatalf("IsBuiltinWhereGeography: capital=%v sul=%v ipi=%v",
			IsBuiltinWhereGeography("capital"), IsBuiltinWhereGeography("zona-sul"), IsBuiltinWhereGeography("ipiranga"))
	}
}

func TestBuiltinWhereGeographyExamples(t *testing.T) {
	got := BuiltinWhereGeographyExamples()
	if len(got) < 8 {
		t.Fatalf("expected several geography help tokens, got %#v", got)
	}
}
