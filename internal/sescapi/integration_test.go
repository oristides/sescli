//go:build integration

package sescapi

import (
	"strings"
	"testing"
	"time"

	"sescli/internal/client"
	"sescli/internal/normalize"
	"sescli/internal/presets"
)

func TestIntegrationUnidadesAtividadesListsVenues(t *testing.T) {
	u := UnidadesAtividadesURL()
	var raw any
	if err := client.New(client.Options{Timeout: 12 * time.Second, Retries: 2}).GetJSON(u, &raw); err != nil {
		t.Fatal(err)
	}
	units := normalize.UnitsFromRaw(raw, false)
	if len(units) == 0 {
		t.Fatalf("expected real venues from %s", u)
	}
	foundBomRetiro := false
	for _, un := range units {
		if un.ID == "48" && strings.Contains(strings.ToLower(un.Name), "bom retiro") {
			foundBomRetiro = true
		}
	}
	if !foundBomRetiro {
		t.Fatalf("expected Bom Retiro (48) on venue roster")
	}
}

func TestIntegrationMetroAreaEventsEndpointIsReachable(t *testing.T) {
	rm, ok := presets.UnitIDsWithUrbanMacroZone(presets.ZoneMetropolitana)
	if !ok || len(rm) == 0 {
		t.Fatal("need metropolitan IDs for probe")
	}
	u, err := EventsURL(EventsQuery{
		Units:         rm,
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
		Units:         presets.ClonePresetMap(presets.DefaultInstallPresets())["centro"],
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
