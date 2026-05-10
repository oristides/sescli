package cli

import (
	"context"
	"sort"
	"time"

	"sescli/internal/normalize"
	"sescli/internal/presets"
	querymodel "sescli/internal/query"
)

type minResultsGather struct {
	Events          []normalize.Event
	Source          string
	ReportedTotal   *int
	EffectiveDateTo string
	WidenedDate     bool
	WidenedWhere    bool
	MinTarget       int
}

func saoPaulo() *time.Location {
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		return time.FixedZone("America/Sao_Paulo", -3*60*60)
	}
	return loc
}

func eventDedupeKey(e normalize.Event) string {
	if e.ID != "" {
		return "id:" + e.ID
	}
	if e.URL != "" {
		return "url:" + e.URL
	}
	return "x:" + e.Title + "|" + e.DateStart + "|" + e.Venue
}

func mergeEventMap(dst map[string]normalize.Event, list []normalize.Event) {
	for _, e := range list {
		dst[eventDedupeKey(e)] = e
	}
}

func sortedEventsFromMap(m map[string]normalize.Event) []normalize.Event {
	out := make([]normalize.Event, 0, len(m))
	for _, e := range m {
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].DateStart != out[j].DateStart {
			return out[i].DateStart < out[j].DateStart
		}
		if out[i].TimeStart != out[j].TimeStart {
			return out[i].TimeStart < out[j].TimeStart
		}
		return out[i].Title < out[j].Title
	})
	return out
}

// gatherEventsWithMinResults runs the primary fetch and, when minTarget > 0 and
// page is 1, may refetch with a wider calendar window and finally all
// municipal units (capital) until enough distinct events are collected or caps hit.
func gatherEventsWithMinResults(
	ctx context.Context,
	fetch EventFetcher,
	dq querymodel.Query,
	minTarget int,
) (minResultsGather, error) {
	api := dq.EventsQuery()
	out, err := fetch(ctx, api)
	if err != nil {
		return minResultsGather{}, err
	}

	limit := api.PerPage
	if limit <= 0 {
		limit = 40
	}
	page := api.Page
	if page <= 0 {
		page = 1
	}

	if minTarget <= 0 || page > 1 {
		events := out.Events
		return minResultsGather{
			Events:          events,
			Source:          out.Source,
			ReportedTotal:   out.ReportedTotal,
			EffectiveDateTo: dq.When.To,
			MinTarget:       minTarget,
		}, nil
	}

	merged := map[string]normalize.Event{}
	mergeEventMap(merged, out.Events)
	widenedDate := false
	widenedWhere := false
	effectiveTo := dq.When.To
	source := out.Source
	reported := out.ReportedTotal

	if len(merged) >= minTarget || len(merged) >= limit {
		events := sortedEventsFromMap(merged)
		if len(events) > limit {
			events = events[:limit]
		}
		return minResultsGather{
			Events:          events,
			Source:          source,
			ReportedTotal:   reported,
			EffectiveDateTo: effectiveTo,
			MinTarget:       minTarget,
		}, nil
	}

	loc := saoPaulo()
	origTo, err := time.ParseInLocation(time.DateOnly, dq.When.To, loc)
	if err != nil {
		events := sortedEventsFromMap(merged)
		if len(events) > limit {
			events = events[:limit]
		}
		return minResultsGather{
			Events:          events,
			Source:          source,
			ReportedTotal:   reported,
			EffectiveDateTo: effectiveTo,
			MinTarget:       minTarget,
		}, nil
	}

	dqWork := dq
	for add := 7; len(merged) < minTarget && add <= 28; add += 7 {
		newTo := origTo.AddDate(0, 0, add).Format(time.DateOnly)
		dqWork.When.To = newTo
		effectiveTo = newTo
		widenedDate = true
		reported = nil
		out2, err := fetch(ctx, dqWork.EventsQuery())
		if err != nil {
			break
		}
		source = out2.Source
		mergeEventMap(merged, out2.Events)
		if len(merged) >= minTarget {
			break
		}
	}

	if len(merged) < minTarget {
		capIDs, ok := presets.UnitIDsWithUrbanMacroZone("capital")
		if ok && len(capIDs) > 0 {
			dqCap := dqWork
			dqCap.Where.IDs = append([]string(nil), capIDs...)
			widenedWhere = true
			widenedDate = widenedDate || dqCap.When.To != dq.When.To
			reported = nil
			out3, err := fetch(ctx, dqCap.EventsQuery())
			if err == nil {
				source = out3.Source
				effectiveTo = dqCap.When.To
				mergeEventMap(merged, out3.Events)
			}
		}
	}

	events := sortedEventsFromMap(merged)
	if len(events) > limit {
		events = events[:limit]
	}
	return minResultsGather{
		Events:          events,
		Source:          source,
		ReportedTotal:   reported,
		EffectiveDateTo: effectiveTo,
		WidenedDate:     widenedDate,
		WidenedWhere:    widenedWhere,
		MinTarget:       minTarget,
	}, nil
}
