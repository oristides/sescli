# SESC SP API Inventory

This file records the currently confirmed hidden API parameters that `sescli`
maps into typed Go structs and CLI flags. Refresh with DevTools/HAR + static JS
search when the SESC programação UI changes.

## Base

- `https://www.sescsp.org.br/wp-json/wp/v1/dinamico`
- `https://www.sescsp.org.br/wp-json/wp/v1/atividades/filter`
- `https://www.sescsp.org.br/wp-json/wp/v1/atividades/search` — text search (`s`), pagination (`ppp`); **`atividade[]` rows match the filter list shape** (not a detail view).
- `https://www.sescsp.org.br/wp-json/wp/v1/unidades-atividades` — canonical roster
  for all venues (`group_id`, `group_slug`, `description` = capital/interior/litoral).

## Required Headers

- `Accept: application/json`
- `Referer: https://www.sescsp.org.br/programacao/`
- `User-Agent: sescli/...`

## `dinamico`

Known query keys:

- `unidades`
- `categorias`
- `gratuito`
- `online`
- `publico_tag`
- `tipos_atividades`
- `tipos_linguagens`
- `subcategoria`
- `modes`

Known `modes`:

- `unidade`
- `tipos_linguagens`
- `acesso`

Current production shape observed during live tests:

- Top-level object with keys like `categorias`, `unidades`.
- `unidades` may be grouped by region in the payload (often labeled `capital` /
  interior / litoral in the site's own grouping — not the same as SESCLI presets).
  Items include fields such as `groupID`, `groupName`, `groupLink`.
- `sescli` does **not** mirror that API grouping as a preset. Default footprint is
  preset **`centro`** in `config.json`, **seeded at install** from SESCLI's
  **zonacentral** heuristic (same geography as **`--where zona-central`** unless
  you edit **`presets.centro`**). Use **`metropolitana`**, **`zona-*`**, etc. via
  `--where` for larger regional filters.
- Editorial **`zone`** labels on roster units mirror those geographic buckets —
  orthogonal to **`presets`**, which remain user-maintained IDs.
- `sescli` keeps a static reverse-search unit index so users can write
  `--unit ipiranga` or run `units search ipiranga`; this resolves names/slugs
  to the numeric `local` IDs required by `atividades/filter`.
- CLI pagination maps `--limit` to `ppp` and `--page` to `page`; these flags are
  available on shortcut commands as well as `events`.
- `--profile all` intentionally sends an empty `atividade` filter to broaden
  results.
- Canonical CLI input is now modeled as `when`, `where`, `what`, page, and
  output. `--what` maps to the API `atividade` filter and audience defaults.
  Invalid `--what` values are **rejected before the HTTP call** (see
  `internal/what/Validate`, `sescli info what`, `skills/references/WHAT.md`).
- `--when` is parsed by deterministic rules first, then by
  `github.com/anatol/naturaldate.go` for natural English phrases.
- Root CLI help intentionally exposes only canonical query flags and the
  supporting commands `config` and `info`. Compatibility commands/flags remain
  available but hidden.

## `atividades/filter`

Known query keys:

- `local`
- `categoria`
- `gratuito`
- `online`
- `publico`
- `atividade`
- `linguagem`
- `data_inicial`
- `data_final`
- `tipo`
- `dinamico`
- `ppp`
- `page`

## `atividades/search`

Known query keys (observed):

- `s` — search string (program slug fragment or keywords).
- `ppp` — page size (optional).

## `atividades/filter` response

Current production shape observed during live tests:

- Top-level object with keys like `editorial`, `atividade`, `total`.
- `atividade` is the event list.
- Events currently include useful fields such as `id`, `titulo`, `link`,
  `unidade`, `tipos_linguagens`, `dataProxSessao`, `dataPrimeiraSessao`,
  `dataUltimaSessao`, `gratuito`, `preco`, `valores`, `complemento`.
- **List payloads (filter + search)** do **not** ship WordPress-style long copy
  (`post_excerpt`, `post_content`, `resumo`, `sinopse`, `description`). Short
  subtitle/credits live in **`complemento`** — SESCLI maps that into JSON
  `summary` (after excerpt/resumo when present). See
  [`programacao-urls-and-slugs.md`](programacao-urls-and-slugs.md) for URL/slug
  notes and detail limitations.

## MCP surface

[`cmd/sesmcp`](../../cmd/sesmcp) exposes `sesc_search_events` / `sesc_search_venues` over stdio MCP. Those tools call `internal/exec.BuildQuery`, the HTTP stack in `internal/client`, and normalization in `internal/normalize` — not a distinct protocol contract beyond MCP JSON-RPC.
