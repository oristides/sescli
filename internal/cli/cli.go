package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"sescli/internal/client"
	"sescli/internal/config"
	"sescli/internal/eventtime"
	"sescli/internal/exec"
	outfmt "sescli/internal/format"
	"sescli/internal/normalize"
	"sescli/internal/presets"
	querymodel "sescli/internal/query"
	"sescli/internal/sescapi"
)

type EventFetcher func(context.Context, sescapi.EventsQuery) ([]normalize.Event, string, error)
type UnitFetcher func(context.Context, sescapi.DinamicoQuery) ([]normalize.Unit, string, error)
type FacetFetcher func(context.Context, sescapi.DinamicoQuery) (any, string, error)

type App struct {
	FetchEvents EventFetcher
	FetchUnits  UnitFetcher
	FetchFacets FacetFetcher
	Now         func() time.Time
	Stdin       io.Reader
}

type options struct {
	format     string
	includeRaw bool
	units      []string
	unitNames  []string
	audience   string
	profile    string
	what       string
	when       string
	where      string
	from       string
	to         string
	perPage    int
	page       int
	fromNow    bool
	preset     string
	configPath string
	force      bool
	cfg        config.Config
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
	cmd.SetHelpTemplate(rootHelpTemplate())
	addGlobalFlags(cmd, opts)
	cmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		opts.loadConfig()
	}

	capitalCmd := a.capitalCommand(ctx, stdout, opts)
	capitalCmd.Hidden = true
	cmd.AddCommand(capitalCmd)
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
  sescli --when today --where centro --what all --limit 10

WHEN:
  today, tomorrow, from-now, YYYY-MM-DD, weekend, next-weekend,
  "from tomorrow to sunday", "next friday"

WHERE:
  centro, capital, venue name/slug/code like ipiranga, cinesesc, 56

WHAT:
  cultural, all, cinema, teatro, sports, or a comma-separated activity slug list

OPTIONS:
  --format json|whatsapp|pretty|table
  --limit N
  --page N`
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

func (a App) capitalCommand(ctx context.Context, stdout io.Writer, opts *options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "capital",
		Short: "Use configured central Sao Paulo defaults",
	}
	cmd.AddCommand(a.dateCommand(ctx, stdout, opts, "today", 0))
	cmd.AddCommand(a.dateCommand(ctx, stdout, opts, "tomorrow", 1))
	return cmd
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
		Short: "List SESC units/facets from dinamico modes=unidade",
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
			return printPayload(stdout, opts.format, payload, nil)
		},
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "search QUERY",
		Short: "Search local unit index by name, slug, typo, or code",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			matches := presets.SearchUnits(args[0])
			payload := outfmt.Response("units", matches, outfmt.Meta{Total: len(matches), Query: "units:search"})
			return printPayload(stdout, opts.format, payload, nil)
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
		Short: "Lookup supporting information such as venues and facets",
	}
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
			return printPayload(stdout, opts.format, payload, nil)
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
		Short: "Create a default config file with centro units",
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
	cmd.PersistentFlags().BoolVar(&opts.force, "force", opts.force, "overwrite existing config")
	return cmd
}

func (a App) printEvents(ctx context.Context, stdout io.Writer, dq querymodel.Query, opts *options) error {
	fetch := a.FetchEvents
	if fetch == nil {
		fetch = realEventFetcher(opts.includeRaw)
	}
	api := dq.EventsQuery()
	events, source, err := fetch(ctx, api)
	if err != nil {
		return err
	}
	if dq.When.FromNow {
		events = eventtime.DropStartedBefore(events, a.now())
	}
	payload := outfmt.Response("events", events, outfmt.Meta{Total: len(events), Source: source, Query: "events"})
	return printPayload(stdout, opts.format, payload, events)
}

func printPayload(stdout io.Writer, format string, payload any, events []normalize.Event) error {
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
	case "table":
		out = outfmt.Table(events)
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
	cmd.PersistentFlags().StringVar(&opts.what, "what", opts.what, "activity/profile/audience concept: cultural, all, cinema, teatro")
	cmd.PersistentFlags().StringVar(&opts.when, "when", opts.when, "date shortcut or date: today, tomorrow, YYYY-MM-DD")
	cmd.PersistentFlags().StringVar(&opts.where, "where", opts.where, "venue/preset: centro, capital, ipiranga, cinesesc, 56")
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
	cmd.PersistentFlags().StringVar(&opts.preset, "preset", opts.preset, "unit preset from config: centro, capital, or custom")
	cmd.PersistentFlags().StringVar(&opts.configPath, "config", opts.configPath, "config file path (default: OS user config dir, or SESCLI_CONFIG)")
	cmd.PersistentFlags().BoolVar(&opts.includeRaw, "include-raw", opts.includeRaw, "include raw WordPress payload in JSON")
	for _, name := range []string{"audience", "profile", "units", "venues", "unit", "venue", "from", "to", "from-now", "preset", "include-raw"} {
		_ = cmd.PersistentFlags().MarkHidden(name)
	}
}

func defaultOptions() options {
	defaults := presets.Defaults()
	cfg := config.Default()
	return options{
		format:   cfg.Format,
		audience: defaults.Audience,
		profile:  defaults.Profile,
		preset:   defaults.Preset,
		perPage:  defaults.PerPage,
		page:     defaults.Page,
		cfg:      cfg,
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
	if o.preset == "" || o.preset == "centro" {
		o.preset = cfg.DefaultPreset
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
		When:       o.when,
		Where:      o.where,
		Preset:     o.preset,
		Profile:    o.profile,
		What:       o.what,
		Audience:   o.audience,
		From:       o.from,
		To:         o.to,
		FromNow:    o.fromNow,
		Units:      o.units,
		UnitNames:  o.unitNames,
		PerPage:    o.perPage,
		Page:       o.page,
		Format:     o.format,
		IncludeRaw: o.includeRaw,
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
	switch strings.ToLower(strings.TrimSpace(where)) {
	case "centro", "center", "capital":
		o.preset = strings.ToLower(strings.TrimSpace(where))
	default:
		o.unitNames = append(o.unitNames, where)
	}
}

func realEventFetcher(includeRaw bool) EventFetcher {
	return func(ctx context.Context, q sescapi.EventsQuery) ([]normalize.Event, string, error) {
		rawURL, err := sescapi.EventsURL(q)
		if err != nil {
			return nil, "", err
		}
		var raw any
		err = client.New(client.Options{}).GetJSON(rawURL, &raw)
		if err != nil {
			return nil, rawURL, err
		}
		return normalize.EventsFromRaw(raw, includeRaw), rawURL, nil
	}
}

func realUnitFetcher(includeRaw bool) UnitFetcher {
	return func(ctx context.Context, q sescapi.DinamicoQuery) ([]normalize.Unit, string, error) {
		rawURL, err := sescapi.DinamicoURL(q)
		if err != nil {
			return nil, "", err
		}
		var raw any
		err = client.New(client.Options{}).GetJSON(rawURL, &raw)
		if err != nil {
			return nil, rawURL, err
		}
		return normalize.UnitsFromRaw(raw, includeRaw), rawURL, nil
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

func Run() {
	app := App{}
	_, _ = config.Ensure("")
	if err := app.Execute(context.Background(), os.Args[1:], os.Stdout, os.Stderr); err != nil {
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
