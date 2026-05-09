# sescli

[![CI](https://github.com/oristides/sescli/actions/workflows/ci.yml/badge.svg)](https://github.com/oristides/sescli/actions/workflows/ci.yml)
[![Coverage](https://codecov.io/gh/oristides/sescli/graph/badge.svg)](https://app.codecov.io/gh/oristides/sescli)

`sescli` is a small Go CLI for querying **SESC São Paulo** public programming (venues, facets, normalized events).

## Installation

### One-liner

```bash
curl -fsSL https://raw.githubusercontent.com/oristides/sescli/main/install.sh | sh
```

Behaviour:

1. **If** the [latest GitHub Release](https://github.com/oristides/sescli/releases) publishes **GoReleaser** artefacts whose names include **`sescli`** / **`sesmcp`** plus your OS/arch (examples: **`_linux_amd64.tar.gz`**, **`_darwin_arm64.tar.gz`**, **`_windows_amd64.zip`**), both binaries install without Go.

2. **Otherwise** the script clones `main` (or **`REF`**) and runs **`go install ./cmd/sescli ./cmd/sesmcp`** with **`GOBIN`** set to **`~/.local/bin`**, so **`Git` + `Go`** are required (Go version ≥ [`go.mod`](go.mod)).

**Install dir:** defaults to **`~/.local/bin`**. Override:

```bash
curl -fsSL https://raw.githubusercontent.com/oristides/sescli/main/install.sh | env INSTALL_DIR="$HOME/bin" sh
```

**Git ref** when building from source (branch or tag): `REF=v0.1.0 curl ... | sh` (or export `REF` before `sh`).

Add to **PATH** if needed:

```bash
export PATH="$PATH:$HOME/.local/bin"
```

### Automated releases (how it works)

**You already have this wired up:** [.github/workflows/release.yml](.github/workflows/release.yml) runs [**GoReleaser**](https://goreleaser.com/) [.goreleaser.yaml](.goreleaser.yaml) whenever a **`v*`** semver tag is **pushed** to GitHub. It attaches **`sescli_…`** / **`sesmcp_…`** archives and **`checksums.txt`** to a **GitHub Release** — no manual zip uploads.

**One-time prerequisites**

1. The workflow and config are **committed on the default branch** (usually **`main`**).
2. **Actions must be enabled** on the repo (forks sometimes default Actions off).

**Publish a release (what you actually do)**

1. Merge whatever you want in this release onto **`main`** and **`git pull`** locally.
2. Pick the next semver, e.g. **`v0.3.1`** (the **`Release`** workflow only matches tags like **`v1.2.3`**).
3. Create and push the tag (**either option**):

   ```bash
   # Option A — helper script (from repo root)
   ./scripts/release.sh v0.3.1
   ```

   ```bash
   # Option B — by hand
   git tag -a v0.3.1 -m "Release v0.3.1"
   git push origin v0.3.1
   ```

4. Open **GitHub → Actions** and wait for the **Release** run to turn green.
5. Open **GitHub → Releases** and confirm artefacts; **`curl … install.sh | sh`** will then prefer **latest release** binaries when asset names match the platform.

If a workflow **retry** fails with **`already_exists`** on uploads, GitHub kept the filenames from an earlier partial run — GoReleaser is configured with **`release.replace_existing_artifacts`** in [`.goreleaser.yaml`](.goreleaser.yaml) so a successful run deletes conflicting assets and re-uploads them. Without that behaviour, manually remove release assets or open a draft release cleanly.

Do not **force-move** an existing semver tag to a different commit. For a new intentional release, use a **fresh `v*` tag**; **`install.sh`** resolves **`releases/latest`**.

### With Go (without the script)

From the published module (works once `go.mod` declares **`module github.com/oristides/sescli`** on the default branch you install from):

```bash
go install github.com/oristides/sescli/cmd/sescli@latest
go install github.com/oristides/sescli/cmd/sesmcp@latest
```

If `go install` errors about the module path, this repo may still use **`module sescli`** — install from a clone instead (next section).

Requires a Go version **at least** the one in [`go.mod`](go.mod), and **`$(go env GOPATH)/bin`** on your **PATH**.

### From source

```bash
git clone https://github.com/oristides/sescli.git
cd sescli
go mod tidy
go build -o sescli ./cmd/sescli
go build -o sesmcp ./cmd/sesmcp
# Or install into GOPATH/bin:
go install ./cmd/sescli ./cmd/sesmcp
```

The primary query model mirrors **when / where / what**:

```bash
go run ./cmd/sescli --when tomorrow --where ipiranga --what cinema --format whatsapp --limit 20
```

See `sescli help` after install. Agent-oriented usage is documented in [`skills/sescli/SKILL.md`](skills/sescli/SKILL.md).

**References:** [references/config.md](references/config.md), [references/WHERE.md](references/WHERE.md), [references/WHEN.md](references/WHEN.md), [references/MCP.md](references/MCP.md).

## Development

```bash
go test ./... -race
go test -tags=integration ./...   # fewer tests; hits the real API
```

In GitHub Actions, unit tests upload coverage to Codecov **when** the repository defines a `CODECOV_TOKEN` secret; the workflow stays green without it (`fail_ci_if_error: false`).

Badges assume workflow [`.github/workflows/ci.yml`](.github/workflows/ci.yml) on **`oristides/sescli`**. For a fork, change the **`oristides/sescli`** segments in both badge URLs to your **`owner/repo`**. Codecov stays grey until that project exists on codecov.io and has reports.

## MCP (Model Context Protocol)

[`cmd/sesmcp`](cmd/sesmcp) is a **stdio** MCP server: the IDE spawns `sesmcp`, tools run over stdin/stdout (there is **no HTTP port** opened by sescli).

| Tool | What it does |
|------|----------------|
| **`sesc_search_events`** | **Live API:** same **`when` / `where` / `what`** stack as the CLI (`exec.BuildQuery`, `sescapi`, HTTP client); returns formatted JSON/WhatsApp/table text. |
| **`sesc_search_venues`** | **Offline:** fuzzy search against the curated venue roster in the binary (`presets.SearchUnits`). |

On startup it runs **`config.Ensure`** and **`config.Load`** — **`SESCLI_CONFIG`** and **`~/.config/sescli/config.json`** (Linux) behave like the CLI.

**Full docs (every argument, Cursor setup, etiquette):** [references/MCP.md](references/MCP.md).

Minimal Cursor (**Settings → MCP**):

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

Use an **absolute path** to `sesmcp` (`go install ./cmd/sesmcp`) when `sesmcp` is not on the host PATH, or `"command": "go", "args": ["run", "./cmd/sesmcp"]` with **`cwd`** set to this repo if your client supports it.
