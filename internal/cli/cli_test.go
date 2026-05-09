package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"sescli/internal/normalize"
	"sescli/internal/sescapi"
)

func TestCapitalTodayDefaultsToCompactJSON(t *testing.T) {
	app := App{FetchEvents: func(ctx context.Context, q sescapi.EventsQuery) ([]normalize.Event, string, error) {
		if q.Audience != "adulto" {
			t.Fatalf("expected adulto audience, got %q", q.Audience)
		}
		if q.From == "" || q.To == "" {
			t.Fatalf("today should set date window: %#v", q)
		}
		for _, far := range []string{"55", "65", "71"} {
			for _, got := range q.Units {
				if got == far {
					t.Fatalf("default query should use centro config, not far capital unit %s in %#v", far, q.Units)
				}
			}
		}
		if len(q.Units) < 5 {
			t.Fatalf("centro preset not applied: %#v", q.Units)
		}
		if len(q.ActivityTypes) < 2 {
			t.Fatalf("cultural profile not applied: %#v", q.ActivityTypes)
		}
		return []normalize.Event{{ID: "1", Title: "Cinema", URL: "https://sescsp.org.br/e"}}, "mock://events", nil
	}}
	var stdout, stderr bytes.Buffer

	err := app.Execute(context.Background(), []string{"capital", "today"}, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("unexpected stderr: %q", stderr.String())
	}
	if strings.Contains(stdout.String(), "\n") {
		t.Fatalf("default JSON should be compact, got %q", stdout.String())
	}

	var decoded map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &decoded); err != nil {
		t.Fatalf("stdout must be JSON: %v\n%s", err, stdout.String())
	}
	if _, ok := decoded["events"]; !ok {
		t.Fatalf("missing events key: %#v", decoded)
	}
}

func TestConfigInitWritesDefaultConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	app := App{}
	var stdout, stderr bytes.Buffer

	err := app.Execute(context.Background(), []string{"--config", path, "config", "init"}, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"default_preset": "centro"`) {
		t.Fatalf("expected centro config, got %s", string(b))
	}
	if !strings.Contains(stdout.String(), path) {
		t.Fatalf("expected config path in stdout, got %q", stdout.String())
	}
}

func TestConfigSetupCommandUsesMockedInput(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	app := App{Stdin: strings.NewReader("centro\nadulto\ncinema\nwhatsapp\n15\n")}
	var stdout, stderr bytes.Buffer

	err := app.Execute(context.Background(), []string{"--config", path, "config", "setup"}, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"what": "cinema"`) || !strings.Contains(string(b), `"format": "whatsapp"`) {
		t.Fatalf("unexpected setup file: %s", string(b))
	}
}

func TestEventsResolvesUnitNamesToIDs(t *testing.T) {
	app := App{FetchEvents: func(ctx context.Context, q sescapi.EventsQuery) ([]normalize.Event, string, error) {
		if len(q.Units) != 1 || q.Units[0] != "56" {
			t.Fatalf("expected ipiranga to resolve to unit 56, got %#v", q.Units)
		}
		return []normalize.Event{{Title: "Ipi"}}, "", nil
	}}
	var stdout, stderr bytes.Buffer

	err := app.Execute(context.Background(), []string{"events", "--unit", "ipiranga"}, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnitsSearchReturnsStaticIndex(t *testing.T) {
	app := App{}
	var stdout, stderr bytes.Buffer

	err := app.Execute(context.Background(), []string{"units", "search", "ipiranga"}, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), `"id":"56"`) || !strings.Contains(stdout.String(), "Ipiranga") {
		t.Fatalf("unexpected search output: %s", stdout.String())
	}
}

func TestEventsUnknownUnitNameReturnsHelpfulError(t *testing.T) {
	app := App{}
	var stdout, stderr bytes.Buffer

	err := app.Execute(context.Background(), []string{"events", "--unit", "not-a-real-unit"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected unknown unit error")
	}
	if !strings.Contains(err.Error(), "unknown unit") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCapitalTodayWhatsAppFormat(t *testing.T) {
	app := App{FetchEvents: func(ctx context.Context, q sescapi.EventsQuery) ([]normalize.Event, string, error) {
		return []normalize.Event{{Title: "Teatro", URL: "https://sescsp.org.br/t", Venue: "Sesc 24 de Maio", PriceLabel: "R$ 20"}}, "", nil
	}}
	var stdout, stderr bytes.Buffer

	err := app.Execute(context.Background(), []string{"capital", "today", "--format", "whatsapp"}, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
	out := stdout.String()
	if !strings.Contains(out, "Teatro") || !strings.Contains(out, "https://sescsp.org.br/t") {
		t.Fatalf("unexpected whatsapp output: %q", out)
	}
	if strings.Contains(out, "{") {
		t.Fatalf("whatsapp output should not be JSON: %q", out)
	}
}

func TestCapitalTodayAcceptsLimitFlag(t *testing.T) {
	app := App{FetchEvents: func(ctx context.Context, q sescapi.EventsQuery) ([]normalize.Event, string, error) {
		if q.PerPage != 10 {
			t.Fatalf("expected shortcut --limit to set ppp=10, got %d", q.PerPage)
		}
		return []normalize.Event{{Title: "A"}}, "", nil
	}}
	var stdout, stderr bytes.Buffer

	err := app.Execute(context.Background(), []string{"capital", "today", "--format", "whatsapp", "--limit", "10"}, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
}

func TestTomorrowTopLevelAcceptsWhereArgument(t *testing.T) {
	app := App{
		Now: func() time.Time {
			return time.Date(2026, 5, 8, 20, 16, 0, 0, time.FixedZone("BRT", -3*60*60))
		},
		FetchEvents: func(ctx context.Context, q sescapi.EventsQuery) ([]normalize.Event, string, error) {
			if q.From != "2026-05-09" || q.To != "2026-05-09" {
				t.Fatalf("expected tomorrow date window, got %s to %s", q.From, q.To)
			}
			if len(q.Units) != 1 || q.Units[0] != "56" {
				t.Fatalf("expected ipiranga venue lookup, got %#v", q.Units)
			}
			if q.PerPage != 15 {
				t.Fatalf("expected limit 15, got %d", q.PerPage)
			}
			return []normalize.Event{{Title: "Ipi"}}, "", nil
		},
	}
	var stdout, stderr bytes.Buffer

	err := app.Execute(context.Background(), []string{"tomorrow", "ipiranga", "--format", "whatsapp", "--limit", "15"}, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRootWhenWhereFlagsWork(t *testing.T) {
	app := App{
		Now: func() time.Time {
			return time.Date(2026, 5, 8, 20, 16, 0, 0, time.FixedZone("BRT", -3*60*60))
		},
		FetchEvents: func(ctx context.Context, q sescapi.EventsQuery) ([]normalize.Event, string, error) {
			if q.From != "2026-05-09" || q.To != "2026-05-09" {
				t.Fatalf("expected tomorrow date window, got %s to %s", q.From, q.To)
			}
			if len(q.Units) != 1 || q.Units[0] != "56" {
				t.Fatalf("expected ipiranga venue lookup, got %#v", q.Units)
			}
			if q.PerPage != 20 {
				t.Fatalf("expected limit 20, got %d", q.PerPage)
			}
			return []normalize.Event{{Title: "Ipi"}}, "", nil
		},
	}
	var stdout, stderr bytes.Buffer

	err := app.Execute(context.Background(), []string{"--when", "tomorrow", "--where", "ipiranga", "--format", "whatsapp", "--limit", "20"}, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRootFromToWhatFlagsWork(t *testing.T) {
	app := App{
		Now: func() time.Time {
			return time.Date(2026, 5, 8, 20, 16, 0, 0, time.FixedZone("BRT", -3*60*60))
		},
		FetchEvents: func(ctx context.Context, q sescapi.EventsQuery) ([]normalize.Event, string, error) {
			if q.From != "2026-05-09" || q.To != "2026-05-10" {
				t.Fatalf("expected tomorrow to sunday range, got %s to %s", q.From, q.To)
			}
			if len(q.ActivityTypes) != 1 || q.ActivityTypes[0] != "teatro" {
				t.Fatalf("expected what teatro, got %#v", q.ActivityTypes)
			}
			return []normalize.Event{{Title: "Teatro"}}, "", nil
		},
	}
	var stdout, stderr bytes.Buffer

	err := app.Execute(context.Background(), []string{"--from", "tomorrow", "--to", "sunday", "--where", "centro", "--what", "teatro"}, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
}

func TestShortFormTomorrowCentroWhatCultural(t *testing.T) {
	app := App{
		Now: func() time.Time {
			return time.Date(2026, 5, 8, 20, 16, 0, 0, time.FixedZone("BRT", -3*60*60))
		},
		FetchEvents: func(ctx context.Context, q sescapi.EventsQuery) ([]normalize.Event, string, error) {
			if q.From != "2026-05-09" || q.To != "2026-05-09" {
				t.Fatalf("expected tomorrow date, got %s to %s", q.From, q.To)
			}
			if len(q.ActivityTypes) < 3 {
				t.Fatalf("expected cultural activity bundle, got %#v", q.ActivityTypes)
			}
			return []normalize.Event{{Title: "Cultural"}}, "", nil
		},
	}
	var stdout, stderr bytes.Buffer

	err := app.Execute(context.Background(), []string{"tomorrow", "centro", "--what", "cultural", "--format", "whatsapp"}, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCentroTomorrowWorks(t *testing.T) {
	app := App{
		Now: func() time.Time {
			return time.Date(2026, 5, 8, 20, 16, 0, 0, time.FixedZone("BRT", -3*60*60))
		},
		FetchEvents: func(ctx context.Context, q sescapi.EventsQuery) ([]normalize.Event, string, error) {
			if q.From != "2026-05-09" || q.To != "2026-05-09" {
				t.Fatalf("expected tomorrow date window, got %s to %s", q.From, q.To)
			}
			for _, far := range []string{"55", "65", "71"} {
				for _, got := range q.Units {
					if got == far {
						t.Fatalf("centro tomorrow should not include far venue %s in %#v", far, q.Units)
					}
				}
			}
			return []normalize.Event{{Title: "Centro"}}, "", nil
		},
	}
	var stdout, stderr bytes.Buffer

	err := app.Execute(context.Background(), []string{"centro", "tomorrow", "--format", "whatsapp", "--limit", "15"}, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
}

func TestVenueAliasesWork(t *testing.T) {
	app := App{FetchEvents: func(ctx context.Context, q sescapi.EventsQuery) ([]normalize.Event, string, error) {
		if len(q.Units) != 1 || q.Units[0] != "56" {
			t.Fatalf("expected venue alias to resolve ipiranga, got %#v", q.Units)
		}
		return []normalize.Event{{Title: "Ipi"}}, "", nil
	}}
	var stdout, stderr bytes.Buffer

	err := app.Execute(context.Background(), []string{"events", "--venue", "ipiranga"}, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRootHelpShowsCanonicalQueryModelOnly(t *testing.T) {
	app := App{}
	var stdout, stderr bytes.Buffer

	err := app.Execute(context.Background(), []string{"--help"}, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
	help := stderr.String()
	for _, hidden := range []string{"\n  capital    ", "\n  centro     ", "\n  today      ", "\n  tomorrow   ", "\n  events     ", "\n  facets     ", "\n  venues     ", "\n  completion "} {
		if strings.Contains(help, hidden) {
			t.Fatalf("compatibility command %q should be hidden from root help:\n%s", hidden, help)
		}
	}
	for _, visible := range []string{"\n  config", "\n  info"} {
		if !strings.Contains(help, visible) {
			t.Fatalf("canonical command %q missing from help:\n%s", visible, help)
		}
	}
	for _, section := range []string{"WHEN", "WHERE", "WHAT", "OPTIONS"} {
		if !strings.Contains(help, section) {
			t.Fatalf("help should explain %s section:\n%s", section, help)
		}
	}
	for _, hiddenFlag := range []string{"--venue ", "--venues ", "--profile ", "--preset ", "--from ", "--to ", "--audience "} {
		if strings.Contains(help, hiddenFlag) {
			t.Fatalf("redundant flag %q should be hidden from root help:\n%s", hiddenFlag, help)
		}
	}
}

func TestInfoVenuesSearchWorks(t *testing.T) {
	app := App{}
	var stdout, stderr bytes.Buffer

	err := app.Execute(context.Background(), []string{"info", "venues", "search", "ipiranga"}, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), `"id":"56"`) || !strings.Contains(stdout.String(), "Ipiranga") {
		t.Fatalf("unexpected info venues output: %s", stdout.String())
	}
}

func TestProfileAllRemovesActivityFilter(t *testing.T) {
	app := App{FetchEvents: func(ctx context.Context, q sescapi.EventsQuery) ([]normalize.Event, string, error) {
		if len(q.ActivityTypes) != 0 {
			t.Fatalf("expected profile all to remove activity filter, got %#v", q.ActivityTypes)
		}
		return []normalize.Event{{Title: "A"}}, "", nil
	}}
	var stdout, stderr bytes.Buffer

	err := app.Execute(context.Background(), []string{"capital", "today", "--profile", "all"}, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
}

func TestFromNowFiltersPastEvents(t *testing.T) {
	app := App{
		Now: func() time.Time {
			return time.Date(2026, 5, 8, 20, 16, 0, 0, time.FixedZone("BRT", -3*60*60))
		},
		FetchEvents: func(ctx context.Context, q sescapi.EventsQuery) ([]normalize.Event, string, error) {
			return []normalize.Event{
				{Title: "Past", DateStart: "2026-05-08T14:00"},
				{Title: "Future", DateStart: "2026-05-08T21:00"},
			}, "", nil
		},
	}
	var stdout, stderr bytes.Buffer

	err := app.Execute(context.Background(), []string{"capital", "today", "--from-now"}, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(stdout.String(), "Past") || !strings.Contains(stdout.String(), "Future") {
		t.Fatalf("unexpected from-now output: %s", stdout.String())
	}
}
