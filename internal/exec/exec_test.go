package exec

import (
	"slices"
	"testing"
	"time"

	"sescli/internal/presets"
	"sescli/internal/sescapi"
	"sescli/internal/what"
)

func TestBuildQueryVariants(t *testing.T) {
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		t.Fatal(err)
	}
	base := time.Date(2026, 11, 4, 10, 0, 0, 0, loc)

	t.Run("when_range_dot_form", func(t *testing.T) {
		in := QueryInput{
			When:      "today..tomorrow",
			Preset:    "centro",
			Profile:   "cultural",
			Audience:  "adulto",
			PerPage:   40,
			Page:      2,
		}
		q, err := BuildQuery(in, base)
		if err != nil {
			t.Fatal(err)
		}
		got := q.EventsQuery()
		if got.From != "2026-11-04" || got.To != "2026-11-05" {
			t.Fatalf("from/to: %#v", got)
		}
		if got.Audience != "adulto" || got.PerPage != 40 || got.Page != 2 {
			t.Fatalf("metadata: %#v", got)
		}
		central, ok := presets.UnitIDsWithUrbanMacroZone(presets.ZoneCentral)
		if !ok {
			t.Fatal("zonacentral empty")
		}
		if !slices.Equal(got.Units, central) {
			t.Fatalf("centro units mismatch: %#v", got.Units)
		}
		exp := what.Resolve(what.Filter{Expression: "cultural", Audience: "adulto"})
		if !slices.Equal(got.ActivityTypes, exp.ActivityTypes) {
			t.Fatalf("activities: %+v vs %+v", got.ActivityTypes, exp.ActivityTypes)
		}
	})

	t.Run("explicit_from_to_dates", func(t *testing.T) {
		rm, ok := presets.UnitIDsWithUrbanMacroZone(presets.ZoneMetropolitana)
		if !ok || len(rm) < 2 {
			t.Fatalf("metropolitan zone ids: ok=%v got %v", ok, rm)
		}
		in := QueryInput{
			From:     "2026-06-01",
			To:       "2026-06-15",
			Where:    "metropolitana",
			What:     "all",
			Audience: "adulto",
			PerPage:  0,
			Page:     0,
		}
		q, err := BuildQuery(in, base)
		if err != nil {
			t.Fatal(err)
		}
		got := q.EventsQuery()
		exp := sescapi.EventsQuery{
			Units:         rm,
			Audience:      "adulto",
			ActivityTypes: nil,
			From:          "2026-06-01",
			To:            "2026-06-15",
			PerPage:       40,
			Page:          1,
		}
		if got.From != exp.From || got.To != exp.To || got.PerPage != exp.PerPage || got.Page != exp.Page {
			t.Fatalf("when/page: %+v vs %+v", got, exp)
		}
		if got.ActivityTypes != nil {
			t.Fatalf("expected cleared profile filter, got %+v", got.ActivityTypes)
		}
		if !slices.Equal(got.Units, exp.Units) {
			t.Fatal("metropolitana unit list differs")
		}
	})

	t.Run("implicit_today_centro", func(t *testing.T) {
		in := QueryInput{
			Where: "centro",
			What:  "all",
		}
		q, err := BuildQuery(in, base)
		if err != nil {
			t.Fatal(err)
		}
		if q.When.From != base.Format(time.DateOnly) || q.When.To != q.When.From {
			t.Fatalf("expected implicit today: %#v", q.When)
		}
	})

	t.Run("from_now_merges_with_when_today", func(t *testing.T) {
		in := QueryInput{
			When:    "today",
			FromNow: true,
			Where:   "centro",
			What:    "all",
		}
		q, err := BuildQuery(in, base)
		if err != nil || !q.When.FromNow {
			t.Fatalf("%#v %v", q.When, err)
		}
	})

	t.Run("explicit_unit_slug", func(t *testing.T) {
		in := QueryInput{
			When:      "tomorrow",
			Preset:    "",
			Where:     "",
			Profile:   "cultural",
			Audience:  "adulto",
			UnitNames: []string{"ipiranga"},
			PerPage:   5,
			Page:      3,
		}
		got, err := BuildQuery(in, base)
		if err != nil {
			t.Fatal(err)
		}
		ev := got.EventsQuery()
		found := slices.Contains(ev.Units, "56")
		if !found {
			t.Fatalf("expected Ipiranga 56 in %v", ev.Units)
		}
		if ev.From != "2026-11-05" || ev.To != "2026-11-05" {
			t.Fatalf("unexpected when: %+v", ev)
		}
		if ev.PerPage != 5 || ev.Page != 3 {
			t.Fatalf("pagination: %+v", ev)
		}
	})
}
