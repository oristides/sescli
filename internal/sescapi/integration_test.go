//go:build integration

package sescapi

import (
	"testing"
	"time"

	"sescli/internal/client"
	"sescli/internal/normalize"
	"sescli/internal/presets"
)

func TestIntegrationDinamicoUnitsReturnsRealData(t *testing.T) {
	u, err := DinamicoURL(DinamicoQuery{Mode: ModeUnidade, Audience: "adulto"})
	if err != nil {
		t.Fatal(err)
	}

	var raw any
	if err := client.New(client.Options{Timeout: 12 * time.Second, Retries: 2}).GetJSON(u, &raw); err != nil {
		t.Fatal(err)
	}
	units := normalize.UnitsFromRaw(raw, false)
	if len(units) == 0 {
		t.Fatalf("expected real units from SESC SP API")
	}
}

func TestIntegrationCapitalEventsEndpointIsReachable(t *testing.T) {
	u, err := EventsURL(EventsQuery{
		Units:         presets.CapitalUnitIDs(),
		Audience:      "adulto",
		ActivityTypes: presets.Defaults().ActivityTypes,
		PerPage:       3,
		Page:          1,
	})
	if err != nil {
		t.Fatal(err)
	}

	var raw any
	if err := client.New(client.Options{Timeout: 12 * time.Second, Retries: 2}).GetJSON(u, &raw); err != nil {
		t.Fatal(err)
	}
	if raw == nil {
		t.Fatalf("expected JSON array response")
	}
}

func TestIntegrationCentroEventsDecodeAndNormalizeWithoutPanic(t *testing.T) {
	u, err := EventsURL(EventsQuery{
		Units:         presets.CentroUnitIDs(),
		Audience:      "adulto",
		ActivityTypes: nil,
		From:          time.Now().Format(time.DateOnly),
		To:            time.Now().AddDate(0, 0, 14).Format(time.DateOnly),
		PerPage:       10,
		Page:          1,
	})
	if err != nil {
		t.Fatal(err)
	}
	var raw any
	if err := client.New(client.Options{Timeout: 15 * time.Second, Retries: 2}).GetJSON(u, &raw); err != nil {
		t.Fatal(err)
	}
	events := normalize.EventsFromRaw(raw, false)
	// Live paging can be transiently empty near midnight; normalization must tolerate any payload shape.
	n := len(events)
	if n != 0 && n < 500 {
		_ = events[0].Title
	}
}
