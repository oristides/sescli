# Config (`config.json`)

## Where the file lives

- **Override path:** set environment variable `SESCLI_CONFIG` to the full path of your config file.
- **Default (XDG / OS convention):** user config directory + `sescli/config.json`  
  - Linux: typically `~/.config/sescli/config.json`  
  - macOS: `~/Library/Application Support/sescli/config.json`  
  - Windows: `%AppData%\sescli\config.json`

If the file is missing, **built-in defaults** apply (no file is required to run queries).

## What it stores

`sescli` uses **JSON**. The canonical shape (`internal/config/config.go`) is:

- **`defaults`** — values used when you do not pass matching flags:
  - `where` — default venue preset name (normally `centro`)
  - `what` — interests profile (often `cultural` or overridden by `--what`)
  - `audience` — audience tag passed to the API (e.g. `adulto`)
  - `format` — default output format (`json`, `pretty`, `whatsapp`, `table`, …)
  - `limit` — default `--limit` / page size (`ppp`)
  - `page` — default page number (`1`-based)

- **`default_preset`** — kept in sync with `defaults.where` after load/normalize.

- **`presets`** — map of **preset name → list of unit IDs** strings (SESC venue codes).  
  On a fresh install, **`centro`** is seeded from the **zonacentral** geography list (see [WHERE.md](WHERE.md)). You can extend or rename presets via `sescli config …`.

- **Top-level mirrors** (`audience`, `profile`, `format`, `limit`, `page`): legacy/compatibility duplicates of things also under `defaults`; the loader merges them into `defaults` during `normalize`.

## CLI commands

Typical workflows:

```bash
sescli config init          # wizard; writes defaults + presets if file missing / with --force
sescli config show          # prints effective presets and defaults path
sescli preset list          # list preset keys in config
sescli preset show centro   # show unit IDs under a preset name
sescli preset set myzone 49,55,56
sescli preset add myzone 66
sescli preset remove myzone 66
sescli preset unset myzone # drop preset; built-in centro again follows zonacentral if unset applies
```

## How flags interact with config

- `--config /path/to/config.json` forces a path for that invocation (overrides `SESCLI_CONFIG` for that run only, per CLI implementation).
- Empty or default flags for **format**, **audience**, **profile**, **preset name**, **limit**, **page** are filled from loaded config when you `loadConfig()` (see `internal/cli/cli.go`).

**`--where`** is separate from the **preset** field: it can be a macro zone (`zona-sul`), a **preset key** from `presets`, or a **venue** expression. Resolution order is documented in [WHERE.md](WHERE.md).

## Safety notes

- Config is **local**; it only affects which unit IDs and defaults your CLI sends to the public API.
- Invalid JSON or unknown fields are handled conservatively: bad files may fall back to defaults—check `sescli config show` after edits.
