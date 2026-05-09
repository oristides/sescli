---
name: sesc-sp-cli
description: Use this skill whenever the user asks about SESC São Paulo events, shows, cinema, theater, workshops, or cultural programming — including questions like "what's on  this weekend?", "find a play near Ipiranga", "SESC events tomorrow", or "programação do SESC". Also use when building scripts, WhatsApp bots, or automations that query SESC SP schedules. Always use this skill for any SESC SP event lookup — don't try to answer from memory, the data is live.
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
```

## Reference documentation

In-repo specs (resolution order, presets, timezone, edge cases):

- **`--when`:** [WHEN.md](../../references/WHEN.md)
- **`--where`:** [WHERE.md](../../references/WHERE.md)
- **Config & presets:** [config.md](../../references/config.md)
- **MCP server (`sesmcp`):** [MCP.md](../../references/MCP.md)

## Key flags

| Flag | What it does |
|---|---|
| `--when` | `today`, `tomorrow`, `weekend`, `YYYY-MM-DD`, natural phrases — [full guide](../../references/WHEN.md) |
| `--where` | Venue name (`ipiranga`), zone (`zona-sul`), preset (`centro`) — [full guide](../../references/WHERE.md) |
| `--what` | Activity type: `cinema`, `teatro`, `oficina`, `all` |
| `--format` | `json` (default), `whatsapp` (one line per event), `table`, `pretty` |
| `--limit` | Max events returned |
| `--from-now` | Skip events already started — use for "what can I still attend right now?" |
| `--profile all` | Remove cultural filter when results are too few |

## Finding venues

```bash
sescli info venues search ipiranga   # find a venue by name/slug
sescli info venues --format pretty   # list all venues
```

Use `--venue NAME` for a single named venue. Use `--venues 2,43,53` only when you already have numeric IDs.

## Output formats

- `--format json` — compact JSON with `_meta` and trimmed event objects (default, good for automation)
- `--format whatsapp` — one terse line per event (good for chat bots)
- `--format table` — human-readable terminal table

## Defaults

- Audience: `adulto`
- Where: `centro` preset (zonacentral unit IDs)
- Activity: cultural (theater, cinema, workshops) — use `--profile all` to broaden

## Getting more results

Use `--page 2 --limit 10` to paginate. Use `--profile all` to remove the cultural filter.

## Config

```bash
sescli config path       # find config file location
sescli config init       # create/reset defaults
sescli config setup      # interactive setup
```

For preset management, custom venue lists, curating `centro`, env vars (`SESCLI_CONFIG`), and CLI commands — see **[config.md](../../references/config.md)**.
