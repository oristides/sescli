---
name: sesc-sp-cli
description: Use when querying SESC Sao Paulo programming/events from agents, scripts, WhatsApp integrations, or terminals — geographic zones (zona-*, metropolitana), named presets seeded in config at install (default `centro` from zonacentral), adult audience defaults, WhatsApp-ready output, compact JSON events.
---

# SESC SP CLI

## Overview

`sescli` is a self-contained Go CLI for SESC Sao Paulo programação data. Its
default persona is adult audience, preset **`centro`** (zonacentral unit IDs seeded
into `config.json`), and cultural interests like theater, cinema, and workshops.

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

## `--where`: geographic zones versus named presets

- **`--where zona-*`**, **`metropolitana`**, **`interior`**, **`litoral`** — expand to **all** unit IDs in that heuristic geography bucket (`zona-central` is the central-city macro-area).
- **`--where centro`** — the **preset** named **`centro`**. On a fresh **`config init`**, **`presets.centro`** is **seeded from the same zonacentral ID list**, but you curate it in config (or **`sescli config presets …`**). If you **`unset`** the preset in config or leave it empty, resolution falls back to **live zonacentral geography** again.
- A **venue** token (`ipiranga`, slug, numeric id) resolves to exactly one unit.

Aliases such as **`zonasul`** or **`zona sul`** normalize to **`zona-sul`**.

Agents can attach the **`sesmcp`** stdio server (same query stack): see the repository `README.md` MCP section.

## Defaults

- Audience defaults to `adulto`.
- Venue footprint defaults to preset **`centro`**: seeded at **`config init`** / first run as **zonacentral** IDs; edit **`presets.centro`** to curate — no separate “capital preset” exists in SESCLI anymore.
- Broader São Paulo‑metro coverage: **`--where metropolitana`** (or other zones).
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

### Curating `centro` / custom presets

- In `config.json`, the **`presets`** map holds name → list of numeric unit IDs.
- If **`presets.centro`** is present and **non-empty**, it **replaces** the fallback used for **`--where centro`** (otherwise SESCLI resolves **`centro`** from **zonacentral** geography—the same IDs **`config init`** seeds into the file).
- Add your own keys (e.g. **`"perto": ["43","52","56"]`**) and call **`sescli --where perto`**.
- **`defaults.where`** can be set to one of those names so you do not repeat the flag.
- Use **`sescli info venues search …`** against the live roster to double-check IDs.
- **`sescli config init --force`** overwrites the file — keep a backup.

**From the CLI (no manual JSON edits):**

```bash
sescli config presets list              # effective IDs per preset name
sescli config presets show centro       # one preset, newline then CSV ids

# New custom list → then use `--where trabalho`.
sescli config presets set trabalho 43,52,56

# Tweaking centro: merge into effective list (config overrides → else zonacentral).
sescli config presets add centro 56
sescli config presets remove centro 761

# Drop your centro override → back to zonacentral geography.
sescli config presets unset centro
```

Preset names **`center`** and **`default`** are normalized to **`centro`**. Paths follow `sescli config path` unless you pass **`--config`**.

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
