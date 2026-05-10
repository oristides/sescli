package where

import (
	"fmt"
	"strings"

	"sescli/internal/presets"
)

// ValidOptionsHelp explains allowed --where values for humans and agents.
func ValidOptionsHelp() string {
	geo := strings.Join(presets.BuiltinWhereGeographyExamples(), ", ")
	return strings.TrimSpace(fmt.Sprintf(`WHAT --where IS
  --where selects which SESC units (venues) are queried. Resolution order:
  (1) a preset key from config.json "presets" (if the name matches),
  (2) geography / special labels below (zona-*, capital, interior, …,
      or centro/center/default → zona-central bucket),
  (3) one venue token: name, slug, or numeric unit ID from the baked-in roster.

  Arbitrary place names or user-invented strings are rejected.

QUICK CHOICE
  Municipal union (most city units):  --where capital
  Central macro bucket:               --where centro
  One macro zone:                     --where zona-sul
  Single venue:                       --where ipiranga

ALLOWED_GEOGRAPHY_AND_SPECIAL_LABELS (hyphen/spacing variants often work — see WHERE.md):
  %s

CONFIG PRESETS
  --where <key> uses your config.json "presets" entry when the key exists.
  sescli config presets list

VENUE ROSTER (name, slug, or id)
  sescli info venues search "<needle>"
  sescli info venues --format pretty

DOCS: skills/references/WHERE.md`,
		geo,
	))
}

func invalidExpressionErr(expr string) error {
	e := strings.TrimSpace(expr)
	return fmt.Errorf("invalid --where %q: no matching config preset, geography label, or venue in the local list.\n\n%s", e, ValidOptionsHelp())
}
