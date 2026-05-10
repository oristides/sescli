package programacao

import (
	"testing"

	"sescli/internal/normalize"
)

func TestSlugFromUserArg(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"nadine-2", "nadine-2"},
		{"/programacao/nadine-2", "nadine-2"},
		{"https://www.sescsp.org.br/programacao/nadine-2/", "nadine-2"},
		{"https://www.sescsp.org.br/programacao/nadine-2/extra", "nadine-2"},
		{"/programacao/foo?x=1", "foo"},
	}
	for _, tc := range cases {
		if got := SlugFromUserArg(tc.in); got != tc.want {
			t.Fatalf("SlugFromUserArg(%q) = %q want %q", tc.in, got, tc.want)
		}
	}
}

func TestPickEventByProgramSlug(t *testing.T) {
	events := []normalize.Event{
		{Title: "Other", URL: "https://www.sescsp.org.br/programacao/other"},
		{Title: "Nadine", URL: "https://www.sescsp.org.br/programacao/nadine-2"},
	}
	e, ok := PickEventByProgramSlug(events, "nadine-2")
	if !ok || e.Title != "Nadine" {
		t.Fatalf("pick: %#v ok=%v", e, ok)
	}
	_, ok = PickEventByProgramSlug(events, "missing")
	if ok {
		t.Fatal("expected no match")
	}
}
