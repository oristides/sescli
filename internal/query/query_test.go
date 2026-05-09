package query

import (
	"testing"

	"sescli/internal/what"
	"sescli/internal/when"
	"sescli/internal/where"
)

func TestQueryConvertsToEventsQuery(t *testing.T) {
	q := Query{
		When: when.Filter{From: "2026-05-09", To: "2026-05-09"},
		Where: where.Resolved{
			IDs: []string{"56"},
		},
		What: what.Resolved{
			Audience:      "adulto",
			ActivityTypes: []string{"cinema"},
		},
		Page: PageOptions{Limit: 20, Page: 2},
	}

	got := q.EventsQuery()
	if got.From != "2026-05-09" || got.To != "2026-05-09" {
		t.Fatalf("unexpected when conversion: %#v", got)
	}
	if len(got.Units) != 1 || got.Units[0] != "56" {
		t.Fatalf("unexpected where conversion: %#v", got.Units)
	}
	if got.Audience != "adulto" || len(got.ActivityTypes) != 1 || got.ActivityTypes[0] != "cinema" {
		t.Fatalf("unexpected what conversion: %#v", got)
	}
	if got.PerPage != 20 || got.Page != 2 {
		t.Fatalf("unexpected page conversion: %#v", got)
	}
}

func TestEventsQueryPagingDefaultsIgnoreZeros(t *testing.T) {
	q := Query{
		When:  when.Filter{From: "2026-06-02", To: "2026-06-04"},
		Where: where.Resolved{IDs: []string{"2"}},
		What: what.Resolved{
			Audience:      "infantil",
			ActivityTypes: []string{"cinema"},
		},
		Page: PageOptions{Limit: 0, Page: 0},
	}
	got := q.EventsQuery()
	if got.PerPage != 40 || got.Page != 1 {
		t.Fatalf("pagination defaults: %#v", got)
	}
}
