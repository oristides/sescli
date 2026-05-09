# `--when` (dates)

Temporal filters are parsed in **`America/Sao_Paulo`** and produce **inclusive calendar-day API bounds** (`YYYY-MM-DD`) for `data_inicial` / `data_final`, unless noted below. Parser: `internal/when/when.go`.

## Defaults

- **Empty `--when`** → **`today`** (single day).

## Built-in literals (normalized to lowercase internally)

| Input | Meaning |
|-------|---------|
| `today` | Anchor day = “now” transformed to Sao Paulo calendar date |
| `tomorrow`, `amanha`, `amanhã` | Next calendar day |
| `from-now`, `from now`, `now` | Same calendar day as `today`, plus **`FromNow=true`** downstream: after fetch, events that **already started** can be dropped (see `--from-now` in CLI) |
| `weekend` | **Saturday–Sunday** range anchored from “today” per implementation (Sunday base rolls to **next** weekend in some paths—see tests in `when_test.go`) |
| `next-weekend`, `next weekend` | The **following** Sat–Sun range after the nearest weekend logic |

## Ranges

- **`from <A> to <B>`** — e.g. `from tomorrow to sunday`. Each side is parsed with the same rules as a single `--when`.
- **`A..B`** — dot-dot range with the same left/right grammar.

Sides can be literals, **`YYYY-MM-DD`**, or natural phrases (below).

## Fixed ISO dates

`YYYY-MM-DD` alone → single day.

## Natural language fallback

Any string that **`looksNaturalDate`** accepts (contains weekday/month tokens, **`next`** / **`last`**, digits, etc.) goes through **`naturaldate.Parse`** with **future bias** (`naturaldate.WithDirection(naturaldate.Future)`).

Examples:

- **`next friday`**, **`next thursday`** — handled by **`naturaldate`** with **future** direction. If the parsed time is **exactly** the anchor instant, `sescli` returns an error (`invalid when`) so “same moment” parses do not silently pass.

Anchoring uses the CLI **`now`** (normally the real clock).

## Interaction with `--from` / `--to`

In **`exec.BuildQuery`**: if **`--from`** or **`--to`** is set, that range wins and **`--when` is ignored**. Use one style or the other (`internal/exec/exec.go`).

## Practical caveats

- **Timezone:** all single-day normalization uses Sao Paulo-local calendar dates regardless of where you run `sescli`.
- **API vs filter:** some responses may contain items whose **shown** schedules span neighboring days; **`--when`** still controls **which day window** was requested—the remote API’s filtering is authoritative.
- **“Next weekday” wording:** Prefer testing with **`sescli`** and `--format json` and inspect **`_meta.source`** URLs for `data_inicial` / `data_final` when doubting parsing.
