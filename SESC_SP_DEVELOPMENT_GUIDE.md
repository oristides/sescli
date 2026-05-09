# SESC SP CLI — Technical Development & Validation Guide

> **Purpose:** Step-by-step guide for developers who want to build, validate, and extend the `sesc.py` CLI tool — from zero to a working, agent-compatible API client. Language-agnostic context is included so teams can choose their stack with confidence.

---

## Table of Contents

1. [Language Selection — Which Programming Language?](#1-language-selection)
2. [Prerequisites & Environment Setup](#2-prerequisites--environment-setup)
3. [Phase 0 — API Reconnaissance (Before Writing a Single Line)](#3-phase-0--api-reconnaissance)
4. [Phase 1 — Validate the API is Reachable](#4-phase-1--validate-the-api-is-reachable)
5. [Phase 2 — Map the Discovery Pipeline](#5-phase-2--map-the-discovery-pipeline)
6. [Phase 3 — Validate Response Shapes](#6-phase-3--validate-response-shapes)
7. [Phase 4 — Build the Core CLI](#7-phase-4--build-the-core-cli)
8. [Phase 5 — Validation Checklist Before Shipping](#8-phase-5--validation-checklist)
9. [Phase 6 — Testing for AI Agent Compatibility](#9-phase-6--ai-agent-compatibility-testing)
10. [Common Failure Modes & How to Detect Them](#10-common-failure-modes)
11. [Maintenance & Monitoring](#11-maintenance--monitoring)

---

## 1. Language Selection

### Recommended: Python 3.10+

Python is the right choice for this project. Here is the reasoning:

**Why Python wins here:**
- Zero external dependencies needed — `urllib`, `json`, `argparse`, `datetime` are all stdlib
- AI agents (Claude, GPT-based tools, LangChain agents) almost universally support subprocess calls to Python scripts
- `jq`-style post-processing is trivial with list comprehensions
- Runs identically on macOS, Linux, and Windows without compilation
- The normalizer/transformer logic is readable and easy to maintain

**When you might choose differently:**

| Language | Choose if... | Tradeoff |
|---|---|---|
| **TypeScript/Node** | The tool will live inside a JS/TS monorepo or be published to npm | Adds `node_modules`; heavier distribution |
| **Go** | You need a single compiled binary with zero runtime deps | Harder to iterate on; overkill for an internal tool |
| **Bash + curl** | Absolute minimal tooling in a CI/CD pipeline | No normalization, no retry logic, hard to maintain |
| **Ruby** | Team already uses Ruby for tooling | Fine, but no advantage over Python here |

**Decision rule:** If you can run `python3 --version` and get `3.10+`, use Python. If the tool must be distributed as a binary to non-developer users, use Go.

---

## 2. Prerequisites & Environment Setup

### What you need installed

```bash
# Check Python version (must be 3.10+)
python3 --version

# Check curl (for manual API validation in Phase 1)
curl --version

# Check jq (optional, for pretty-printing JSON in terminal)
jq --version   # install: brew install jq  /  apt install jq
```

### No pip installs required

The `sesc.py` tool uses **Python standard library only**:

```
urllib.request   — HTTP calls
urllib.parse     — URL encoding
urllib.error     — HTTP/network errors
json             — parse and serialize JSON
argparse         — CLI argument parsing
datetime         — date arithmetic
sys              — stdout/stderr routing
time             — retry delays
```

If you ever extend the tool to add features like HTML stripping, caching, or async requests, these libraries are candidates:

```bash
pip install httpx          # async HTTP, better than urllib for high-volume
pip install rich           # beautiful terminal output
pip install diskcache       # local caching to avoid repeat API calls
```

### Project structure

```
sesc-sp-cli/
├── sesc.py                  # main CLI tool (single file)
├── SKILL.md                 # AI agent skill definition
├── DEVELOPMENT_GUIDE.md     # this document
├── tests/
│   ├── test_normalizers.py  # unit tests for data transforms
│   ├── test_cli.py          # integration tests
│   └── fixtures/
│       ├── sample_units.json
│       ├── sample_events.json
│       └── sample_discover.json
└── .env.example             # if you add config later
```

---

## 3. Phase 0 — API Reconnaissance

**Do this before writing any code.** Skipping this phase is the most common reason developers build against the wrong endpoint and get empty responses.

### Step 0.1 — Open the SESC SP website in Chrome

Navigate to `https://www.sescsp.org.br/programacao/`

### Step 0.2 — Open DevTools Network tab

```
Chrome: F12 → Network tab → check "Preserve log" → check "XHR/Fetch" filter
Firefox: F12 → Network → XHR
```

### Step 0.3 — Interact with the filters

Click on any audience filter (Adulto, Criança, etc.), any unit, any activity type. Watch the Network panel.

### Step 0.4 — Find the pattern

You will see requests to:

```
https://www.sescsp.org.br/wp-json/wp/v1/dinamico?...
https://www.sescsp.org.br/wp-json/wp/v1/atividades/filter?...
```

Right-click each request → "Copy as cURL". Save them in a scratchpad. This gives you:
- The exact headers the browser sends
- The exact query parameter names and values
- The correct order of calls

### Step 0.5 — Save a sample response for each endpoint

For each request, right-click → "Copy Response". Save to `/tests/fixtures/`. These become your test fixtures.

```bash
# Filenames to save
tests/fixtures/raw_units.json          # from /dinamico?modes=unidade
tests/fixtures/raw_categories.json     # from /dinamico?modes=tipos_linguagens
tests/fixtures/raw_activity_types.json # from /dinamico?modes=acesso
tests/fixtures/raw_events.json         # from /atividades/filter
```

> **Why this matters:** The API has no documentation. Your fixtures are your specification. When the API changes, comparing new responses against these fixtures tells you exactly what broke.

---

## 4. Phase 1 — Validate the API is Reachable

Before writing Python, confirm each endpoint responds correctly using `curl`. This eliminates network/auth issues before they become code issues.

### Test 1 — Units endpoint

```bash
curl -s \
  -H "accept: application/json" \
  -H "referer: https://www.sescsp.org.br/programacao/" \
  -H "user-agent: Mozilla/5.0" \
  "https://www.sescsp.org.br/wp-json/wp/v1/dinamico?unidades=&categorias=&gratuito=&online=&publico_tag=adulto&tipos_atividades=&tipos_linguagens=&modes=unidade" \
  | jq 'length'
```

**Expected:** A number greater than 0 (the count of units in the response).

**If you get 0 or an error:** Check your headers. The API returns empty results without the correct `referer` header.

### Test 2 — Events endpoint (minimal)

```bash
curl -s \
  -H "accept: application/json" \
  -H "referer: https://www.sescsp.org.br/programacao/" \
  -H "user-agent: Mozilla/5.0" \
  "https://www.sescsp.org.br/wp-json/wp/v1/atividades/filter?local=&categoria=&gratuito=&online=&publico=adulto&atividade=&linguagem=&data_inicial=&data_final=&tipo=atividade&dinamico=true&ppp=5&page=1" \
  | jq '.'
```

**Expected:** A JSON object or array with event data.

**If you get HTML:** The server is returning a WordPress error page. Your headers are probably wrong — check the `referer` matches exactly.

### Test 3 — Confirm pagination works

```bash
# Page 1 — first 5 events
curl -s -H "accept: application/json" -H "referer: https://www.sescsp.org.br/programacao/" -H "user-agent: Mozilla/5.0" \
  "https://www.sescsp.org.br/wp-json/wp/v1/atividades/filter?publico=adulto&tipo=atividade&dinamico=true&ppp=5&page=1" \
  | jq '.[0].ID // .[0].id // "no id found"'

# Page 2 — next 5 events (IDs must be different from page 1)
curl -s -H "accept: application/json" -H "referer: https://www.sescsp.org.br/programacao/" -H "user-agent: Mozilla/5.0" \
  "https://www.sescsp.org.br/wp-json/wp/v1/atividades/filter?publico=adulto&tipo=atividade&dinamico=true&ppp=5&page=2" \
  | jq '.[0].ID // .[0].id // "no id found"'
```

**Expected:** Different IDs between pages. If the IDs are the same, pagination is not working.

### Phase 1 pass criteria

- [ ] Units endpoint returns JSON (not HTML, not empty)
- [ ] Events endpoint returns a non-empty list
- [ ] Page 2 returns different records than page 1
- [ ] No 403 or 401 errors

---

## 5. Phase 2 — Map the Discovery Pipeline

Now that you have raw fixtures, understand the dependency chain. **You cannot skip steps — each endpoint needs IDs from the previous one to work correctly.**

### The pipeline in detail

```
Step 1: /dinamico?modes=unidade
        → Returns: list of venue objects with IDs (e.g. ID=2 for Pinheiros)
        → You need: the comma-separated list of all IDs

Step 2: /dinamico?modes=tipos_linguagens&unidades=2,43,51,...
        → Returns: available category/language slugs FOR those units
        → You need: the slug values (e.g. "cinema", "teatro")

Step 3: /dinamico?modes=acesso&unidades=2,43,51,...
        → Returns: access/activity type metadata for those units
        → Optional for basic use, but needed for --activity-type filter

Step 4: /atividades/filter?local=2,43,51,...&publico=adulto&...
        → Returns: the actual events, paginated
        → This is the payload you'll expose to users/agents
```

### How to validate the chain manually

```bash
# Step 1: Get unit IDs
UNIT_IDS=$(curl -s \
  -H "accept: application/json" -H "referer: https://www.sescsp.org.br/programacao/" -H "user-agent: Mozilla/5.0" \
  "https://www.sescsp.org.br/wp-json/wp/v1/dinamico?modes=unidade&publico_tag=adulto" \
  | jq -r '[.[].term_id // .[].id // .[].ID] | map(tostring) | join(",")')

echo "Unit IDs: $UNIT_IDS"

# Step 2: Use those IDs to get events
curl -s \
  -H "accept: application/json" -H "referer: https://www.sescsp.org.br/programacao/" -H "user-agent: Mozilla/5.0" \
  "https://www.sescsp.org.br/wp-json/wp/v1/atividades/filter?local=${UNIT_IDS}&publico=adulto&tipo=atividade&dinamico=true&ppp=3&page=1" \
  | jq '[.[].post_title // .[].title]'
```

**Expected:** A list of event titles. If you see titles, the full pipeline is working.

---

## 6. Phase 3 — Validate Response Shapes

This is the most critical and most skipped step. The API returns inconsistent field names. Before writing normalizers, catalogue exactly what fields exist.

### Step 3.1 — Inspect a unit object

```bash
curl -s \
  -H "accept: application/json" -H "referer: https://www.sescsp.org.br/programacao/" -H "user-agent: Mozilla/5.0" \
  "https://www.sescsp.org.br/wp-json/wp/v1/dinamico?modes=unidade&publico_tag=adulto" \
  | jq '.[0] | keys'
```

Write down the field names you see. Map them to your normalized field names:

| API field (what you see) | Normalized field (what you'll output) |
|---|---|
| `term_id` or `ID` | `id` |
| `name` or `post_title` | `name` |
| `slug` or `post_name` | `slug` |
| `link` or `permalink` | `url` |

### Step 3.2 — Inspect an event object

```bash
curl -s \
  -H "accept: application/json" -H "referer: https://www.sescsp.org.br/programacao/" -H "user-agent: Mozilla/5.0" \
  "https://www.sescsp.org.br/wp-json/wp/v1/atividades/filter?publico=adulto&tipo=atividade&dinamico=true&ppp=1&page=1" \
  | jq '.[0] | keys'
```

Record every key. Common ones found:

```
post_title, post_content, post_excerpt, permalink,
data_inicio, data_fim, hora_inicio, hora_fim,
unidade_nome, gratuito, online, categorias,
tipo_atividade, publico, vagas, preco
```

### Step 3.3 — Check for null/empty variants

```bash
# Check how "gratuito" (free) is represented
curl -s [events_url] | jq '[.[].gratuito] | unique'
# Possible outputs: ["0","1"] or [null,"1"] or [true,false]
```

**Document every variant you find.** Your normalizer must handle all of them.

### Step 3.4 — Create your field map document

Create `tests/fixtures/field_map.json`:

```json
{
  "event": {
    "id": ["ID", "id", "post_id"],
    "title": ["post_title", "title", "nome"],
    "date_start": ["data_inicio", "date_start", "date"],
    "is_free_raw_values": ["0", "1", null, true, false, "gratuito"]
  },
  "unit": {
    "id": ["term_id", "ID", "id"],
    "name": ["name", "post_title"]
  }
}
```

This becomes documentation AND a test spec. Your normalizer code should mirror this map exactly.

---

## 7. Phase 4 — Build the Core CLI

Build in this order. Each step produces something testable.

### Step 4.1 — HTTP fetcher (30 min)

Build `fetch_json(url)` first. Requirements:
- Sets correct headers (referer, user-agent, accept)
- Implements retry with exponential backoff (3 attempts)
- Returns parsed JSON dict/list or `None` on failure
- All errors go to `stderr`, never `stdout`

**Test it:**
```python
# Paste into Python REPL
import sys
sys.path.insert(0, '.')
from sesc import fetch_json

result = fetch_json("https://www.sescsp.org.br/wp-json/wp/v1/dinamico?modes=unidade&publico_tag=adulto")
print(type(result), len(result) if result else 0)
# Expected: <class 'list'> 25  (or similar non-zero count)
```

### Step 4.2 — Normalizers (45 min)

Build `normalize_units(raw)` and `normalize_events(raw)` using your field map from Step 3.4.

Rules for normalizers:
- Try multiple field name variants with `or` chaining: `item.get('ID') or item.get('id')`
- Strip `None` values from output: `{k: v for k, v in event.items() if v is not None}`
- Convert `is_free` to a real boolean (not a string)
- Truncate and clean HTML from descriptions
- Never raise exceptions — return empty list on bad input

**Test it with your fixtures:**
```python
import json
from sesc import normalize_events

raw = json.load(open('tests/fixtures/raw_events.json'))
events = normalize_events(raw)

# Assertions
assert isinstance(events, list), "Must return list"
assert len(events) > 0, "Must not be empty"
assert 'title' in events[0], "Must have title"
assert isinstance(events[0].get('is_free'), bool), "is_free must be bool"
assert events[0].get('description') is None or len(events[0]['description']) <= 503, "Description must be truncated"
print(f"✅ {len(events)} events normalized correctly")
```

### Step 4.3 — API call functions (30 min)

Build `get_unidades()`, `get_categorias()`, `get_events()`. Each function:
- Constructs URL with `urllib.parse.urlencode`
- Calls `fetch_json()`
- Returns raw result (normalization is separate)

**Test each one:**
```python
from sesc import get_unidades, get_events

units = get_unidades()
assert units is not None, "Units must not be None"
print(f"✅ Units returned {len(units)} items")

events = get_events(limit=3, page=1)
assert events is not None
print(f"✅ Events returned")
```

### Step 4.4 — CLI commands (45 min)

Build `cmd_events`, `cmd_unidades`, `cmd_discover`, `cmd_today`. Each command:
- Accepts the parsed `args` namespace
- Calls the appropriate API function(s)
- Calls the normalizer
- Calls `output_json()` if `args.json` else `output_summary()`

**Test with subprocess (simulates real CLI use):**
```python
import subprocess, json

def run_cli(args_str):
    result = subprocess.run(
        f"python3 sesc.py {args_str}",
        shell=True, capture_output=True, text=True
    )
    return result.stdout, result.stderr, result.returncode

stdout, stderr, code = run_cli("unidades --json")
assert code == 0, f"Exit code should be 0, got {code}"
data = json.loads(stdout)
assert 'units' in data, "JSON must have 'units' key"
assert '_meta' in data, "JSON must have '_meta' key"
print("✅ unidades --json works")
```

### Step 4.5 — argparse setup (20 min)

Build `build_parser()`. Validate:
- All commands are registered as subparsers
- `--json` flag exists on every command
- `--compact` flag exists on every command
- `--help` on each subcommand shows correct options

```bash
python3 sesc.py --help
python3 sesc.py events --help
python3 sesc.py discover --help
# Should print usage without errors
```

---

## 8. Phase 5 — Validation Checklist

Run through every item before calling the tool "done."

### 5.1 — Functional validation

```bash
# Test 1: Minimal smoke test
python3 sesc.py events --limit 5 --json
# Expected: Valid JSON with _meta and events[] array

# Test 2: Free events filter
python3 sesc.py events --free --json | python3 -c "
import sys, json
data = json.load(sys.stdin)
non_free = [e for e in data['events'] if not e.get('is_free')]
assert len(non_free) == 0, f'Found {len(non_free)} non-free events'
print('✅ Free filter works')
"

# Test 3: Today shortcut
python3 sesc.py today --json
# Expected: Valid JSON (may have 0 events on quiet days — that's OK)

# Test 4: Discover pipeline
python3 sesc.py discover --json | python3 -c "
import sys, json
data = json.load(sys.stdin)
assert 'units' in data
assert 'events' in data
assert 'summary' in data
assert data['summary']['total_units'] > 0
print('✅ Discover pipeline works')
"

# Test 5: Pagination
PAGE1=$(python3 sesc.py events --limit 3 --page 1 --json | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['events'][0].get('id',''))")
PAGE2=$(python3 sesc.py events --limit 3 --page 2 --json | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['events'][0].get('id',''))")
[ "$PAGE1" != "$PAGE2" ] && echo "✅ Pagination works" || echo "❌ Pages return same data"
```

### 5.2 — stdout/stderr separation

```bash
# stdout must be clean JSON; stderr must have the log lines
python3 sesc.py events --limit 3 --json 2>/dev/null | python3 -m json.tool > /dev/null
echo "Exit code: $?"   # Must be 0

# Confirm logs go to stderr (not stdout)
STDERR=$(python3 sesc.py events --limit 3 --json 2>&1 1>/dev/null)
echo "Stderr output: $STDERR"   # Should show "ℹ️  Fetching events..."
```

### 5.3 — Exit codes

```bash
# Valid command: must exit 0
python3 sesc.py events --limit 1 --json
echo "Exit: $?"   # must be 0

# Invalid command: must exit non-zero
python3 sesc.py invalid_command 2>/dev/null
echo "Exit: $?"   # must be 2 (argparse error)
```

### 5.4 — JSON schema validation

Every response must have:

```python
import json, subprocess

def assert_valid_response(cmd):
    result = subprocess.run(f"python3 sesc.py {cmd}", shell=True, capture_output=True, text=True)
    assert result.returncode == 0, f"Non-zero exit: {result.returncode}"
    data = json.loads(result.stdout)  # must not throw
    assert '_meta' in data, "Missing _meta"
    assert 'fetched_at' in data['_meta'], "Missing fetched_at"
    assert 'total_returned' in data['_meta'] or 'total_units' in data.get('summary', {}), "Missing count"
    print(f"✅ {cmd}")

assert_valid_response("events --limit 5 --json")
assert_valid_response("unidades --json")
assert_valid_response("discover --json")
assert_valid_response("today --json")
```

### 5.5 — Resilience validation

```bash
# Test with no internet (should exit cleanly, not crash)
# Simulate by using a bad URL temporarily in the code
# Expected: error message on stderr, exit code 1, NO Python traceback on stdout
```

---

## 9. Phase 6 — AI Agent Compatibility Testing

These tests confirm the tool works correctly when called by an AI agent (Claude, GPT tools, LangChain, etc.).

### Test A — Can the agent parse the output?

```python
import subprocess, json

# Simulate what an AI agent would do
result = subprocess.run(
    ["python3", "sesc.py", "discover", "--json", "--compact"],
    capture_output=True, text=True
)

# Agent checks for success
assert result.returncode == 0

# Agent parses JSON
data = json.loads(result.stdout)

# Agent extracts what it needs
free_events = [e for e in data['events'] if e.get('is_free')]
online_events = [e for e in data['events'] if e.get('is_online')]
total = data['summary']['total_events']

print(f"Total: {total}, Free: {len(free_events)}, Online: {len(online_events)}")
print("✅ Agent can parse output")
```

### Test B — Agent workflow: "What's free this weekend?"

```python
import subprocess, json
from datetime import datetime, timedelta

today = datetime.now()
saturday = today + timedelta(days=(5 - today.weekday()) % 7)
sunday = saturday + timedelta(days=1)

result = subprocess.run([
    "python3", "sesc.py", "events",
    "--from", saturday.strftime("%Y-%m-%d"),
    "--to", sunday.strftime("%Y-%m-%d"),
    "--free", "--json", "--compact"
], capture_output=True, text=True)

data = json.loads(result.stdout)
events = data['events']

# Agent builds a response
for e in events[:5]:
    print(f"- {e.get('title')} @ {e.get('venue')} on {e.get('date_start')}")

print(f"\n✅ Weekend free events query works ({len(events)} found)")
```

### Test C — Token budget check (compact mode)

Agents have context window limits. Validate that `--compact` keeps output small.

```python
import subprocess, json

result = subprocess.run(
    ["python3", "sesc.py", "events", "--limit", "10", "--json", "--compact"],
    capture_output=True, text=True
)

byte_size = len(result.stdout.encode('utf-8'))
# 10 events compact should be well under 50KB
assert byte_size < 50_000, f"Output too large for agent context: {byte_size} bytes"
print(f"✅ Compact output size: {byte_size:,} bytes")
```

### Test D — Stderr is invisible to agent

```python
import subprocess, json

result = subprocess.run(
    ["python3", "sesc.py", "events", "--limit", "3", "--json"],
    capture_output=True, text=True
)

# stdout must be parseable JSON with no log noise
data = json.loads(result.stdout)  # this must not throw

# stderr may have logs — that's fine, it's separate
print("Stderr (logs, not visible to agent):", result.stderr[:100])
print("✅ stdout is clean JSON only")
```

---

## 10. Common Failure Modes

These are the bugs you will encounter. Here's how to detect each one.

### Failure 1 — Empty events list (`[]`)

**Symptom:** `events` array is always empty, no error.

**Cause:** Missing headers, especially `referer`. The API silently returns empty data without `https://www.sescsp.org.br/programacao/` as the referer.

**Debug:**
```bash
# Compare: with referer vs without
curl -s -H "referer: https://www.sescsp.org.br/programacao/" "URL..." | jq 'length'
curl -s "URL..." | jq 'length'
# If first returns >0 and second returns 0, referer is required
```

### Failure 2 — API returns HTML instead of JSON

**Symptom:** `json.JSONDecodeError` when parsing response.

**Cause:** WordPress is returning an error page. Check that `accept: application/json` header is present AND the URL is correct (no typo in path).

**Debug:**
```bash
curl -v "URL..." 2>&1 | grep "Content-Type"
# Should be: application/json
# If it says: text/html — you have a URL or header problem
```

### Failure 3 — Field name mismatch (KeyError)

**Symptom:** Some events parse correctly, others throw `KeyError` or return `None` for all fields.

**Cause:** The API uses different field names for different event types (e.g. regular activities vs. courses vs. cinema screenings).

**Debug:**
```bash
python3 sesc.py events --activity-type cinema --json | python3 -c "
import sys, json
data = json.load(sys.stdin)
print('Cinema event keys:', list(data['events'][0].keys()) if data['events'] else 'empty')
"

python3 sesc.py events --activity-type teatro --json | python3 -c "
import sys, json
data = json.load(sys.stdin)
print('Teatro event keys:', list(data['events'][0].keys()) if data['events'] else 'empty')
"
# Compare — different activity types may have different fields
```

### Failure 4 — Rate limiting (429 errors)

**Symptom:** Works for first few calls, then returns 429 or empty.

**Cause:** Too many requests too quickly.

**Fix:** Ensure retry delay is implemented in `fetch_json()`. For batch use, add `time.sleep(1.5)` between requests.

### Failure 5 — Pagination returns same results

**Symptom:** Page 2 returns same events as page 1.

**Cause:** `page` parameter is not being passed, or `ppp` (posts per page) is larger than total events.

**Debug:**
```bash
python3 sesc.py events --limit 3 --page 1 --json | python3 -c "import sys,json;d=json.load(sys.stdin);print([e.get('id') for e in d['events']])"
python3 sesc.py events --limit 3 --page 2 --json | python3 -c "import sys,json;d=json.load(sys.stdin);print([e.get('id') for e in d['events']])"
```

### Failure 6 — Date filter returns all events (filter not working)

**Symptom:** `--today` returns events from other days.

**Cause:** Date format is wrong. API expects `YYYY-MM-DD`. Check that `data_inicial` and `data_final` are not sent as empty strings when not needed.

**Debug:**
```bash
# Check what dates are in the response
python3 sesc.py events --today --json | python3 -c "
import sys, json
data = json.load(sys.stdin)
dates = list(set(e.get('date_start', 'no date') for e in data['events']))
print('Dates in response:', dates)
"
```

---

## 11. Maintenance & Monitoring

### How to know when the API has changed

The SESC SP website is a WordPress installation. It updates without notice.

**Set up a weekly validation script:**

```bash
#!/bin/bash
# save as: scripts/health_check.sh
# run weekly via cron: 0 9 * * 1 /path/to/health_check.sh

set -e

echo "Running SESC SP API health check..."

# 1. Units endpoint
UNIT_COUNT=$(python3 sesc.py unidades --json | python3 -c "import sys,json;print(json.load(sys.stdin)['_meta']['total'])")
[ "$UNIT_COUNT" -gt 0 ] && echo "✅ Units: $UNIT_COUNT" || echo "❌ Units: returned 0"

# 2. Events endpoint
EVENT_COUNT=$(python3 sesc.py events --limit 5 --json | python3 -c "import sys,json;print(json.load(sys.stdin)['_meta']['total_returned'])")
[ "$EVENT_COUNT" -gt 0 ] && echo "✅ Events: $EVENT_COUNT" || echo "❌ Events: returned 0"

# 3. Key fields present
python3 sesc.py events --limit 1 --json | python3 -c "
import sys, json
data = json.load(sys.stdin)
e = data['events'][0] if data['events'] else {}
required = ['title', 'date_start', 'venue', 'is_free']
missing = [f for f in required if f not in e]
if missing:
    print(f'❌ Missing fields: {missing}')
else:
    print('✅ All required fields present')
"
```

### When to update the normalizer

Update `normalize_events()` when:
- A new field appears in the API response that you want to expose
- An existing field changes its name (common after WordPress plugin updates)
- The `gratuito` or `online` fields change their value format

The safest way to detect this: periodically diff a fresh API response against your saved fixture:

```bash
# Save fresh fixture
python3 -c "
from sesc import get_events
import json
raw = get_events(limit=1, page=1)
print(json.dumps(list(raw[0].keys()) if isinstance(raw, list) else list(raw.keys()), indent=2))
" > /tmp/fresh_keys.json

# Compare with saved fixture
diff tests/fixtures/field_map.json /tmp/fresh_keys.json
```

---

## Quick Reference: Validation Commands

Copy-paste these to verify a fresh installation:

```bash
# 1. Environment check
python3 --version && echo "✅ Python OK" || echo "❌ Python missing"
curl -s "https://www.sescsp.org.br" > /dev/null && echo "✅ Network OK" || echo "❌ No network"

# 2. API reachability
python3 sesc.py unidades --json | python3 -m json.tool > /dev/null && echo "✅ API OK" || echo "❌ API unreachable"

# 3. Core functions
python3 sesc.py events --limit 3 --json | python3 -c "import sys,json;d=json.load(sys.stdin);print(f'✅ {len(d[\"events\"])} events returned')"

# 4. Filters
python3 sesc.py events --free --limit 3 --json | python3 -c "import sys,json;d=json.load(sys.stdin);assert all(e['is_free'] for e in d['events']);print('✅ Free filter OK')"

# 5. Full pipeline
python3 sesc.py discover --json --compact | python3 -c "import sys,json;d=json.load(sys.stdin);print(f'✅ Discover: {d[\"summary\"][\"total_units\"]} units, {d[\"summary\"][\"total_events\"]} events')"
```

All 5 must print ✅ before the tool is considered ready for use.

---

*Document version: 1.0 — last updated based on SESC SP API as observed in May 2025. Field names and API behavior are subject to change without notice. See [Maintenance & Monitoring](#11-maintenance--monitoring) for how to detect and respond to changes.*
