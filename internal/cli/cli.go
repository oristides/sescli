package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/spf13/cobra"

	"sescli/internal/bilheteria"
	"sescli/internal/client"
	"sescli/internal/config"
	"sescli/internal/eventtime"
	"sescli/internal/exec"
	outfmt "sescli/internal/format"
	"sescli/internal/normalize"
	"sescli/internal/presets"
	"sescli/internal/programacao"
	"sescli/internal/programpage"
	querymodel "sescli/internal/query"
	"sescli/internal/sescapi"
	"sescli/internal/what"
)

// EventFetch is the outcome of an atividades/filter request.
type EventFetch struct {
	Events        []normalize.Event
	Source        string
	ReportedTotal *int
}

type EventFetcher func(context.Context, sescapi.EventsQuery) (EventFetch, error)
type UnitFetcher func(context.Context, sescapi.DinamicoQuery) ([]normalize.Unit, string, error)
type FacetFetcher func(context.Context, sescapi.DinamicoQuery) (any, string, error)

type App struct {
	FetchEvents          EventFetcher
	FetchUnits           UnitFetcher
	FetchFacets          FacetFetcher
	FetchProgramBySlug   func(ctx context.Context, slug string) (normalize.Event, string, error)
	FetchProgramSynopsis func(ctx context.Context, pageURL string) (string, error)
	FetchActivityPricing func(ctx context.Context, javaID, referer string) (*normalize.EventPricing, error)
	Now                  func() time.Time
	Stdin                io.Reader
}

type options struct {
	format       string
	includeRaw   bool
	units        []string
	unitNames    []string
	audience     string
	profile      string
	what         string
	when         string
	where        string
	from         string
	to           string
	perPage      int
	page         int
	fromNow      bool
	preset       string
	configPath   string
	force        bool
	cfg          config.Config
	minResults   int
	summaryChars int
}

func (a App) Execute(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	opts := defaultOptions()
	cmd := a.rootCommand(ctx, stdout, stderr, &opts)
	cmd.SetArgs(args)
	return cmd.Execute()
}

func (a App) rootCommand(ctx context.Context, stdout, stderr io.Writer, opts *options) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "sescli",
		Short:         "Fetch succinct SESC SP event information",
		Long:          rootHelpText(),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.when == "" && opts.where == "" {
				return cmd.Help()
			}
			if opts.where != "" {
				opts.applyWhere(opts.where)
			}
			dq, err := opts.domainQuery(a.now())
			if err != nil {
				return err
			}
			return a.printEvents(ctx, stdout, dq, opts)
		},
	}
	cmd.SetOut(stderr)
	cmd.SetErr(stderr)
	cmd.CompletionOptions.DisableDefaultCmd = true
	setRootOnlyHelp(cmd)
	addGlobalFlags(cmd, opts)
	cmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		opts.loadConfig()
	}

	centroCmd := a.centroCommand(ctx, stdout, opts)
	centroCmd.Hidden = true
	cmd.AddCommand(centroCmd)
	todayCmd := a.dateCommand(ctx, stdout, opts, "today", 0)
	todayCmd.Hidden = true
	cmd.AddCommand(todayCmd)
	tomorrowCmd := a.dateCommand(ctx, stdout, opts, "tomorrow", 1)
	tomorrowCmd.Hidden = true
	cmd.AddCommand(tomorrowCmd)
	eventsCmd := a.eventsCommand(ctx, stdout, opts)
	eventsCmd.Hidden = true
	cmd.AddCommand(eventsCmd)
	unitsCmd := a.unitsCommand(ctx, stdout, opts)
	unitsCmd.Hidden = true
	cmd.AddCommand(unitsCmd)
	venuesCmd := a.venuesCommand(ctx, stdout, opts)
	venuesCmd.Hidden = true
	cmd.AddCommand(venuesCmd)
	facetsCmd := a.facetsCommand(ctx, stdout, opts)
	facetsCmd.Hidden = true
	cmd.AddCommand(facetsCmd)
	cmd.AddCommand(a.infoCommand(ctx, stdout, opts))
	cmd.AddCommand(a.configCommand(stdout, opts))
	return cmd
}

func rootHelpText() string {
	return `Fetch succinct SESC SP event information.

Primary structure:
  sescli --when <date> --where <venue/preset> --what <activity/profile> [options]

Examples:
  sescli --when tomorrow --where ipiranga --what cinema --format whatsapp --limit 20
  sescli --when "from tomorrow to sunday" --where centro --what teatro
  sescli --when today --where "zona-sul" --what all --limit 15

WHEN:
  today, tomorrow, from-now, YYYY-MM-DD, weekend, next-weekend,
  "from tomorrow to sunday", "next friday"

WHERE:
  centro     — preset named centro (zonacentral seed in config.json; edits are yours)
  zona-norte, zona-sul, zona-leste, zona-oeste, zona-central — all mapped units in that heuristic bucket
  metropolitana, interior, litoral — state / RM buckets
  or a venue name/slug/id: ipiranga, cinesesc, 56

WHAT:
  Profiles: cultural, all, teatro, sports | esportes, any | todos | todas
  Or comma-separated slugs (sescli info what lists allowed API values).
  Typos like "espetaculo" are rejected with a hint; use shows-espetaculos-e-performances or teatro.

OPTIONS:
  --format json|whatsapp|pretty|table
  --limit N (page size / max rows returned after de-duplication)
  --min-results N  Page 1 only: if fewer than N distinct events, widen end date (+7..+28d) then municipal units (capital) until N or caps. Output still capped by --limit.
  --summary-chars N  Max length for JSON event summaries (0 = no truncation). Default 220.
  --page N` + rootHelpTextInfo()
}

func rootHelpTextInfo() string {
	return `

INFO:
  sescli info event SLUG_OR_URL   Program page: API row + HTML synopsis + pricing when id_java present; --no-synopsis skips HTML only.
  Rebuild (go install ./cmd/sescli) if "sescli info --help" does not list event and what subcommands`
}

func rootHelpTemplate() string {
	return `{{.Long}}

Usage:
  {{.UseLine}}

Commands:
{{range .Commands}}{{if (and (not .Hidden) (ne .Name "help"))}}  {{rpad .Name .NamePadding }} {{.Short}}
{{end}}{{end}}  help       Help about any command

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}
`
}

// setRootOnlyHelp installs help rendering that keeps the custom root layout while letting
// subcommands use Cobra's default help so Examples, usage, and global flags show correctly.
func setRootOnlyHelp(root *cobra.Command) {
	root.SetHelpFunc(func(c *cobra.Command, args []string) {
		out := c.OutOrStdout()
		if c.Root() != c {
			// Subcommand: same as cobra.defaultHelpFunc (not the narrow root template).
			usage := strings.TrimSpace(c.Long)
			if usage == "" {
				usage = strings.TrimSpace(c.Short)
			}
			if usage != "" {
				fmt.Fprintln(out, usage)
				fmt.Fprintln(out)
			}
			if c.Runnable() || c.HasSubCommands() {
				fmt.Fprint(out, c.UsageString())
			}
			return
		}
		t := template.New("sescli-root-help")
		t.Funcs(template.FuncMap{
			"trimTrailingWhitespaces": strings.TrimSpace,
			"rpad": func(s string, width int) string {
				if width <= 0 {
					return s
				}
				return fmt.Sprintf("%-*s", width, s)
			},
		})
		template.Must(t.Parse(rootHelpTemplate()))
		if err := t.Execute(out, c); err != nil {
			c.PrintErrln(err)
		}
	})
}

func (a App) centroCommand(ctx context.Context, stdout io.Writer, opts *options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "centro",
		Short: "Use the central-unit default preset",
	}
	cmd.AddCommand(a.dateCommand(ctx, stdout, opts, "today", 0))
	cmd.AddCommand(a.dateCommand(ctx, stdout, opts, "tomorrow", 1))
	return cmd
}

func (a App) dateCommand(ctx context.Context, stdout io.Writer, opts *options, name string, offsetDays int) *cobra.Command {
	return &cobra.Command{
		Use:   name + " [where]",
		Short: "Events " + name + " using configured defaults or optional venue/preset",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// The hidden `centro today` / `centro tomorrow` shortcuts are namespaced
			// under the centro command — they must keep zona-central semantics even
			// though the default geography preset is `capital`.
			if cmd.Parent() != nil && cmd.Parent().Name() == "centro" {
				opts.preset = "centro"
				opts.where = ""
				opts.unitNames = nil
			}
			if len(args) == 1 {
				opts.applyWhere(args[0])
			}
			q, err := opts.domainQuery(a.now())
			if err != nil {
				return err
			}
			from, to := a.dateWindow(offsetDays)
			q.When.From = from
			q.When.To = to
			return a.printEvents(ctx, stdout, q, opts)
		},
	}
}

func (a App) eventsCommand(ctx context.Context, stdout io.Writer, opts *options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "events",
		Short: "Fetch SESC SP events with explicit filters",
		RunE: func(cmd *cobra.Command, args []string) error {
			dq, err := opts.domainQuery(a.now())
			if err != nil {
				return err
			}
			return a.printEvents(ctx, stdout, dq, opts)
		},
	}
	return cmd
}

func (a App) unitsCommand(ctx context.Context, stdout io.Writer, opts *options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "units",
		Short: "List SESC venues via /unidades-atividades (or dinamico fallback for other modes)",
		RunE: func(cmd *cobra.Command, args []string) error {
			fetch := a.FetchUnits
			if fetch == nil {
				fetch = realUnitFetcher(opts.includeRaw)
			}
			q := sescapi.DinamicoQuery{Mode: sescapi.ModeUnidade, Audience: opts.audience}
			units, source, err := fetch(ctx, q)
			if err != nil {
				return err
			}
			payload := outfmt.Response("units", units, outfmt.Meta{Total: len(units), Source: source, Query: "units"})
			return printPayload(stdout, opts.format, payload, nil, outfmt.Meta{})
		},
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "search QUERY",
		Short: "Search local unit index by name, slug, typo, or code",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			matches := presets.SearchUnits(args[0])
			payload := outfmt.Response("units", matches, outfmt.Meta{Total: len(matches), Query: "units:search"})
			return printPayload(stdout, opts.format, payload, nil, outfmt.Meta{})
		},
	})
	return cmd
}

func (a App) venuesCommand(ctx context.Context, stdout io.Writer, opts *options) *cobra.Command {
	cmd := a.unitsCommand(ctx, stdout, opts)
	cmd.Use = "venues"
	cmd.Short = "List/search SESC venues (unidades)"
	return cmd
}

func (a App) infoCommand(ctx context.Context, stdout io.Writer, opts *options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Look up one event by slug/URL, valid --what values, venues, or API facets",
		Long: `Subcommands:

  event   One programação activity: search row + synopsis (HTML) + ticket pricing when id_java exists
  what    Valid profile keywords, synonyms, and --what activity slugs
  venues  List or search SESC units (unidades)
  facets  Raw facet metadata from the dinamico API

The program slug is the path segment after /programacao/ on the public site (e.g. .../programacao/nadine-2/ → nadine-2).

Examples:
  sescli info event shows-espetaculos-e-performances/my-show --format json
  sescli info what
  sescli info venues search pinheiros

If "sescli info --help" does not list the event and what subcommands, rebuild from this repo: go install ./cmd/sescli`,
	}
	// Most specific / common first so they appear early in "sescli info --help".
	cmd.AddCommand(a.eventInfoCommand(ctx, stdout, opts))
	cmd.AddCommand(a.whatInfoCommand(stdout))
	venues := a.venuesCommand(ctx, stdout, opts)
	venues.Use = "venues"
	venues.Hidden = false
	cmd.AddCommand(venues)
	facets := a.facetsCommand(ctx, stdout, opts)
	facets.Use = "facets"
	facets.Hidden = false
	cmd.AddCommand(facets)
	return cmd
}

func (a App) whatInfoCommand(stdout io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "what",
		Short: "List valid --what profile keywords, synonyms, and activity slugs",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := fmt.Fprintf(stdout, "Profiles (single token; do not combine with commas):\n  %s\n\n", strings.Join(what.ProfileTokens, ", "))
			if err != nil {
				return err
			}
			_, err = fmt.Fprint(stdout, "Synonyms (accepted aliases → API slug):\n")
			if err != nil {
				return err
			}
			var keys []string
			for k := range what.ExpressionSynonyms {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				if _, e := fmt.Fprintf(stdout, "  %s → %s\n", k, what.ExpressionSynonyms[k]); e != nil {
					return e
				}
			}
			slugs := what.DocActivitySlugs()
			if _, err := fmt.Fprintf(stdout, "\nActivity slugs (single token or comma-separated):\n  %s\n", strings.Join(slugs, ", ")); err != nil {
				return err
			}
			_, err = fmt.Fprint(stdout, "\nNotes:\n  • teatro alone → shows bucket + linguagem=teatro. For all show types, use shows-espetaculos-e-performances.\n  • cultural (default) bundles several slugs; see CulturalBundleSlugs in the source.\n")
			return err
		},
	}
}

func (a App) eventInfoCommand(ctx context.Context, stdout io.Writer, opts *options) *cobra.Command {
	var apiOnly bool
	cmd := &cobra.Command{
		Use:   "event SLUG_OR_URL",
		Short: "One programação event: API metadata, HTML sinopse, and ticket pricing when id_java exists",
		Long: `Argument SLUG_OR_URL identifies one activity page under the site's programação section:

  • Program slug alone — the last segment of the URL path, e.g. "nadine-2" from:
      https://www.sescsp.org.br/programacao/nadine-2/
  • Full URL or path — same slug is extracted; quotes help in the shell if the URL has "&" or other characters.

What it does:
  1) GET /wp-json/wp/v1/atividades/search with that slug, then picks the result whose link matches /programacao/{slug}.
  2) Merges list fields into JSON: title, url, venue, dates/times, summary (from complemento/short list text), categories, id, id_java, etc.
  3) By default, GET the public HTML program page and fills synopsis (long sinopse); use --no-synopsis to skip this.
  4) If the row has id_java, loads bilheteria ticket data (pricing, aggregate price, is_free when applicable) in parallel with step 3.

Global flags apply here (--format, --summary-chars). JSON output wraps the event in the usual sescli envelope with _meta.source.

HTML extraction is layout-dependent; throttle automated calls if scripting`,
		Example: `  sescli info event nadine-2 --format json
  sescli info event 'https://www.sescsp.org.br/programacao/nadine-2/' --format json
  sescli info event cinema-foo --no-synopsis --format json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			slug := programacao.SlugFromUserArg(args[0])
			if slug == "" {
				return fmt.Errorf("could not parse program slug from %q", args[0])
			}
			norm := normalize.NormalizeOpts{SummaryMax: opts.summaryChars}
			ev, source, err := a.resolveProgramEvent(ctx, slug, norm)
			if err != nil {
				return err
			}
			needSynopsis := !apiOnly
			needPricing := strings.TrimSpace(ev.JavaID) != ""
			if needSynopsis && strings.TrimSpace(ev.URL) == "" {
				return fmt.Errorf("cannot fetch synopsis: event has no URL")
			}
			ref := strings.TrimSpace(ev.URL)
			if ref == "" {
				ref = "https://www.sescsp.org.br/programacao/"
			}
			var synText string
			var synErr error
			var pricing *normalize.EventPricing
			var priceErr error

			runSynopsis := func() {
				if a.FetchProgramSynopsis != nil {
					synText, synErr = a.FetchProgramSynopsis(ctx, ev.URL)
				} else {
					synText, synErr = programpage.FetchSynopsis(ctx, ev.URL)
				}
			}
			runPricing := func() {
				if a.FetchActivityPricing != nil {
					pricing, priceErr = a.FetchActivityPricing(ctx, ev.JavaID, ref)
				} else {
					pricing, priceErr = bilheteria.FetchActivityPricing(ctx, ev.JavaID, ref)
				}
			}

			switch {
			case needSynopsis && needPricing:
				var wg sync.WaitGroup
				wg.Add(2)
				go func() {
					defer wg.Done()
					runSynopsis()
				}()
				go func() {
					defer wg.Done()
					runPricing()
				}()
				wg.Wait()
				if synErr != nil {
					return fmt.Errorf("synopsis: %w", synErr)
				}
				ev.Synopsis = synText
				if priceErr == nil && pricing != nil {
					normalize.ApplyBilheteriaPricing(&ev, pricing)
				}
			case needSynopsis:
				runSynopsis()
				if synErr != nil {
					return fmt.Errorf("synopsis: %w", synErr)
				}
				ev.Synopsis = synText
			case needPricing:
				runPricing()
				if priceErr == nil && pricing != nil {
					normalize.ApplyBilheteriaPricing(&ev, pricing)
				}
			}
			meta := outfmt.Meta{Source: source, Query: "event:" + slug}
			payload := outfmt.Response("event", ev, meta)
			return printPayload(stdout, opts.format, payload, []normalize.Event{ev}, meta)
		},
	}
	cmd.Flags().BoolVar(&apiOnly, "no-synopsis", false, "skip HTML fetch; return only search/API fields (no synopsis)")
	return cmd
}

func (a App) resolveProgramEvent(ctx context.Context, slug string, norm normalize.NormalizeOpts) (normalize.Event, string, error) {
	_ = ctx
	if a.FetchProgramBySlug != nil {
		return a.FetchProgramBySlug(ctx, slug)
	}
	rawURL := sescapi.AtividadesSearchURL(slug, 50)
	var raw any
	if err := client.New(client.Options{Timeout: 25 * time.Second, Retries: 2}).GetJSON(rawURL, &raw); err != nil {
		return normalize.Event{}, rawURL, err
	}
	events := normalize.EventsFromRawOpts(raw, false, norm)
	ev, ok := programacao.PickEventByProgramSlug(events, slug)
	if !ok {
		return normalize.Event{}, rawURL, fmt.Errorf("no activity in search results matched program slug %q (empty or unrelated results; try another slug)", slug)
	}
	return ev, rawURL, nil
}

func (a App) facetsCommand(ctx context.Context, stdout io.Writer, opts *options) *cobra.Command {
	var mode string
	cmd := &cobra.Command{
		Use:   "facets",
		Short: "Inspect raw SESC facet metadata (languages, access/activity filters)",
		RunE: func(cmd *cobra.Command, args []string) error {
			fetch := a.FetchFacets
			if fetch == nil {
				fetch = realFacetFetcher()
			}
			q := sescapi.DinamicoQuery{
				Mode:          mode,
				Audience:      opts.audience,
				Units:         opts.unitIDsForPreset(),
				ActivityTypes: presets.ActivityTypesForProfile(opts.profile),
			}
			raw, source, err := fetch(ctx, q)
			if err != nil {
				return err
			}
			payload := outfmt.Response("facets", raw, outfmt.Meta{Source: source, Query: "facets:" + mode})
			return printPayload(stdout, opts.format, payload, nil, outfmt.Meta{})
		},
	}
	cmd.Flags().StringVar(&mode, "mode", sescapi.ModeLinguagens, "dinamico mode: unidade, tipos_linguagens, acesso")
	return cmd
}

func (a App) configCommand(stdout io.Writer, opts *options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Create or inspect sescli defaults",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "init",
		Short: "Create a default config file (presets seeded from zonacentral)",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Default()
			path := opts.configPath
			if path == "" {
				var err error
				path, err = config.Path()
				if err != nil {
					return err
				}
			}
			if err := config.Write(path, cfg, opts.force); err != nil {
				if os.IsExist(err) {
					return fmt.Errorf("config already exists at %s (use --force to overwrite)", path)
				}
				return err
			}
			_, err := fmt.Fprintf(stdout, "created %s", path)
			return err
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "path",
		Short: "Print the config path",
		RunE: func(cmd *cobra.Command, args []string) error {
			path := opts.configPath
			if path == "" {
				var err error
				path, err = config.Path()
				if err != nil {
					return err
				}
			}
			_, err := fmt.Fprint(stdout, path)
			return err
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "setup",
		Short: "Interactively configure sescli defaults",
		RunE: func(cmd *cobra.Command, args []string) error {
			input := a.Stdin
			if input == nil {
				input = os.Stdin
			}
			_, err := config.Setup(input, stdout, opts.configPath, true)
			return err
		},
	})
	cmd.AddCommand(a.presetsConfigCommand(stdout, opts))
	cmd.PersistentFlags().BoolVar(&opts.force, "force", opts.force, "overwrite existing config")
	return cmd
}

func resolveConfigPath(optPath string) (string, error) {
	if optPath != "" {
		return optPath, nil
	}
	return config.Path()
}

func (a App) presetsConfigCommand(stdout io.Writer, opts *options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "presets",
		Short: "Manage unit ID lists (--where presets) stored in config",
	}

	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "Print effective preset keys and comma-separated unit IDs",
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := resolveConfigPath(opts.configPath)
			if err != nil {
				return err
			}
			cfg, err := config.Load(opts.configPath)
			if err != nil {
				return err
			}
			for _, name := range config.PresetNamesForList(cfg) {
				ids := config.EffectivePreset(cfg, name)
				_, err := fmt.Fprintf(stdout, "%s\t%s\n", name, strings.Join(ids, ","))
				if err != nil {
					return err
				}
			}
			_, err = fmt.Fprintf(stdout, "(config %s)\n", path)
			return err
		},
	}

	showCmd := &cobra.Command{
		Use:   "show PRESET",
		Short: "Print one preset's effective IDs",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := resolveConfigPath(opts.configPath)
			if err != nil {
				return err
			}
			cfg, err := config.Load(opts.configPath)
			if err != nil {
				return err
			}
			canonical, err := config.NormalizePresetName(args[0])
			if err != nil {
				return err
			}
			ids := config.EffectivePreset(cfg, canonical)
			_, err = fmt.Fprintf(stdout, "%s\n%s\n(config %s)\n",
				canonical,
				strings.Join(ids, ","),
				path,
			)
			return err
		},
	}

	setCmd := &cobra.Command{
		Use:     "set PRESET IDS...",
		Short:   "Replace preset with these unit IDs (comma or space separated)",
		Example: "  sescli config presets set meu-lado 43,52,56\n\n  sescli config presets add centro 56",
		Args:    cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := resolveConfigPath(opts.configPath)
			if err != nil {
				return err
			}
			ids := config.ParseUnitIDArgs(args[1:])
			if len(ids) == 0 {
				return fmt.Errorf("provide at least one unit id")
			}
			if err := config.Update(opts.configPath, func(cfg *config.Config) error {
				return config.SetPresetIDs(cfg, args[0], ids)
			}); err != nil {
				return err
			}
			canonical, err := config.NormalizePresetName(args[0])
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(stdout, "updated preset %q (%d ids) → %s\n", canonical, len(ids), path)
			return err
		},
	}

	addCmd := &cobra.Command{
		Use:   "add PRESET IDS...",
		Short: "Add IDs to the effective list (starts from zonacentral for centro when unset)",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := resolveConfigPath(opts.configPath)
			if err != nil {
				return err
			}
			extra := config.ParseUnitIDArgs(args[1:])
			if len(extra) == 0 {
				return fmt.Errorf("provide at least one unit id to add")
			}
			if err := config.Update(opts.configPath, func(cfg *config.Config) error {
				return config.AddPresetIDs(cfg, args[0], extra)
			}); err != nil {
				return err
			}
			cfg, err := config.Load(opts.configPath)
			if err != nil {
				return err
			}
			canonical, err := config.NormalizePresetName(args[0])
			if err != nil {
				return err
			}
			n := len(config.EffectivePreset(cfg, canonical))
			_, err = fmt.Fprintf(stdout, "updated preset %q (now %d ids) → %s\n", canonical, n, path)
			return err
		},
	}

	removeCmd := &cobra.Command{
		Use:     "remove PRESET IDS...",
		Aliases: []string{"rm"},
		Short:   "Remove IDs from the effective list; result is stored as the override",
		Args:    cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := resolveConfigPath(opts.configPath)
			if err != nil {
				return err
			}
			rem := config.ParseUnitIDArgs(args[1:])
			if len(rem) == 0 {
				return fmt.Errorf("provide at least one unit id to remove")
			}
			if err := config.Update(opts.configPath, func(cfg *config.Config) error {
				return config.RemovePresetIDs(cfg, args[0], rem)
			}); err != nil {
				return err
			}
			cfg, err := config.Load(opts.configPath)
			if err != nil {
				return err
			}
			canonical, err := config.NormalizePresetName(args[0])
			if err != nil {
				return err
			}
			n := len(config.EffectivePreset(cfg, canonical))
			_, err = fmt.Fprintf(stdout, "updated preset %q (now %d ids) → %s\n", canonical, n, path)
			return err
		},
	}

	unsetCmd := &cobra.Command{
		Use:   "unset PRESET",
		Short: "Drop preset from config; centro falls back to zonacentral geography again",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := resolveConfigPath(opts.configPath)
			if err != nil {
				return err
			}
			canonical, err := config.NormalizePresetName(args[0])
			if err != nil {
				return err
			}
			if err := config.Update(opts.configPath, func(cfg *config.Config) error {
				return config.UnsetPreset(cfg, canonical)
			}); err != nil {
				return err
			}
			_, err = fmt.Fprintf(stdout, "removed config entry for preset %q → %s\n", canonical, path)
			return err
		},
	}

	cmd.AddCommand(listCmd, showCmd, setCmd, addCmd, removeCmd, unsetCmd)
	return cmd
}

func (a App) printEvents(ctx context.Context, stdout io.Writer, dq querymodel.Query, opts *options) error {
	fetch := a.FetchEvents
	if fetch == nil {
		fetch = realEventFetcher(opts.includeRaw, normalize.NormalizeOpts{SummaryMax: opts.summaryChars})
	}
	g, err := gatherEventsWithMinResults(ctx, fetch, dq, opts.minResults)
	if err != nil {
		return err
	}
	events := g.Events
	if dq.When.FromNow {
		events = eventtime.DropStartedBefore(events, a.now())
	}
	api := dq.EventsQuery()
	meta := outfmt.EventListMeta(len(events), api.Page, api.PerPage, g.ReportedTotal, g.Source, "events")
	meta.DateFrom = dq.When.From
	meta.DateTo = g.EffectiveDateTo
	if opts.minResults > 0 && api.Page <= 1 {
		meta.MinResultsTarget = opts.minResults
		meta.MinResultsWidenedDate = g.WidenedDate
		meta.MinResultsWidenedWhere = g.WidenedWhere
	}
	payload := outfmt.Response("events", events, meta)
	return printPayload(stdout, opts.format, payload, events, meta)
}

func printPayload(stdout io.Writer, format string, payload any, events []normalize.Event, meta outfmt.Meta) error {
	var (
		out string
		err error
	)
	switch format {
	case "json", "":
		out, err = outfmt.JSON(payload, false)
	case "pretty":
		out, err = outfmt.JSON(payload, true)
	case "whatsapp", "wa", "chat":
		out = outfmt.WhatsApp(events)
		out += outfmt.WhatsAppQueryFooter(meta)
		out += outfmt.WhatsAppPaginationHint(meta)
	case "table":
		out = outfmt.Table(events)
		out += outfmt.WhatsAppQueryFooter(meta)
		out += outfmt.WhatsAppPaginationHint(meta)
	default:
		return fmt.Errorf("unknown format %q", format)
	}
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(stdout, out)
	return err
}

func addGlobalFlags(cmd *cobra.Command, opts *options) {
	cmd.PersistentFlags().StringVar(&opts.format, "format", opts.format, "output format: json, whatsapp, pretty, table")
	cmd.PersistentFlags().StringVar(&opts.audience, "audience", opts.audience, "audience tag, default adulto")
	cmd.PersistentFlags().StringVar(&opts.profile, "profile", opts.profile, "activity profile: cultural, sports, or explicit activity slug")
	cmd.PersistentFlags().StringVar(&opts.what, "what", opts.what, "activity profile or comma-separated slugs (sescli info what)")
	cmd.PersistentFlags().StringVar(&opts.when, "when", opts.when, "date shortcut or date: today, tomorrow, YYYY-MM-DD")
	cmd.PersistentFlags().StringVar(&opts.where, "where", opts.where, "preset key (default centro), heuristic zone (zona-sul,...), venue name/slug/id")
	cmd.PersistentFlags().StringSliceVar(&opts.units, "units", opts.units, "comma-separated unit IDs, or repeat the flag")
	cmd.PersistentFlags().StringSliceVar(&opts.units, "venues", opts.units, "comma-separated venue/unit IDs, or repeat the flag")
	cmd.PersistentFlags().StringSliceVar(&opts.unitNames, "unit", opts.unitNames, "unit name/slug/code (example: ipiranga, cinesesc, 56); repeat or comma-separate")
	cmd.PersistentFlags().StringSliceVar(&opts.unitNames, "venue", opts.unitNames, "venue name/slug/code (example: ipiranga, cinesesc, 56); repeat or comma-separate")
	_ = cmd.PersistentFlags().MarkHidden("unit")
	_ = cmd.PersistentFlags().MarkHidden("units")
	cmd.PersistentFlags().StringVar(&opts.from, "from", opts.from, "start date YYYY-MM-DD")
	cmd.PersistentFlags().StringVar(&opts.to, "to", opts.to, "end date YYYY-MM-DD")
	cmd.PersistentFlags().IntVar(&opts.perPage, "limit", opts.perPage, "events per page (maps to ppp)")
	cmd.PersistentFlags().IntVar(&opts.page, "page", opts.page, "page number")
	cmd.PersistentFlags().BoolVar(&opts.fromNow, "from-now", opts.fromNow, "drop events whose start time is before now (Sao Paulo time)")
	cmd.PersistentFlags().StringVar(&opts.preset, "preset", opts.preset, "unit preset from config (default centro name) or geographic override")
	cmd.PersistentFlags().StringVar(&opts.configPath, "config", opts.configPath, "config file path (default: OS user config dir, or SESCLI_CONFIG)")
	cmd.PersistentFlags().IntVar(&opts.minResults, "min-results", 0, "page 1: widen date/region until at least N distinct events (capped by --limit)")
	cmd.PersistentFlags().IntVar(&opts.summaryChars, "summary-chars", 220, "max runes for event summary in JSON output (0 disables truncation)")
	cmd.PersistentFlags().BoolVar(&opts.includeRaw, "include-raw", opts.includeRaw, "include raw WordPress payload in JSON")
	for _, name := range []string{"audience", "profile", "units", "venues", "unit", "venue", "from", "to", "from-now", "preset", "include-raw", "summary-chars"} {
		_ = cmd.PersistentFlags().MarkHidden(name)
	}
}

func defaultOptions() options {
	defaults := presets.Defaults()
	cfg := config.Default()
	return options{
		format:       cfg.Format,
		audience:     defaults.Audience,
		profile:      defaults.Profile,
		preset:       "",
		perPage:      defaults.PerPage,
		page:         defaults.Page,
		summaryChars: 220,
		cfg:          cfg,
	}
}

func (o *options) loadConfig() {
	cfg, err := config.Load(o.configPath)
	if err != nil {
		return
	}
	o.cfg = cfg
	if o.format == "" || o.format == "json" {
		o.format = cfg.Format
	}
	if o.audience == "" || o.audience == "adulto" {
		o.audience = cfg.Audience
	}
	if o.profile == "" || o.profile == "cultural" {
		o.profile = cfg.Profile
	}
	if o.preset == "" {
		o.preset = cfg.DefaultPreset
	}
	if o.preset == "" {
		o.preset = "capital"
	}
	if o.perPage == 0 || o.perPage == presets.Defaults().PerPage {
		o.perPage = cfg.Limit
	}
	if o.page == 0 || o.page == 1 {
		o.page = cfg.Page
	}
}

func (o options) domainQuery(base time.Time) (querymodel.Query, error) {
	return exec.BuildQuery(exec.QueryInput{
		When:          o.when,
		Where:         o.where,
		Preset:        o.preset,
		Profile:       o.profile,
		What:          o.what,
		Audience:      o.audience,
		From:          o.from,
		To:            o.to,
		FromNow:       o.fromNow,
		Units:         o.units,
		UnitNames:     o.unitNames,
		PerPage:       o.perPage,
		Page:          o.page,
		Format:        o.format,
		IncludeRaw:    o.includeRaw,
		SummaryChars:  o.summaryChars,
		PresetUnitIDs: o.cfg.Presets,
	}, base)
}

func (o options) unitIDsForPreset() []string {
	if ids, ok := o.cfg.Presets[o.preset]; ok && len(ids) > 0 {
		return ids
	}
	return presets.UnitIDsForPreset(o.preset)
}

func (o *options) applyWhere(where string) {
	if ids, ok := o.cfg.Presets[where]; ok && len(ids) > 0 {
		o.preset = where
		return
	}
	if presets.IsBuiltinWhereGeography(where) {
		// Zona-*, interior, litoral, metropolitana, capital/cidade/… — keep --where literal for where.Resolve.
		return
	}
	switch strings.ToLower(strings.TrimSpace(where)) {
	case "centro", "center", "default":
		o.preset = strings.ToLower(strings.TrimSpace(where))
	default:
		o.unitNames = append(o.unitNames, where)
	}
}

func realEventFetcher(includeRaw bool, norm normalize.NormalizeOpts) EventFetcher {
	return func(ctx context.Context, q sescapi.EventsQuery) (EventFetch, error) {
		_ = ctx
		rawURL, err := sescapi.EventsURL(q)
		if err != nil {
			return EventFetch{}, err
		}
		var raw any
		err = client.New(client.Options{}).GetJSON(rawURL, &raw)
		if err != nil {
			return EventFetch{}, err
		}
		return EventFetch{
			Events:        normalize.EventsFromRawOpts(raw, includeRaw, norm),
			Source:        rawURL,
			ReportedTotal: normalize.FilterReportedTotalPtr(raw),
		}, nil
	}
}

func annotateNormalizeUnitZones(units []normalize.Unit) {
	for i := range units {
		if units[i].ID == "" {
			continue
		}
		z := presets.UrbanMacro(units[i].ID)
		if z != "" {
			units[i].Zone = z
		}
	}
}

func realUnitFetcher(includeRaw bool) UnitFetcher {
	return func(ctx context.Context, q sescapi.DinamicoQuery) ([]normalize.Unit, string, error) {
		switch q.Mode {
		case sescapi.ModeUnidade, "":
			rawURL := sescapi.UnidadesAtividadesURL()
			var raw any
			err := client.New(client.Options{}).GetJSON(rawURL, &raw)
			if err != nil {
				return nil, rawURL, err
			}
			units := normalize.UnitsFromRaw(raw, includeRaw)
			annotateNormalizeUnitZones(units)
			return units, rawURL, nil
		default:
			rawURL, err := sescapi.DinamicoURL(q)
			if err != nil {
				return nil, "", err
			}
			var raw any
			err = client.New(client.Options{}).GetJSON(rawURL, &raw)
			if err != nil {
				return nil, rawURL, err
			}
			units := normalize.UnitsFromRaw(raw, includeRaw)
			annotateNormalizeUnitZones(units)
			return units, rawURL, nil
		}
	}
}

func realFacetFetcher() FacetFetcher {
	return func(ctx context.Context, q sescapi.DinamicoQuery) (any, string, error) {
		rawURL, err := sescapi.DinamicoURL(q)
		if err != nil {
			return nil, "", err
		}
		var raw any
		err = client.New(client.Options{}).GetJSON(rawURL, &raw)
		if err != nil {
			return nil, rawURL, err
		}
		return raw, rawURL, nil
	}
}

func (a App) dateWindow(offsetDays int) (string, string) {
	date := a.now().AddDate(0, 0, offsetDays).Format(time.DateOnly)
	return date, date
}

func (a App) now() time.Time {
	if a.Now != nil {
		return a.Now()
	}
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		loc = time.FixedZone("America/Sao_Paulo", -3*60*60)
	}
	return time.Now().In(loc)
}

func splitCSV(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			if trimmed := strings.TrimSpace(part); trimmed != "" {
				out = append(out, trimmed)
			}
		}
	}
	return out
}

func trimGoRunSentinel(args []string) []string {
	for len(args) > 0 && args[0] == "--" {
		args = args[1:]
	}
	return args
}

func Run() {
	app := App{}
	_, _ = config.Ensure("")
	args := trimGoRunSentinel(os.Args[1:])
	if err := app.Execute(context.Background(), args, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(exitCode(err))
	}
}

func exitCode(err error) int {
	if err != nil && strings.Contains(err.Error(), "unknown") {
		return 2
	}
	if err != nil && strings.Contains(err.Error(), "required") {
		return 2
	}
	if err != nil && strings.Contains(err.Error(), "invalid") {
		return 2
	}
	if err != nil && strings.Contains(err.Error(), "flag") {
		return 2
	}
	return 1
}
