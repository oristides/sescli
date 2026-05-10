package cli

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"sescli/internal/normalize"
	querymodel "sescli/internal/query"
	"sescli/internal/sescapi"
	"sescli/internal/what"
	"sescli/internal/when"
	"sescli/internal/where"
)

func TestGatherEventsWithMinResultsWidensThenStops(t *testing.T) {
	var nCalls int
	fetch := func(ctx context.Context, q sescapi.EventsQuery) (EventFetch, error) {
		nCalls++
		if nCalls == 1 {
			return EventFetch{
				Events: []normalize.Event{{Title: "A", ID: "1", DateStart: "2026-05-16"}},
			}, nil
		}
		var bulk []normalize.Event
		for i := range 15 {
			bulk = append(bulk, normalize.Event{
				Title:     fmt.Sprintf("E%d", i),
				ID:        strconv.Itoa(100 + i),
				DateStart: "2026-05-20",
			})
		}
		return EventFetch{Events: bulk}, nil
	}

	dq := querymodel.Query{
		When:  when.Filter{From: "2026-05-16", To: "2026-05-16"},
		Where: where.Resolved{IDs: []string{"47"}},
		What:  what.Resolve(what.Filter{Expression: "all", Audience: "adulto"}),
		Page:  querymodel.PageOptions{Limit: 50, Page: 1},
	}

	g, err := gatherEventsWithMinResults(context.Background(), fetch, dq, 10)
	if err != nil {
		t.Fatal(err)
	}
	if nCalls < 2 {
		t.Fatalf("expected widen fetch, calls=%d", nCalls)
	}
	if len(g.Events) < 10 {
		t.Fatalf("expected at least 10 events, got %d", len(g.Events))
	}
	if !g.WidenedDate {
		t.Fatal("expected WidenedDate")
	}
	if g.EffectiveDateTo == "2026-05-16" {
		t.Fatal("expected EffectiveDateTo extended")
	}
}

func TestGatherEventsWithMinResultsNoWidenWhenEnough(t *testing.T) {
	nCalls := 0
	fetch := func(ctx context.Context, q sescapi.EventsQuery) (EventFetch, error) {
		nCalls++
		var evs []normalize.Event
		for i := range 12 {
			evs = append(evs, normalize.Event{Title: strconv.Itoa(i), ID: strconv.Itoa(i)})
		}
		return EventFetch{Events: evs}, nil
	}
	dq := querymodel.Query{
		When:  when.Filter{From: "2026-05-16", To: "2026-05-16"},
		Where: where.Resolved{IDs: []string{"47"}},
		What:  what.Resolve(what.Filter{Expression: "all", Audience: "adulto"}),
		Page:  querymodel.PageOptions{Limit: 50, Page: 1},
	}
	g, err := gatherEventsWithMinResults(context.Background(), fetch, dq, 10)
	if err != nil {
		t.Fatal(err)
	}
	if nCalls != 1 {
		t.Fatalf("expected single fetch, got %d", nCalls)
	}
	if g.WidenedDate || g.WidenedWhere {
		t.Fatalf("unexpected widen: %+v", g)
	}
}
