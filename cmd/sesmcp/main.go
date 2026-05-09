// Package main is the MCP (Model Context Protocol) stdio adapter for sescli —
// exposes search tools backed by the same query stack as the CLI.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"sescli/internal/client"
	"sescli/internal/config"
	"sescli/internal/eventtime"
	"sescli/internal/exec"
	outfmt "sescli/internal/format"
	"sescli/internal/normalize"
	"sescli/internal/presets"
	"sescli/internal/sescapi"
)

func main() {
	_, _ = config.Ensure("")
	cfg, err := config.Load("")
	if err != nil {
		cfg = config.Default()
	}

	server := mcp.NewServer(&mcp.Implementation{Name: "sescli", Version: "0.2.0"}, nil)

	type eventsArgs struct {
		When       string `json:"when" jsonschema:"today, tomorrow, YYYY-MM-DD, weekend, or dotted/natural ranges"`
		Where      string `json:"where" jsonschema:"preset or venue slug: centro, capital, ipiranga, …"`
		What       string `json:"what" jsonschema:"cultural, all, cinema, teatro, sports, or CSV slugs"`
		Format     string `json:"format" jsonschema:"json | pretty | whatsapp | table"`
		Limit      int    `json:"limit"`
		Page       int    `json:"page"`
		FromNow    bool   `json:"from_now"`
		Audience   string `json:"audience"`
		IncludeRaw bool   `json:"include_raw"`
		From       string `json:"from" jsonschema:"YYYY-MM-DD; optional explicit range"`
		To         string `json:"to"`
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "sesc_search_events",
		Description: `Fetch upcoming SESC SP events from the WordPress-backed API using the canonical when/where/what model (mirror of sescli root flags). Respect rate limits.`,
	}, func(ctx context.Context, req *mcp.CallToolRequest, args eventsArgs) (*mcp.CallToolResult, any, error) {
		_, _ = ctx, req
		where := args.Where
		if where == "" {
			where = cfg.Defaults.Where
		}
		what := args.What
		if what == "" {
			what = cfg.Defaults.What
		}
		format := args.Format
		if format == "" {
			format = cfg.Defaults.Format
		}
		limit := args.Limit
		if limit == 0 {
			limit = cfg.Defaults.Limit
		}
		page := args.Page
		if page == 0 {
			page = cfg.Defaults.Page
		}
		audience := args.Audience
		if audience == "" {
			audience = cfg.Defaults.Audience
		}

		domain, err := exec.BuildQuery(exec.QueryInput{
			When:       args.When,
			Where:      where,
			Preset:     cfg.DefaultPreset,
			Profile:    cfg.Profile,
			What:       what,
			Audience:   audience,
			From:       args.From,
			To:         args.To,
			FromNow:    args.FromNow,
			PerPage:    limit,
			Page:       page,
			Format:     format,
			IncludeRaw: args.IncludeRaw,
		}, nowSP())
		if err != nil {
			return nil, nil, err
		}
		apiURL, err := sescapi.EventsURL(domain.EventsQuery())
		if err != nil {
			return nil, nil, err
		}
		var raw any
		if err := client.New(client.Options{Timeout: 25 * time.Second, Retries: 2}).GetJSON(apiURL, &raw); err != nil {
			return nil, nil, err
		}
		evs := normalize.EventsFromRaw(raw, args.IncludeRaw)
		if domain.When.FromNow {
			evs = eventtime.DropStartedBefore(evs, nowSP())
		}
		payload := outfmt.Response("events", evs, outfmt.Meta{
			Total:  len(evs),
			Source: apiURL,
			Query:  "events",
		})

		txt, ferr := marshalFormat(format, payload, evs)
		if ferr != nil {
			return nil, nil, ferr
		}
		return &mcp.CallToolResult{Content: []mcp.Content{
			&mcp.TextContent{Text: txt},
		}}, nil, nil
	})

	type venuesArgs struct {
		Query string `json:"query" jsonschema:"venue name fragment, slug, or typo"`
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "sesc_search_venues",
		Description: "Offline fuzzy search against the curated SESC SP venue roster embedded in sescli.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args venuesArgs) (*mcp.CallToolResult, any, error) {
		_, _ = ctx, req
		matches := presets.SearchUnits(args.Query)
		blob, err := outfmt.JSON(outfmt.Response("venues", matches, outfmt.Meta{
			Total: len(matches),
			Query: args.Query,
		}), true)
		if err != nil {
			return nil, nil, err
		}
		return &mcp.CallToolResult{Content: []mcp.Content{
			&mcp.TextContent{Text: blob},
		}}, nil, nil
	})

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("sesmcp: %v", err)
	}
}

func marshalFormat(format string, payload any, events []normalize.Event) (string, error) {
	switch format {
	case "json", "":
		return outfmt.JSON(payload, false)
	case "pretty":
		return outfmt.JSON(payload, true)
	case "whatsapp", "wa", "chat":
		return outfmt.WhatsApp(events), nil
	case "table":
		return outfmt.Table(events), nil
	default:
		return "", fmt.Errorf("unknown format %q", format)
	}
}

func nowSP() time.Time {
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		loc = time.FixedZone("America/Sao_Paulo", -3*60*60)
	}
	return time.Now().In(loc)
}
