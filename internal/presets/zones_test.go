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
	if CanonicalUrbanMacroZone("centro") != "" || CanonicalUrbanMacroZone("capital") != "" {
		t.Fatal("preset names must not parse as heuristic zones")
	}
}
