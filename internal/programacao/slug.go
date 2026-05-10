package programacao

import (
	"strings"

	"sescli/internal/normalize"
)

// SlugFromUserArg extracts the program slug from a bare slug, a /programacao/… path, or a full sescsp URL.
func SlugFromUserArg(arg string) string {
	s := strings.TrimSpace(arg)
	if s == "" {
		return ""
	}
	const needle = "sescsp.org.br/programacao/"
	lower := strings.ToLower(s)
	if i := strings.Index(lower, needle); i >= 0 {
		s = s[i+len(needle):]
	} else if strings.HasPrefix(lower, "/programacao/") {
		s = s[len("/programacao/"):]
	}
	s = strings.Trim(s, `/`)
	if i := strings.IndexByte(s, '/'); i >= 0 {
		s = s[:i]
	}
	s = strings.TrimSpace(s)
	if i := strings.IndexByte(s, '?'); i >= 0 {
		s = s[:i]
	}
	return strings.TrimSpace(s)
}

// PickEventByProgramSlug returns the list row whose canonical URL ends with /programacao/{slug}.
func PickEventByProgramSlug(events []normalize.Event, slug string) (normalize.Event, bool) {
	slug = strings.Trim(strings.TrimSpace(slug), "/")
	if slug == "" {
		return normalize.Event{}, false
	}
	suffix := "/programacao/" + slug
	for _, e := range events {
		u := strings.TrimSuffix(e.URL, "/")
		if strings.HasSuffix(strings.ToLower(u), strings.ToLower(suffix)) {
			return e, true
		}
	}
	return normalize.Event{}, false
}
