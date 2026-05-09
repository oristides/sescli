package presets

import "testing"

func TestResolveUnitQueriesFindsNamesSlugsIDsAndTypos(t *testing.T) {
	tests := []struct {
		query string
		want  string
	}{
		{"ipiranga", "56"},
		{"iporanga", "56"},
		{"sao caetano", "65"},
		{"São Caetano", "65"},
		{"52", "52"},
	}

	for _, tt := range tests {
		got, err := ResolveUnitIDs([]string{tt.query})
		if err != nil {
			t.Fatalf("ResolveUnitIDs(%q): %v", tt.query, err)
		}
		if len(got) != 1 || got[0] != tt.want {
			t.Fatalf("ResolveUnitIDs(%q) = %#v, want [%s]", tt.query, got, tt.want)
		}
	}
}

func TestSearchUnitsReturnsCodeAndName(t *testing.T) {
	matches := SearchUnits("ipiranga")
	if len(matches) == 0 {
		t.Fatalf("expected ipiranga match")
	}
	if matches[0].ID != "56" || matches[0].Name != "Ipiranga" {
		t.Fatalf("unexpected match: %#v", matches[0])
	}
}
