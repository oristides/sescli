package what

import (
	"strings"
	"testing"
)

func TestValidateRejectsUnknown(t *testing.T) {
	for _, token := range []string{"espetaculo", "palavera-teste"} {
		err := Validate(token)
		if err == nil {
			t.Fatalf("expected error for %q", token)
		}
		msg := err.Error()
		if !strings.Contains(msg, "invalid --what") || !strings.Contains(msg, "ALLOWED_PROFILE_VALUES") || !strings.Contains(msg, "ALLOWED_ACTIVITY_SLUGS") {
			t.Fatalf("error for %q should explain fix and list options: %s", token, msg)
		}
	}
}

func TestValidateAcceptsProfilesAndSlugs(t *testing.T) {
	for _, s := range []string{"", "cultural", "all", "teatro", "cinema", "shows-espetaculos-e-performances"} {
		if err := Validate(s); err != nil {
			t.Fatalf("%q: %v", s, err)
		}
	}
}

func TestValidateCommaList(t *testing.T) {
	if err := Validate("teatro,cinema"); err != nil {
		t.Fatal(err)
	}
	if err := Validate("cultural,cinema"); err == nil {
		t.Fatal("expected error mixing cultural with comma")
	}
}

func TestValidateSportsSynonymInCSV(t *testing.T) {
	if err := Validate("sports,cinema"); err != nil {
		t.Fatal(err)
	}
}
