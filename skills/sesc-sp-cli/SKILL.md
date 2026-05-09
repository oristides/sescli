---
name: sesc-sp-cli
description: Use when querying SESC Sao Paulo programming/events from agents, scripts, WhatsApp integrations, or terminals, especially when filtering by capital units, adult audience, dates, activity profile, units, or needing compact JSON event output.
---

# SESC SP CLI

## Overview

`sescli` is a self-contained Go CLI for SESC Sao Paulo programação data. Its
default persona is adult audience, **centro** units, and cultural interests like
theater, cinema, and workshops.

Default output is compact JSON for automation. Use `--format whatsapp` for terse
plain-text lines that can be pasted into chat.

## Quickstart

```bash
sescli --when tomorrow --where ipiranga --what cinema --format whatsapp --limit 20
sescli --from tomorrow --to sunday --where centro --what teatro
sescli --when today --where centro --what all --limit 10
sescli config path
sescli info venues search ipiranga
```

Agents can attach the **`sesmcp`** stdio server (same query stack): see the repository `README.md` MCP section.

## Defaults

- Audience defaults to `adulto`.
- Venues default to the built-in/configured `centro` preset:
  `2,43,51,52,53,60,61,66,761`.
- The SESC API's own `capital` group is broader and includes venues such as
  Interlagos, Sao Caetano, Guarulhos, etc.; do not use that group when the user
  means central venues.
- Activity profile defaults to `cultural`: theater, cinema, and workshops.
- Use `--profile all` to remove the activity filter when cultural defaults
  return too few events.
- Prefer canonical `--what`; `--profile` remains available as a compatibility
  alias.
- Pagination defaults are tuned for a daily digest: page `1`, `ppp`/`--limit`
  handled internally.
- Root help is organized as `WHEN`, `WHERE`, `WHAT`, and `OPTIONS`. Redundant
  flags/commands are hidden but kept for compatibility.

## Commands

```bash
sescli --when tomorrow --where ipiranga --what cinema --format whatsapp --limit 20
sescli --from tomorrow --to sunday --where centro --what teatro
sescli --when "from tomorrow to sunday" --where centro --what teatro
sescli --when today --where centro --what all --limit 10
sescli config init
sescli --config /tmp/sescli.json config init
sescli events --from 2026-05-08 --to 2026-05-08 --limit 20
sescli events --venue ipiranga --from 2026-05-08 --to 2026-05-10
sescli events --venues 2,43,53 --profile cultural
sescli events --preset capital --format whatsapp
sescli events --preset centro --profile all --page 2 --limit 10
sescli info venues search ipiranga
sescli info venues --format pretty
sescli info facets --mode tipos_linguagens --format pretty
sescli info facets --mode acesso --format pretty
```

## Formats

- `--format json` is the default: compact JSON with `_meta` and trimmed event
  objects.
- `--format whatsapp` emits one terse line per event.
- `--format pretty` emits indented JSON.
- `--format table` is for terminal inspection.

## Time and More Results

- `--from-now` removes events whose parsed start time is before the current Sao
  Paulo time. Use it for "what can I still attend now?".
- `--when` supports deterministic values (`today`, `tomorrow`, `from-now`,
  `YYYY-MM-DD`, `today..tomorrow`, `from tomorrow to sunday`, `weekend`,
  `next-weekend`) and falls back to `naturaldate.go` for phrases such as
  `next friday`.
- Without `--from-now`, the CLI shows what the SESC API returns for the date
  window.
- Use `--page 2 --limit 10` for the next page.
- Use `--profile all` to remove theater/cinema/workshop filtering and get a
  broader list.

## Config

The real binary auto-creates a default config file on startup if it is missing.
Find it with:

```bash
sescli config path
```

Create defaults interactively with:

```bash
sescli config setup
```

Or create/overwrite defaults explicitly with:

```bash
sescli config init
sescli config init --force
```

The config is JSON and contains grouped `defaults` (`where`, `what`, `audience`,
`format`, `limit`, `page`) plus named `presets`. Old flat configs still load.
Use `SESCLI_CONFIG` or `--config PATH` to select a different file.

## Venue Lookup

Use `sescli info venues search QUERY` to find venue IDs by name, slug, typo, or ID:

```bash
sescli info venues search ipiranga
sescli info venues search sao-caetano
```

For event queries, prefer `--venue NAME` when a person names a place:

```bash
sescli events --venue ipiranga --from 2026-05-08 --to 2026-05-10
```

Use `--venues 2,43,53` only when the caller already knows codes.

## Important Constraints

The SESC endpoint is an undocumented WordPress JSON API. Requests need browser-like
headers, especially:

- `Referer: https://www.sescsp.org.br/programacao/`
- `Accept: application/json`

If results become empty or HTML appears, first verify those headers and run the
live integration tests.

## Limitations

- "Close to me" means the selected venue preset, not GPS distance.
- The hidden API can change without notice. Refresh `testdata/discovery` and
  rerun integration tests when the website UI changes.
- Broader profiles for other cities/audiences can be added later through config
  and presets.
