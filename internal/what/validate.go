package what

import (
	"fmt"
	"sort"
	"strings"
)

var allowedActivitySlug = buildAllowedActivitySlug()

func buildAllowedActivitySlug() map[string]struct{} {
	m := make(map[string]struct{})
	for _, s := range CulturalBundleSlugs {
		m[normalizeExprToken(s)] = struct{}{}
	}
	m[ShowsPerformancesActivity] = struct{}{}
	m["esporte-e-atividade-fisica"] = struct{}{}
	m["cursos-e-oficinas"] = struct{}{}
	for _, syn := range ExpressionSynonyms {
		m[syn] = struct{}{}
	}
	return m
}

func isProfileToken(lower string) bool {
	switch lower {
	case "all", "any", "todos", "todas", "cultural", "sports", "esportes", "teatro":
		return true
	default:
		return false
	}
}

// wholeExpressionProfileParts may not appear as one segment of a comma list.
func isWholeExpressionProfilePart(lower string) bool {
	switch lower {
	case "all", "any", "todos", "todas", "cultural":
		return true
	default:
		return false
	}
}

// NormalizeExprToken lowercases, trims, and applies ExpressionSynonyms.
func NormalizeExprToken(s string) string {
	return normalizeExprToken(s)
}

func normalizeExprToken(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return s
	}
	if rep, ok := ExpressionSynonyms[s]; ok {
		return rep
	}
	return s
}

// ValidOptionsHelp is appended to validation errors. Wording is tuned so humans
// and agents see rules, allowlists, and concrete examples in one place.
func ValidOptionsHelp() string {
	slugs := strings.Join(DocActivitySlugs(), ", ")
	profiles := strings.Join(ProfileTokens, ", ")
	return strings.TrimSpace(fmt.Sprintf(`WHAT THIS FLAG IS
  --what must be either (A) one profile keyword, or (B) one or more activity
  slugs from SESC, comma-separated. Invented words, translations, or guesses
  are always rejected.

QUICK CHOICE (common cases)
  Default cultural bundle:     --what cultural
  No activity filter (widest): --what all
  Plays / theater language only (API adds linguagem=teatro): --what teatro
  All "shows & performances" types (not only theater): --what shows-espetaculos-e-performances
  Cinema only:                  --what cinema
  Sports:                       --what sports   or   --what esportes
  Two slug filters together:    --what cinema,teatro   (each part must be a slug below; do not put "cultural" in a comma list)

VALID EXAMPLES (shape only; keep your own --when/--where)
  sescli ... --what cultural
  sescli ... --what all
  sescli ... --what teatro
  sescli ... --what shows-espetaculos-e-performances
  sescli ... --what cinema
  sescli ... --what cinema,teatro

ALLOWED_PROFILE_VALUES --exactly one of these as the entire --what value (no commas):
  %s

ALLOWED_ACTIVITY_SLUGS --use one slug alone, or several joined by commas; spell with hyphens exactly:
  %s

ALIASES (expanded before validation), e.g. oficina, sports:  sescli info what

DOCS: skills/references/WHAT.md`,
		profiles,
		slugs,
	))
}

// Validate returns an error if --what (or profile fallback) is not a known
// profile keyword, a comma-separated list of allowed activity slugs, or empty
// (empty means cultural downstream in Resolve).
func Validate(expression string) error {
	expr := strings.TrimSpace(expression)
	if expr == "" {
		return nil
	}
	lower := strings.ToLower(expr)
	if strings.Contains(expr, ",") {
		for _, part := range splitCSV(expr) {
			raw := strings.ToLower(strings.TrimSpace(part))
			t := normalizeExprToken(part)
			if t == "" {
				return fmt.Errorf("invalid --what: empty item in %q (remove extra commas).\n\n%s", expr, ValidOptionsHelp())
			}
			if isWholeExpressionProfilePart(t) || isWholeExpressionProfilePart(raw) {
				return fmt.Errorf("invalid --what: %q is a profile keyword and cannot be part of a comma-separated list. Use %q alone, or use only slugs with commas.\n\n%s", t, t, ValidOptionsHelp())
			}
			if _, ok := allowedActivitySlug[t]; !ok {
				return unknownWhatErr(t)
			}
		}
		return nil
	}
	if isProfileToken(lower) {
		return nil
	}
	t := normalizeExprToken(expr)
	if _, ok := allowedActivitySlug[t]; ok {
		return nil
	}
	return unknownWhatErr(lower)
}

func unknownWhatErr(token string) error {
	intro := fmt.Sprintf("invalid --what %q: that string is not in ALLOWED_PROFILE_VALUES or ALLOWED_ACTIVITY_SLUGS below. --what does not accept free text; pick a token from the lists or run: sescli info what\n\n", token)
	if hint, ok := TypoHints[token]; ok {
		return fmt.Errorf(intro+"Close match in this tool: %q\n\n%s", hint, ValidOptionsHelp())
	}
	return fmt.Errorf(intro+"%s", ValidOptionsHelp())
}

// DocActivitySlugs returns sorted unique API slugs users may pass (after synonyms).
func DocActivitySlugs() []string {
	out := make([]string, 0, len(allowedActivitySlug))
	for s := range allowedActivitySlug {
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}
