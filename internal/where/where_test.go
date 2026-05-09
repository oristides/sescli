package where

import "testing"

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
