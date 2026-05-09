package where

import "testing"

func TestResolveConfigOverridesCentro(t *testing.T) {
	got, err := Resolve(Filter{
		Expression: "centro",
		ConfigPresets: map[string][]string{
			"centro": {"56", "56", "61"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.IDs) != 2 || got.IDs[0] != "56" || got.IDs[1] != "61" {
		t.Fatalf("want custom centro ids, got %#v", got.IDs)
	}
}

func TestResolveConfigCustomPresetName(t *testing.T) {
	got, err := Resolve(Filter{
		Expression:    "perto",
		ConfigPresets: map[string][]string{"perto": {"43", "52"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.IDs) != 2 || got.IDs[0] != "43" {
		t.Fatalf("got %#v", got.IDs)
	}
}

func TestResolveUrbanMacroZone(t *testing.T) {
	got, err := Resolve(Filter{Expression: "zona-sul"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.IDs) < 3 {
		t.Fatalf("expected several sul ids, got %#v", got.IDs)
	}
	want := map[string]bool{"49": true, "55": true, "56": true, "63": true, "66": true}
	for _, id := range got.IDs {
		if !want[id] {
			t.Fatalf("unexpected id %s in zona-sul", id)
		}
	}
}

func TestResolvePresetAndVenueNames(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"centro", "2"},
		{"ipiranga", "56"},
		{"iporanga", "56"},
		{"56", "56"},
	}

	for _, tt := range tests {
		got, err := Resolve(Filter{Expression: tt.input})
		if err != nil {
			t.Fatalf("Resolve(%q): %v", tt.input, err)
		}
		if len(got.IDs) == 0 || got.IDs[0] != tt.want {
			t.Fatalf("Resolve(%q) = %#v, want first %s", tt.input, got.IDs, tt.want)
		}
	}
}

func TestResolveUnknownVenue(t *testing.T) {
	_, err := Resolve(Filter{Expression: "not-a-venue"})
	if err == nil {
		t.Fatal("expected unknown venue error")
	}
}
