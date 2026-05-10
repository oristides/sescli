# Programação URLs, slugs, and detail discovery

Read-only notes for SESCLI (May 2026). Public site: `https://www.sescsp.org.br`.

## URL shape

- Event (program) pages: `https://www.sescsp.org.br/programacao/{program-slug}/`
- Trailing slash is common; slug is the first path segment after `programacao/`.
- **Program slugs** (activity pages) are **not** the same namespace as **venue** roster slugs (`group_slug` from `unidades-atividades`) used by `--where ipiranga`, etc.

## List items (`atividades/filter`, `atividades/search`)

- `link` is often a **relative** path, e.g. `/programacao/nadine-2`.
- `titulo` is display title; duplicates can exist with different slugs (e.g. `nadine` vs `nadine-2`).
- `id` / `id_java`: stable numeric identifiers on list rows (Java + WP id split).
- No `post_excerpt`, `post_content`, `resumo`, `sinopse`, or `description` on these payloads.
- Short editorial line: **`complemento`** (credits, subtitle).

## Ticket prices (bilheteria portal)

- List/search rows often omit **`preco`** / **`gratuito`** even when the activity sells tickets.
- **`id_java`** on the row matches **`idAtividade`** on the Java portal, e.g. `.../bilheteria/atividade.action?idAtividade={id_java}`.
- The public site loads that JSON through WordPress **`admin-ajax.php`** with **`action=sesc_requester_proxy`** (GET, `route=` URL-encoded portal URL). SESCLI mirrors that in **`internal/bilheteria`** and attaches **`pricing`** (and backfills **`price`**) in **`sescli info event`** when **`id_java`** is present.

## `wp-json` routes (v1 atividades)

Observed namespace index: `GET /wp-json/wp/v1/atividades` returns route metadata only.

Concrete routes used by the public site:

| Route | Role |
| --- | --- |
| `/wp-json/wp/v1/atividades/filter?...` | Filtered list (`atividade[]` + `total`). |
| `/wp-json/wp/v1/atividades/search?s=...&ppp=...` | Search list; **same row shape as filter** (`atividade[]`). Param `s` is the search string (slug or keywords). |

Some query keys (e.g. unsupported `slug=` on `/atividades`) return the namespace **discovery** JSON instead of filtered rows.

## Detail / long sinopse

- **No** authenticated `wp/v2/posts/{id}` JSON was available anonymously (401 / DRA in samples).
- **No** dedicated “single activity” JSON route was found under `wp/v1/atividades/{slug}` (404).
- List and search responses **do not** include a long synopsis field; full copy is likely only in **HTML** on the program page or behind authenticated REST.

**Implication:** **`sescli info event`** returns the normalized list/search row **and by default** loads long sinopse from the public HTML (**.principal--post--conteudo** + **`og:description`** fallback; see **`internal/programpage`**). Use **`--no-synopsis`** for API-only data. Extraction is **layout-dependent** if the site template changes.

## Client headers

- `Accept: application/json`
- `Referer: https://www.sescsp.org.br/programacao/` (documented; list endpoints work with minimal clients in tests)
- Reasonable `User-Agent`
