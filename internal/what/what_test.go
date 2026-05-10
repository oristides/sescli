package what

import "testing"

func TestResolveProfilesAndActivities(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"cultural", []string{"shows-espetaculos-e-performances", "cinema", "tecnologias-e-artes"}},
		{"all", nil},
		{"cinema", []string{"cinema"}},
		{"oficina", []string{"cursos-e-oficinas"}},
		{"teatro", []string{"shows-espetaculos-e-performances"}},
	}

	for _, tt := range tests {
		got := Resolve(Filter{Expression: tt.input, Audience: ""})
		if got.Audience != "adulto" {
			t.Fatalf("expected default adulto audience, got %q", got.Audience)
		}
		if tt.input == "teatro" && got.Language != "teatro" {
			t.Fatalf("Resolve(teatro) Language = %q, want teatro", got.Language)
		}
		if tt.input != "teatro" && got.Language != "" {
			t.Fatalf("Resolve(%q) Language = %q, want empty", tt.input, got.Language)
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
