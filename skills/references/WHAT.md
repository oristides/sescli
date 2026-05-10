# `--what` (activity / profile)

`--what` is **not** free text. Invalid values fail fast with a hint (instead of an empty API result).

## Profiles (single token only)

Use **one** of these for the whole expression — **do not** combine them with commas:

| Value | Meaning |
| --- | --- |
| `cultural` | Default bundle: theater (as shows bucket), cinema, música, dança, oficinas, etc. See `CulturalBundleSlugs` in `internal/what/slugs.go`. |
| `all` / `any` / `todos` / `todas` | No `atividade` filter (widest; use with care). |
| `teatro` | `atividade=shows-espetaculos-e-performances` **and** `linguagem=teatro` (theater language only). |
| `sports` / `esportes` | `esporte-e-atividade-fisica`. |

## Activity slugs (one or comma-separated)

Otherwise each segment must be an **allowed API slug** (and `teatro` in a list maps to the shows parent **without** `linguagem` — unlike the standalone profile `teatro`).

Common cases:

- **All show / performance types** (not only teatro): `shows-espetaculos-e-performances`
- **Cinema**: `cinema`
- **Courses / workshops (broad)**: `cursos-e-oficinas` or the specific `*-cursos-e-oficinas` slugs

The canonical list in the CLI is:

```bash
sescli info what
```

## Synonyms (aliases)

These are rewritten to API slugs before the request:

| You type | Becomes |
| --- | --- |
| `oficina` | `cursos-e-oficinas` |
| `sports` / `esportes` | `esporte-e-atividade-fisica` (including inside a comma list) |

## Common mistakes

| Input | Problem | Use instead |
| --- | --- | --- |
| `espetaculo`, `espetáculos` | Not a valid slug | `shows-espetaculos-e-performances`, or profile `teatro` for theater-only |
| `teatro` expecting “all shows” | Profile `teatro` adds **linguagem=teatro** | `shows-espetaculos-e-performances` |
| `cultural,cinema` | `cultural` cannot be part of a comma list | `cinema` alone or `cultural` alone |

## Event `summary` and detail (list vs program page)

List responses (`atividades/filter` and `atividades/search`) do **not** include long WordPress-style sinopses. SESCLI maps **`complemento`** into JSON **`summary`**. Truncation: **`--summary-chars`** (default 220; `0` = no truncation on API-mapped fields).

**Long text (sinopse)** from the public website: **`sescli info event SLUG_OR_URL`** fetches the program HTML **by default** and fills JSON **`synopsis`** (main editorial block + Open Graph fallback — see **`internal/programpage`**). Use **`--no-synopsis`** for API-only output (faster, no second HTTP request).

**Ticket prices:** the same command loads **`pricing`** (inteira/meia/comerciário, gratuito, ingresso status, counts) from the **bilheteria** portal when **`id_java`** is present on the search row — see **`internal/bilheteria`**.

Details: **`testdata/discovery/programacao-urls-and-slugs.md`**.

## Agents / automation

- Prefer **`sescli info what`** or this file over guessing Portuguese labels.
- Prefer **`--format json`** and inspect `_meta.source` when debugging filters.
