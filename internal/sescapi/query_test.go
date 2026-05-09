package sescapi

import (
	"net/url"
	"strings"
	"testing"

	"sescli/internal/presets"
)

func TestEventsURLUsesTypedQueryAndDefaults(t *testing.T) {
	defaults := presets.Defaults()
	u, err := EventsURL(EventsQuery{
		Units:         presets.CapitalUnitIDs(),
		Audience:      defaults.Audience,
		ActivityTypes: defaults.ActivityTypes,
		From:          "2026-05-08",
		To:            "2026-05-08",
		PerPage:       defaults.PerPage,
		Page:          defaults.Page,
	})
	if err != nil {
		t.Fatal(err)
	}

	parsed, err := url.Parse(u)
	if err != nil {
		t.Fatal(err)
	}
	values := parsed.Query()
	if parsed.Path != "/wp-json/wp/v1/atividades/filter" {
		t.Fatalf("unexpected path %q", parsed.Path)
	}
	if values.Get("publico") != "adulto" {
		t.Fatalf("expected adulto, got %q", values.Get("publico"))
	}
	if !strings.Contains(values.Get("local"), "761,2") {
		t.Fatalf("expected capital unit ids, got %q", values.Get("local"))
	}
	if values.Get("data_inicial") != "2026-05-08" || values.Get("data_final") != "2026-05-08" {
		t.Fatalf("expected date window, got %q to %q", values.Get("data_inicial"), values.Get("data_final"))
	}
	if values.Get("dinamico") != "true" || values.Get("tipo") != "atividade" {
		t.Fatalf("missing required defaults: %s", parsed.RawQuery)
	}
}

func TestDinamicoURLBuildsModeQueries(t *testing.T) {
	u, err := DinamicoURL(DinamicoQuery{
		Mode:     ModeUnidade,
		Audience: "adulto",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(u, "modes=unidade") || !strings.Contains(u, "publico_tag=adulto") {
		t.Fatalf("unexpected dinamico url %q", u)
	}
}
