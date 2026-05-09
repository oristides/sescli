# sescli

[![CI](https://github.com/OWNER/REPO/actions/workflows/ci.yml/badge.svg)](https://github.com/OWNER/REPO/actions/workflows/ci.yml)
[![Coverage](https://codecov.io/gh/OWNER/REPO/graph/badge.svg)](https://app.codecov.io/gh/OWNER/REPO)

`sescli` is a small Go CLI for querying **SESC São Paulo** public programming (venues, facets, normalized events).

The primary query model mirrors **when / where / what**:

```bash
go run ./cmd/sescli --when tomorrow --where ipiranga --what cinema --format whatsapp --limit 20
```

See `sescli help` after install. Agent-oriented usage is documented in [`skills/sesc-sp-cli/SKILL.md`](skills/sesc-sp-cli/SKILL.md).

## Development

```bash
go test ./... -race
go test -tags=integration ./...   # fewer tests; hits the real API
```

In GitHub Actions, unit tests upload coverage to Codecov **when** the repository defines a `CODECOV_TOKEN` secret; the workflow stays green without it (`fail_ci_if_error: false`).

Replace `OWNER/REPO` in the badges above with your GitHub slug after moving this tree into its own repo (badges assume workflow file `.github/workflows/ci.yml`).

## MCP (Model Context Protocol)

[`cmd/sesmcp`](cmd/sesmcp) runs a **stdio** MCP server exposing:

- **`sesc_search_events`** — same semantics as the CLI `--when` / `--where` / `--what` surface (respect public rate limits).
- **`sesc_search_venues`** — offline fuzzy lookup over the curated venue roster.

Configure Cursor (**Settings → MCP → Add**) with a snippet like:

```json
{
  "mcpServers": {
    "sescli": {
      "command": "sesmcp",
      "args": [],
      "env": {}
    }
  }
}
```

Use the absolute path to the `sesmcp` binary (`go install ./cmd/sesmcp`) or `"command": "go", "args": ["run", "./cmd/sesmcp"]` inside this module.

The server loads `sescli`’s usual config defaults from `SESCLI_CONFIG` / the OS config path (same as `sescli config …`).
