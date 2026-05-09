# SESC SP API Inventory

This file records the currently confirmed hidden API parameters that `sescli`
maps into typed Go structs and CLI flags. Refresh with DevTools/HAR + static JS
search when the SESC programação UI changes.

## Base

- `https://www.sescsp.org.br/wp-json/wp/v1/dinamico`
- `https://www.sescsp.org.br/wp-json/wp/v1/atividades/filter`

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
- `unidades` may be grouped by region (`capital`, etc.) and contains items with
  fields like `groupID`, `groupName`, `groupLink`.
- Important: the API's `capital` group is not the same as the operator's
  preferred central venue set. `sescli` defaults to a smaller `centro` preset:
  `2,43,51,52,53,60,61,66,761`. Use the broader `capital` preset only when those
  farther units are desired.
- `sescli` keeps a static reverse-search unit index so users can write
  `--unit ipiranga` or run `units search ipiranga`; this resolves names/slugs
  to the numeric `local` IDs required by `atividades/filter`.
- CLI pagination maps `--limit` to `ppp` and `--page` to `page`; these flags are
  available on shortcut commands as well as `events`.
- `--profile all` intentionally sends an empty `atividade` filter to broaden
  results.
- Canonical CLI input is now modeled as `when`, `where`, `what`, page, and
  output. `--what` maps to the API `atividade` filter and audience defaults.
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

Current production shape observed during live tests:

- Top-level object with keys like `editorial`, `atividade`, `total`.
- `atividade` is the event list.
- Events currently include useful fields such as `id`, `titulo`, `link`,
  `unidade`, `tipos_linguagens`, `dataProxSessao`, `dataPrimeiraSessao`,
  `dataUltimaSessao`, `gratuito`, `preco`, `valores`, `complemento`.

## MCP surface

[`cmd/sesmcp`](../../cmd/sesmcp) exposes `sesc_search_events` / `sesc_search_venues` over stdio MCP. Those tools call `internal/exec.BuildQuery`, the HTTP stack in `internal/client`, and normalization in `internal/normalize` — not a distinct protocol contract beyond MCP JSON-RPC.
