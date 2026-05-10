# `--where` (place / venues)

`--where` answers **which SESC units (venues)** participate in the query. The resolver lives in `internal/where/where.go` and uses `internal/presets` for geography and name lookup.

## Resolution order

Given `Filter{Expression: "<what you passed>", ConfigPresets: <from config>}`:

1. **Explicit unit ID list** (only when the higher layer passes `IDs`—e.g. `--units` / plumbing; not the raw `--where` string alone).
2. **Named preset from `config.json`** — case-insensitive key match in `presets`. If you have `presets.myarea: ["49","55"]`, then `--where myarea` uses those IDs.
3. **Empty, `centro`, `center`, or `default`** — uses the **zonacentral** macro-zone unit list (same IDs used to seed the default `centro` preset at install). This is a **geographic bucket**, not the old hand-picked “centro CSV” list.
4. **Urban macro zone** — if the string normalizes to a known bucket, all units in that bucket are used (sorted, deduped).
5. **Single venue expression** — otherwise the string is treated as **one** venue token: name, slug, or numeric **unit ID**, resolved via `presets.ResolveUnitIDs` (fuzzy / alias table in `internal/presets`).

## Urban macro zones (heuristic)

These labels are **approximate geography** for SESC SP units. Input is normalized: spaces/underscores → hyphens; case-insensitive; `zonasul` matches `zona-sul`.

| Canonical label | Meaning (high level) |
|-----------------|----------------------|
| `zona-central` | Municipality “central” bucket in the tool’s table |
| `zona-norte` | North macro-area |
| `zona-sul` | South macro-area |
| `zona-leste` | East macro-area |
| `zona-oeste` | West macro-area |
| `metropolitana` | **Periphery subset** in this tool’s table (a handful of unit IDs), **not** “all of Greater São Paulo”. |
| `interior` | Interior** state** venues |
| `litoral` | Coast** venues |

## Greater São Paulo (municipal) coverage

For “SESC na capital” / most city venues together (all `zona-*` buckets **plus** the periphery `metropolitana` list — **excluding** `interior` and `litoral`), use:

| Input | Meaning |
|--------|---------|
| `capital`, `cidade`, `municipio`, `grande-sp` (spacing/hyphens flexible) | Union of zona-central, zona-norte, zona-sul, zona-leste, zona-oeste, and metropolitana IDs. |

If `config.json` defines a **preset** whose key is `capital`, that preset’s ID list wins over this builtin union (preset resolution runs first).

## Troubleshooting “0 events” (teatro / capital)

| Situation | Cause |
|-----------|--------|
| You omitted `--where` | Default is **`centro` (zona-central only)**. Shows in **Pinheiros** (`zona-oeste`) or **Vila Mariana** (`zona-sul`) are **excluded**. Not an API bug — widen `--where` (e.g. `capital` or `zona-oeste,zona-sul` is not supported as one token; use `capital` or separate runs). |
| `--where capital` still 0 | **Rebuild / reinstall** `sescli` from a revision that includes the **teatro** filter fix (`shows-espetaculos-e-performances` + `linguagem=teatro`) and the **`capital` composite**. Check `_meta.source` in JSON: `atividade=` must not be bare `teatro`. |
| `--where capital` + config | If **`presets.capital`** exists in `config.json`, it **replaces** the builtin municipal union. A list copied from only **metropolitana** IDs will miss oeste/sul → **0**. Rename the preset or use **`grande-sp`** / **`municipio`** (same union, no preset name clash) if you need the builtin. |

Use **`--format json`** and inspect **`_meta.source`** (`local=…`, `atividade=…`, `linguagem=…`) to confirm the request matches what you expect.

The concrete **unit IDs** per bucket are defined in `internal/presets/zones.go` (`urbanMacroByID`). They can change when the curated roster is updated.

## Presets vs zones

| Concept | Example | Notes |
|--------|---------|--------|
| **Preset key** | `centro` | Stored in config as a **list of IDs**. Default install copies **zonacentral** into `presets.centro`. Editable via `sescli preset …`. |
| **Macro zone** | `zona-sul` | **Not** stored in config; computed from `zones.go`. |
| **Venue** | `ipiranga`, `cinesesc`, `66` | Resolves to one unit ID (or errors if ambiguous/unknown). |

## Combining ideas

- `--where centro` uses your **config preset** named `centro` if present; otherwise the builtin default list for the `centro` preset (`UnitIDsForPreset`), which aligns with seeded zonacentral IDs.
- `--where zona-sul` never looks up the word `zona-sul` as a venue name—it always expands to **all IDs** in that macro zone.

## Tips

- For an offline list with names and IDs, use **`sescli venues`** (and optional JSON tooling).
- If a `--where` value is wrong, check whether it’s being interpreted as a **preset key** before a **macro zone** (preset wins when the key exists in config).
