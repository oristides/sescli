# MCP server (`sesmcp`)

`sesmcp` is a **stdio** [Model Context Protocol](https://modelcontextprotocol.io/) server bundled with **sescli**. The host process (Cursor, Claude Desktop, etc.) starts the binary and talks JSON-RPC over stdin/stdout.

It does **not** open ports; traffic is strictly **stdin/stdout between host and child**.

## What it exposes

Two tools mirror what human operators already have on the CLI:

| Tool | Purpose | Hits the network? |
|------|---------|-------------------|
| **`sesc_search_events`** | Resolve **when / where / what**, call the live SESC SP WordPress filter API, normalize events, format text reply | **Yes** |
| **`sesc_search_venues`** | Fuzzy search on the **curated venue list** baked into `sescli` | **No** (offline) |

### `sesc_search_events` parameters

Roughly parallel to **`sescli` root flags**:

| Argument | Meaning |
|---------|---------|
| `when` | Same grammar as **`--when`** (today, ranges, natural dates). Empty → behaves like CLI default (**today**) via BuildQuery rules. See [WHEN.md](WHEN.md). |
| `where` | Same as **`--where`** (preset, zona-*, venue slug.) Empty uses **`defaults.where`** from config. See [WHERE.md](WHERE.md). |
| `what` | Same as **`--what`**. Empty uses **`defaults.what`**. |
| `from` / `to` | Optional explicit **`YYYY-MM-DD`** range (same precedence as CLI: overrides `--when` semantics when set). |
| `format` | `json`, `pretty`, `whatsapp` / `wa` / `chat`, `table`. Empty uses **`defaults.format`**. |
| `limit` | Page size (**`ppp`**). **`0`** means use config **`defaults.limit`**. |
| `page` | Page number (**1**-based). **`0`** means config **`defaults.page`**. |
| `from_now` | If **`true`**, drop events whose start is already past (São Paulo clock), after fetch — same idea as **`--from-now`**. |
| `audience` | E.g. **`adulto`**. Empty → **`defaults.audience`**. |
| `include_raw` | Embed raw WordPress payload in normalized JSON (**advanced**). |

Implementation: **`cmd/sesmcp/main.go`** → **`exec.BuildQuery`** → same URL builder **`sescapi.EventsURL`** → **`internal/client`** HTTP GET → **`normalize.EventsFromRaw`**.

### `sesc_search_venues` parameters

| Argument | Meaning |
|---------|---------|
| `query` | Free text; matched with **`presets.SearchUnits`** (typo‑tolerant, offline roster). |

## Config and defaults

On startup `sesmcp`:

1. Calls **`config.Ensure`** so missing config paths get a sane default file if appropriate.
2. **`config.Load("")`** reads **`SESCLI_CONFIG`** if set, else the OS config path (**same rules as CLI**).

So **`SESCLI_CONFIG`**, presets, and **`defaults`** (where, what, format, limit, audience, …) behave like **`sescli`**. See [config.md](config.md).

## Running it

From the repo:

```bash
go run ./cmd/sesmcp
```

Installed:

```bash
go install ./cmd/sesmcp
# PATH must contain $(go env GOPATH)/bin/sesmcp
```

Cursor (example **Settings → MCP**):

```json
{
  "mcpServers": {
    "sescli": {
      "command": "/absolute/path/to/sesmcp",
      "args": [],
      "env": {}
    }
  }
}
```

Or **`"command": "go"`**, **`"args": ["run", "./cmd/sesmcp"]`** with **`"cwd"`** set to this module directory (if your MCP UI supports **`cwd`**).

The host invokes the binary; you do **not** paste URLs into MCP config.

## Limits and etiquette

- The events tool queries the **public** API — avoid hammering it; backoff on failures.
- Results reflect **upstream** data and normalization (São Paulo time where relevant).
- For offline venue discovery, use **`sesc_search_venues`**, then call **`sesc_search_events`** with a concrete **`where`**.
