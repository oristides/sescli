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

	"sescli/internal/config"
	"sescli/internal/normalize"
	"sescli/internal/sescapi"
)

func TestCentroTodayDefaultsToCompactJSON(t *testing.T) {
	app := App{FetchEvents: func(ctx context.Context, q sescapi.EventsQuery) (EventFetch, error) {
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
		return EventFetch{Events: []normalize.Event{{ID: "1", Title: "Cinema", URL: "https://sescsp.org.br/e"}}, Source: "mock://events"}, nil
	}}
	var stdout, stderr bytes.Buffer

	err := app.Execute(context.Background(), []string{"centro", "today"}, &stdout, &stderr)
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
	if !strings.Contains(string(b), `"default_preset": "capital"`) {
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
	app := App{FetchEvents: func(ctx context.Context, q sescapi.EventsQuery) (EventFetch, error) {
		if len(q.Units) != 1 || q.Units[0] != "56" {
			t.Fatalf("expected ipiranga to resolve to unit 56, got %#v", q.Units)
		}
		return EventFetch{Events: []normalize.Event{{Title: "Ipi"}}}, nil
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
		t.Fatal("expected unknown venue error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "invalid --where") || !strings.Contains(msg, "ALLOWED_GEOGRAPHY_AND_SPECIAL_LABELS") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCentroTodayWhatsAppFormat(t *testing.T) {
	app := App{FetchEvents: func(ctx context.Context, q sescapi.EventsQuery) (EventFetch, error) {
		return EventFetch{Events: []normalize.Event{{Title: "Teatro", URL: "https://sescsp.org.br/t", Venue: "Sesc 24 de Maio", PriceLabel: "R$ 20"}}}, nil
	}}
	var stdout, stderr bytes.Buffer

	err := app.Execute(context.Background(), []string{"centro", "today", "--format", "whatsapp"}, &stdout, &stderr)
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

func TestCentroTodayAcceptsLimitFlag(t *testing.T) {
	app := App{FetchEvents: func(ctx context.Context, q sescapi.EventsQuery) (EventFetch, error) {
		if q.PerPage != 10 {
			t.Fatalf("expected shortcut --limit to set ppp=10, got %d", q.PerPage)
		}
		return EventFetch{Events: []normalize.Event{{Title: "A"}}}, nil
	}}
	var stdout, stderr bytes.Buffer

	err := app.Execute(context.Background(), []string{"centro", "today", "--format", "whatsapp", "--limit", "10"}, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
}

func TestTomorrowTopLevelAcceptsWhereArgument(t *testing.T) {
	app := App{
		Now: func() time.Time {
			return time.Date(2026, 5, 8, 20, 16, 0, 0, time.FixedZone("BRT", -3*60*60))
		},
		FetchEvents: func(ctx context.Context, q sescapi.EventsQuery) (EventFetch, error) {
			if q.From != "2026-05-09" || q.To != "2026-05-09" {
				t.Fatalf("expected tomorrow date window, got %s to %s", q.From, q.To)
			}
			if len(q.Units) != 1 || q.Units[0] != "56" {
				t.Fatalf("expected ipiranga venue lookup, got %#v", q.Units)
			}
			if q.PerPage != 15 {
				t.Fatalf("expected limit 15, got %d", q.PerPage)
			}
			return EventFetch{Events: []normalize.Event{{Title: "Ipi"}}}, nil
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
		FetchEvents: func(ctx context.Context, q sescapi.EventsQuery) (EventFetch, error) {
			if q.From != "2026-05-09" || q.To != "2026-05-09" {
				t.Fatalf("expected tomorrow date window, got %s to %s", q.From, q.To)
			}
			if len(q.Units) != 1 || q.Units[0] != "56" {
				t.Fatalf("expected ipiranga venue lookup, got %#v", q.Units)
			}
			if q.PerPage != 20 {
				t.Fatalf("expected limit 20, got %d", q.PerPage)
			}
			return EventFetch{Events: []normalize.Event{{Title: "Ipi"}}}, nil
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
		FetchEvents: func(ctx context.Context, q sescapi.EventsQuery) (EventFetch, error) {
			if q.From != "2026-05-09" || q.To != "2026-05-10" {
				t.Fatalf("expected tomorrow to sunday range, got %s to %s", q.From, q.To)
			}
			if len(q.ActivityTypes) != 1 || q.ActivityTypes[0] != "shows-espetaculos-e-performances" || q.Language != "teatro" {
				t.Fatalf("expected what teatro → shows + linguagem teatro, got %#v lang=%q", q.ActivityTypes, q.Language)
			}
			return EventFetch{Events: []normalize.Event{{Title: "Teatro"}}}, nil
		},
	}
	var stdout, stderr bytes.Buffer

	err := app.Execute(context.Background(), []string{"--from", "tomorrow", "--to", "sunday", "--where", "centro", "--what", "teatro"}, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
}

func TestInvalidWhatRejected(t *testing.T) {
	app := App{
		Now: func() time.Time {
			return time.Date(2026, 5, 8, 20, 16, 0, 0, time.FixedZone("BRT", -3*60*60))
		},
	}
	var stdout, stderr bytes.Buffer
	err := app.Execute(context.Background(), []string{"--when", "next wednesday", "--what", "espetaculo", "--format", "whatsapp"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for invalid --what")
	}
	msg := err.Error()
	if !strings.Contains(msg, "invalid --what") || !strings.Contains(msg, "ALLOWED_PROFILE_VALUES") {
		t.Fatalf("expected invalid --what error with guidance, got %v", err)
	}
}

func TestInvalidWhereRejected(t *testing.T) {
	app := App{
		Now: func() time.Time {
			return time.Date(2026, 5, 8, 20, 16, 0, 0, time.FixedZone("BRT", -3*60*60))
		},
	}
	var stdout, stderr bytes.Buffer
	err := app.Execute(context.Background(), []string{"--when", "next wednesday", "--where", "palavra-controle", "--format", "whatsapp"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for invalid --where")
	}
	msg := err.Error()
	if !strings.Contains(msg, "invalid --where") || !strings.Contains(msg, "ALLOWED_GEOGRAPHY_AND_SPECIAL_LABELS") {
		t.Fatalf("expected invalid --where error with guidance, got %v", err)
	}
}

func TestShortFormTomorrowCentroWhatCultural(t *testing.T) {
	app := App{
		Now: func() time.Time {
			return time.Date(2026, 5, 8, 20, 16, 0, 0, time.FixedZone("BRT", -3*60*60))
		},
		FetchEvents: func(ctx context.Context, q sescapi.EventsQuery) (EventFetch, error) {
			if q.From != "2026-05-09" || q.To != "2026-05-09" {
				t.Fatalf("expected tomorrow date, got %s to %s", q.From, q.To)
			}
			if len(q.ActivityTypes) < 3 {
				t.Fatalf("expected cultural activity bundle, got %#v", q.ActivityTypes)
			}
			return EventFetch{Events: []normalize.Event{{Title: "Cultural"}}}, nil
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
		FetchEvents: func(ctx context.Context, q sescapi.EventsQuery) (EventFetch, error) {
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
			return EventFetch{Events: []normalize.Event{{Title: "Centro"}}}, nil
		},
	}
	var stdout, stderr bytes.Buffer

	err := app.Execute(context.Background(), []string{"centro", "tomorrow", "--format", "whatsapp", "--limit", "15"}, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
}

func TestWhatTeatroWithoutWhereUsesCapitalUnion(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "config.json")
	cfg := config.Default()
	b, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cfgPath, b, 0o644); err != nil {
		t.Fatal(err)
	}
	app := App{
		Now: func() time.Time {
			return time.Date(2026, 5, 9, 12, 0, 0, 0, time.FixedZone("BRT", -3*60*60))
		},
		FetchEvents: func(ctx context.Context, q sescapi.EventsQuery) (EventFetch, error) {
			if len(q.Units) < 12 {
				t.Fatalf("expected broad default unit list, got %d: %#v", len(q.Units), q.Units)
			}
			seen := map[string]bool{}
			for _, id := range q.Units {
				seen[id] = true
			}
			for _, id := range []string{"60", "56"} {
				if !seen[id] {
					t.Fatalf("expected default query to include municipal units (e.g. Pinheiros 60, Vila Mariana area 56); missing %s in %#v", id, q.Units)
				}
			}
			if q.Language != "teatro" {
				t.Fatalf("expected linguagem teatro for --what teatro, got %q", q.Language)
			}
			return EventFetch{Events: []normalize.Event{{Title: "A"}, {Title: "B"}}}, nil
		},
	}
	var stdout, stderr bytes.Buffer
	err = app.Execute(context.Background(), []string{"--config", cfgPath, "--when", "next wednesday", "--what", "teatro", "--format", "json"}, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "A") {
		t.Fatalf("expected events in output: %s", stdout.String())
	}
}

func TestVenueAliasesWork(t *testing.T) {
	app := App{FetchEvents: func(ctx context.Context, q sescapi.EventsQuery) (EventFetch, error) {
		if len(q.Units) != 1 || q.Units[0] != "56" {
			t.Fatalf("expected venue alias to resolve ipiranga, got %#v", q.Units)
		}
		return EventFetch{Events: []normalize.Event{{Title: "Ipi"}}}, nil
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
	for _, hidden := range []string{
		"centro     Use the central-unit default preset",
		"\n  today      ",
		"\n  tomorrow   ",
		"\n  events     ",
		"\n  facets     ",
		"\n  venues     ",
		"\n  completion ",
	} {
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

func TestInfoHelpListsEventAndWhatSubcommands(t *testing.T) {
	app := App{}
	var stdout, stderr bytes.Buffer
	err := app.Execute(context.Background(), []string{"info", "--help"}, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
	h := stderr.String()
	for _, needle := range []string{"event", "what", "venues", "facets"} {
		if !strings.Contains(h, needle) {
			t.Fatalf("info --help should mention %q subcommand; got:\n%s", needle, h)
		}
	}
	if !strings.Contains(h, "sescli info event") {
		t.Fatalf("info --help should show an event example; got:\n%s", h)
	}
	if !strings.Contains(h, "program slug") || !strings.Contains(h, "go install") {
		t.Fatalf("info --help should explain slug and stale-binary hint; got:\n%s", h)
	}

	stdout.Reset()
	stderr.Reset()
	err = app.Execute(context.Background(), []string{"info", "event", "--help"}, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
	eh := stderr.String()
	if !strings.Contains(eh, "SLUG_OR_URL identifies") || !strings.Contains(eh, "atividades/search") {
		t.Fatalf("info event --help should explain the argument and API behavior; got:\n%s", eh)
	}
	if !strings.Contains(eh, "Examples:") || !strings.Contains(eh, "sescli info event") {
		t.Fatalf("info event --help should include Examples; got:\n%s", eh)
	}
	if !strings.Contains(eh, "Global Flags:") {
		t.Fatalf("info event --help should list global flags; got:\n%s", eh)
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
	app := App{FetchEvents: func(ctx context.Context, q sescapi.EventsQuery) (EventFetch, error) {
		if len(q.ActivityTypes) != 0 {
			t.Fatalf("expected profile all to remove activity filter, got %#v", q.ActivityTypes)
		}
		return EventFetch{Events: []normalize.Event{{Title: "A"}}}, nil
	}}
	var stdout, stderr bytes.Buffer

	err := app.Execute(context.Background(), []string{"centro", "today", "--profile", "all"}, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
}

func TestFromNowFiltersPastEvents(t *testing.T) {
	app := App{
		Now: func() time.Time {
			return time.Date(2026, 5, 8, 20, 16, 0, 0, time.FixedZone("BRT", -3*60*60))
		},
		FetchEvents: func(ctx context.Context, q sescapi.EventsQuery) (EventFetch, error) {
			return EventFetch{Events: []normalize.Event{
				{Title: "Past", DateStart: "2026-05-08T14:00"},
				{Title: "Future", DateStart: "2026-05-08T21:00"},
			}}, nil
		},
	}
	var stdout, stderr bytes.Buffer

	err := app.Execute(context.Background(), []string{"centro", "today", "--from-now"}, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(stdout.String(), "Past") || !strings.Contains(stdout.String(), "Future") {
		t.Fatalf("unexpected from-now output: %s", stdout.String())
	}
}

func TestTrimGoRunSentinel(t *testing.T) {
	got := trimGoRunSentinel([]string{"--", "info", "event", "x"})
	want := []string{"info", "event", "x"}
	if len(got) != len(want) {
		t.Fatalf("%v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("%v vs %v", got, want)
		}
	}
}

func TestInfoEventUsesMockFetcher(t *testing.T) {
	app := App{
		FetchProgramBySlug: func(ctx context.Context, slug string) (normalize.Event, string, error) {
			_ = ctx
			if slug != "x-1" {
				t.Fatalf("slug %q", slug)
			}
			return normalize.Event{
				Title:   "Act",
				URL:     "https://www.sescsp.org.br/programacao/x-1",
				Summary: "sub",
			}, "mock://s", nil
		},
	}
	var stdout, stderr bytes.Buffer

	err := app.Execute(context.Background(), []string{
		"info", "event", "https://www.sescsp.org.br/programacao/x-1/",
		"--no-synopsis",
		"--format", "json",
	}, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
	out := stdout.String()
	if !strings.Contains(out, `"title":"Act"`) || !strings.Contains(out, `"summary":"sub"`) || !strings.Contains(out, `"event"`) {
		t.Fatalf("unexpected output: %s", out)
	}
}

func TestInfoEventAppliesBilheteriaPricingWhenJavaIDSet(t *testing.T) {
	var pricingCalls int
	app := App{
		FetchProgramBySlug: func(ctx context.Context, slug string) (normalize.Event, string, error) {
			_ = ctx
			if slug != "foo" {
				t.Fatalf("slug %q", slug)
			}
			return normalize.Event{
				Title:  "Act",
				URL:    "https://www.sescsp.org.br/programacao/foo",
				JavaID: "253160",
			}, "mock://search", nil
		},
		FetchActivityPricing: func(ctx context.Context, javaID, referer string) (*normalize.EventPricing, error) {
			_ = ctx
			pricingCalls++
			if javaID != "253160" {
				t.Fatalf("javaID %q", javaID)
			}
			if referer != "https://www.sescsp.org.br/programacao/foo" {
				t.Fatalf("referer %q", referer)
			}
			return &normalize.EventPricing{
				Gratuito:         false,
				ValorInteira:     "R$ 40,00",
				ValorMeia:        "R$ 20,00",
				StatusIngresso:   "LIBERADO",
				QtdeIngressosWeb: 12,
			}, nil
		},
	}
	var stdout, stderr bytes.Buffer
	err := app.Execute(context.Background(), []string{
		"info", "event", "foo",
		"--no-synopsis",
		"--format", "json",
	}, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
	if pricingCalls != 1 {
		t.Fatalf("expected 1 pricing fetch, got %d", pricingCalls)
	}
	out := stdout.String()
	if !strings.Contains(out, `"id_java":"253160"`) {
		t.Fatalf("missing id_java: %s", out)
	}
	if !strings.Contains(out, `"valor_inteira":"R$ 40,00"`) || !strings.Contains(out, `"pricing"`) {
		t.Fatalf("missing pricing in output: %s", out)
	}
	if !strings.Contains(out, "Inteira R$ 40,00") || !strings.Contains(out, "Meia R$ 20,00") {
		t.Fatalf("expected aggregate price from bilheteria: %s", out)
	}
}

func TestInfoEventFetchesSynopsisByDefaultWithMock(t *testing.T) {
	app := App{
		FetchProgramBySlug: func(ctx context.Context, slug string) (normalize.Event, string, error) {
			_ = ctx
			return normalize.Event{
				Title: "Act",
				URL:   "https://www.sescsp.org.br/programacao/foo",
			}, "mock://s", nil
		},
		FetchProgramSynopsis: func(ctx context.Context, pageURL string) (string, error) {
			_ = ctx
			if pageURL == "" {
				t.Fatal("empty url")
			}
			return "PARAGRAPH ONE ABOUT THE SHOW.", nil
		},
	}
	var stdout, stderr bytes.Buffer

	err := app.Execute(context.Background(), []string{
		"info", "event", "foo",
		"--format", "json",
	}, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "PARAGRAPH ONE") || !strings.Contains(stdout.String(), `"synopsis"`) {
		t.Fatalf("missing synopsis: %s", stdout.String())
	}
}

func TestInfoEventFetchesSynopsisAndPricingTogether(t *testing.T) {
	var synopsisCalls, pricingCalls int
	app := App{
		FetchProgramBySlug: func(ctx context.Context, slug string) (normalize.Event, string, error) {
			_ = ctx
			return normalize.Event{
				Title:  "Act",
				URL:    "https://www.sescsp.org.br/programacao/foo",
				JavaID: "253160",
			}, "mock://s", nil
		},
		FetchProgramSynopsis: func(ctx context.Context, pageURL string) (string, error) {
			synopsisCalls++
			return "PARAGRAPH SYNOPSIS.", nil
		},
		FetchActivityPricing: func(ctx context.Context, javaID, referer string) (*normalize.EventPricing, error) {
			pricingCalls++
			return &normalize.EventPricing{ValorInteira: "R$ 10,00"}, nil
		},
	}
	var stdout, stderr bytes.Buffer
	err := app.Execute(context.Background(), []string{
		"info", "event", "foo",
		"--format", "json",
	}, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
	if synopsisCalls != 1 || pricingCalls != 1 {
		t.Fatalf("synopsis=%d pricing=%d", synopsisCalls, pricingCalls)
	}
	out := stdout.String()
	if !strings.Contains(out, "PARAGRAPH SYNOPSIS") || !strings.Contains(out, `"valor_inteira":"R$ 10,00"`) {
		t.Fatalf("expected synopsis and pricing: %s", out)
	}
}

func TestInfoEventNoSynopsisSkipsPageFetch(t *testing.T) {
	var synopsisCalls int
	app := App{
		FetchProgramBySlug: func(ctx context.Context, slug string) (normalize.Event, string, error) {
			_ = ctx
			return normalize.Event{Title: "Act", URL: "https://www.sescsp.org.br/programacao/foo"}, "m", nil
		},
		FetchProgramSynopsis: func(ctx context.Context, pageURL string) (string, error) {
			synopsisCalls++
			return "SHOULD NOT APPEAR", nil
		},
	}
	var stdout, stderr bytes.Buffer
	err := app.Execute(context.Background(), []string{
		"info", "event", "foo", "--no-synopsis", "--format", "json",
	}, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
	if synopsisCalls != 0 {
		t.Fatalf("synopsis fetch ran %d times", synopsisCalls)
	}
	if strings.Contains(stdout.String(), "SHOULD NOT APPEAR") || strings.Contains(stdout.String(), `"synopsis"`) {
		t.Fatalf("unexpected synopsis in output: %s", stdout.String())
	}
}
