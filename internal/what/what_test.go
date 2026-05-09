package what

import "testing"

func TestResolveProfilesAndActivities(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"cultural", []string{"teatro", "cinema", "tecnologias-e-artes"}},
		{"all", nil},
		{"cinema", []string{"cinema"}},
		{"teatro", []string{"teatro"}},
	}

	for _, tt := range tests {
		got := Resolve(Filter{Expression: tt.input, Audience: ""})
		if got.Audience != "adulto" {
			t.Fatalf("expected default adulto audience, got %q", got.Audience)
		}
		for _, want := range tt.want {
			found := false
			for _, got := range got.ActivityTypes {
				if got == want {
					found = true
				}
			}
			if !found {
				t.Fatalf("Resolve(%q) = %#v, missing %q", tt.input, got.ActivityTypes, want)
			}
		}
	}
}
