---
name: sesc-sp-cli
description: Use this skill whenever the user asks about SESC São Paulo events, shows, cinema, theater, workshops, or cultural programming — including questions like "what's on  this weekend?", "find a play near Ipiranga", "SESC events tomorrow", or "programação do SESC". Also use when building scripts, WhatsApp bots, or automations that query SESC SP schedules. Always use this skill for any SESC SP event lookup — don't try to answer from memory, the data is live. 
Trigger phrases (always load this skill when user mentions):
- "Sesc", "SESC" (any context — theater, cinema, events, programming)
- "sescli", "sesc-sp"
- Portuguese: "programação do sesc", "o que tem no sesc", 
  "teatro no sesc", "sesc sp", "SESC São Paulo"
---



# SESC SP CLI

`sescli` is a Go CLI that queries the live SESC São Paulo event API. Use it to answer questions about upcoming shows, cinema, theater, workshops, and other cultural programming.

## Quick reference

```bash
# What's on tomorrow at Ipiranga, cinema only, WhatsApp-ready output
sescli --when tomorrow --where ipiranga --what cinema --format whatsapp --limit 20

# Theater this weekend across central São Paulo
sescli --from tomorrow --to sunday --where centro --what teatro --format json

# Everything today at centro, up to 10 results
sescli --when today --where centro --what all --limit 10

# Natural date + zona-sul preset
sescli --when 'next thursday' --where zona-sul --what all --limit 10

# Theater next Wednesday — omitting --where uses default municipal union (capital)
sescli --when "next wednesday" --what teatro --format whatsapp --limit 50

# Explicit centro only (zona-central preset / geography)
sescli --when "next wednesday" --where centro --what teatro --format whatsapp --limit 50


```

## Where is `sescli` on this system?

The shell runs the **first** `sescli` found on **`PATH`** (older copies in `~/.local/bin` vs `$(go env GOPATH)/bin` is a common gotcha).

```bash
command -v sescli                    # directory + name of the binary that will run
type -a sescli                       # bash: list every sescli on PATH (order matters)
which -a sescli 2>/dev/null         # portable: all matches on PATH
ls -l "$(command -v sescli)"         # size/mtime; shows symlinks on many systems
```

On Linux, **`readlink -f "$(command -v sescli)"`** resolves symlinks to the real file.

## Reference documentation

In-repo specs (resolution order, presets, timezone, edge cases):

- **`--when`:** [WHEN.md](references/WHEN.md)
- **`--where`:** [WHERE.md](references/WHERE.md)
- **`--what`:** [WHAT.md](references/WHAT.md) — profiles + allowed API slugs (`sescli info what`)
- **Config & presets:** [config.md](references/config.md)
- **MCP server (`sesmcp`):** [MCP.md](references/MCP.md)

## Key flags

| Flag | What it does |
|---|---|
| `--when` | `today`, `tomorrow`, `weekend`, `YYYY-MM-DD`, natural phrases — [full guide](references/WHEN.md) |
| `--where` | Venue name (`ipiranga`), zone (`zona-sul`), preset (`centro`) — [full guide](references/WHERE.md) |
| `--what` | Profiles (`cultural`, `all`, `teatro`, `sports`, …) or comma-separated **allowed** activity slugs — [WHAT.md](references/WHAT.md); **`sescli info what`** lists values. Not free text. |
| `--format` | `json` (default), `whatsapp` (one line per event), `table`, `pretty` |
| `--limit` | Max events returned **per request** (API `ppp`); merged `--min-results` output is still capped here |
| `--min-results` | Page 1 only: if fewer distinct events than N, automatically widen **end date** (+7…+28d) then **capital** units until N (or give up). Does not raise N above `--limit`. |
| `--summary-chars` | Max runes for `summary` in JSON (default 220; `0` = no truncation). List text is usually **`complemento`**, not a full sinopse. |
| `--from-now` | Skip events already started — use for "what can I still attend right now?" |
| `--profile all` | Remove cultural filter when results are too few |

## Finding venues

```bash
sescli info what         # valid --what profiles, synonyms, and API slugs
sescli info venues search ipiranga   # find a venue by name/slug
sescli info venues --format pretty   # list all venues
sescli info event programa-slug   # or full https://…/programacao/slug/ URL (includes HTML synopsis by default)
sescli info event slug --no-synopsis   # API row only, no page fetch
```

Use `--venue NAME` for a single named venue. Use `--venues 2,43,53` only when you already have numeric IDs.

## Output formats

- `--format json` — compact JSON with `_meta` and trimmed event objects (default, good for automation)
- `--format whatsapp` — one terse line per event (plus an indented **`summary`** line when present)
- `--format table` — TAB-separated table (includes **`SUMMARY`** column)

## Defaults

- Audience: `adulto`
- **Where:** omitting **`--where`** uses **`capital`** (municipal union of all `zona-*` + `metropolitana`). Use **`--where centro`** when you only want zona-central. Interior/litoral are never part of that default — use **`--where interior`** / **`litoral`** when needed. Details: [WHERE.md](references/WHERE.md).
- Activity: cultural (theater, cinema, workshops) — use `--profile all` to broaden

## Getting more results

Use `--page 2 --limit 10` to paginate. Use `--profile all` to remove the cultural filter.

## Config

```bash
sescli config path       # find config file location
sescli config init       # create/reset defaults
sescli config setup      # interactive setup
```

For preset management, custom venue lists, curating `centro`, env vars (`SESCLI_CONFIG`), and CLI commands — see **[config.md](references/config.md)**.
