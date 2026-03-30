# Agent-Artifact Protocol (AAP) Specification

**Version**: 1.0.0-draft
**Status**: Draft
**Date**: 2026-03-29

## 1. Introduction

Large language models regenerate entire artifacts on every edit — a report, a dashboard, a source file — even when only a single value changed. This wastes tokens, increases latency, and inflates cost.

The **Agent-Artifact Protocol (AAP)** is a portable, format-agnostic standard that defines how structured artifacts are declared, generated, updated, streamed, and reprovisioned with minimal token expenditure. Any LLM, agent framework, or rendering tool can implement it.

### 1.1 Design Goals

1. **Token efficiency** — express changes in the fewest tokens possible
2. **Format agnostic** — HTML, source code, JSON, YAML, Markdown, diagrams, configs
3. **Incremental by default** — full regeneration is the fallback, not the norm
4. **Streaming native** — every generation mode works over a stream
5. **Backward compatible** — raw content (no envelope) remains valid input
6. **Progressively adoptable** — conformance levels let implementations start simple

### 1.2 Relationship to Existing Standards

| Standard | Relationship |
|---|---|
| [RFC 6902](https://datatracker.ietf.org/doc/html/rfc6902) (JSON Patch) | Diff operations borrow semantics for JSON artifacts |
| [Unified Diff](https://www.gnu.org/software/diffutils/manual/html_node/Unified-Format.html) | Text diff operations use unified diff addressing |
| [Mustache](https://mustache.github.io/) | Template syntax is a subset of Mustache |
| [JSON Schema](https://json-schema.org/) | All protocol structures have machine-validatable schemas |

---

## 2. Terminology

| Term | Definition |
|---|---|
| **Artifact** | A discrete unit of structured content (an HTML page, a source file, a config) |
| **Envelope** | JSON wrapper carrying artifact metadata and content or operations |
| **Section** | A named, addressable region within an artifact |
| **Chunk** | A unit of streamed content within a chunk frame |
| **Generation** | The act of producing artifact content (initial creation or update) |
| **Reprovisioning** | Updating an existing artifact to a new version |
| **Token budget** | Maximum token allocation for a generation |
| **Flush point** | A semantically meaningful boundary where partial content can be rendered |
| **Producer** | The system generating artifacts (typically an LLM or agent) |
| **Consumer** | The system receiving, applying, and rendering artifacts |

---

## 3. Artifact Model

### 3.1 Envelope

Every protocol-aware payload is wrapped in an **envelope** — a JSON object with the following fields:

| Field | Type | Required | Description |
|---|---|---|---|
| `protocol` | string | YES | Protocol identifier. MUST be `"aap/1.0"` |
| `id` | string | YES | Unique artifact identifier (UUID or user-supplied) |
| `version` | integer | YES | Monotonically increasing version number (starts at 1) |
| `format` | string | YES | MIME type of the artifact content (`text/html`, `text/x-python`, `application/json`, etc.) |
| `mode` | string | YES | Generation mode: `"full"`, `"diff"`, `"section"`, `"template"`, `"composite"`, `"manifest"` |
| `encoding` | string | no | Character encoding. Default: `"utf-8"` |
| `base_version` | integer | no | Version this payload applies against. Required for `diff` and `section` modes |
| `created_at` | string | no | ISO 8601 timestamp of initial creation |
| `updated_at` | string | no | ISO 8601 timestamp of this version |
| `token_budget` | object | no | Token budget constraints (see [Section 7](#7-token-budgeting)) |
| `tokens_used` | integer | no | Actual tokens consumed to produce this payload |
| `checksum` | string | no | `sha256:<hex>` integrity hash of the resolved content |
| `sections` | array | no | Section definitions (see [Section 3.2](#32-sections)) |
| `content` | string | no | Artifact content (for `full` mode) |
| `operations` | array | no | Diff operations (for `diff` mode) |
| `target_sections` | array | no | Section updates (for `section` mode) |
| `template` | string | no | Template content or ID (for `template` mode) |
| `bindings` | object | no | Slot bindings (for `template` mode) |
| `includes` | array | no | Sub-artifact references (for `composite` mode) |
| `skeleton` | string | no | Static scaffold with section markers (for `manifest` mode) |
| `section_prompts` | array | no | Per-section generation instructions (for `manifest` mode) |
| `section_id` | string | no | Section this result fills (for parallel section results) |
| `content_encoding` | string | no | Compression: `"gzip"` or `"zstd"`. Applied to `content` field |

**Example** (minimal full-mode envelope):

```json
{
  "protocol": "aap/1.0",
  "id": "dashboard-001",
  "version": 1,
  "format": "text/html",
  "mode": "full",
  "content": "<!DOCTYPE html><html>...</html>"
}
```

### 3.2 Sections

An artifact MAY be divided into named **sections** — addressable regions that enable targeted updates.

| Field | Type | Required | Description |
|---|---|---|---|
| `id` | string | YES | Unique section identifier within the artifact |
| `label` | string | no | Human-readable label |
| `start_marker` | string | no | Format-specific start boundary |
| `end_marker` | string | no | Format-specific end boundary |

Section markers are format-specific:

| Format | Start marker | End marker |
|---|---|---|
| HTML | `<!-- section:id -->` | `<!-- /section:id -->` |
| Source code | `// #region id` | `// #endregion id` |
| Markdown | `<!-- section:id -->` | `<!-- /section:id -->` |
| JSON/YAML | N/A (use JSON Pointer paths) | N/A |

**Example** (HTML with sections):

```html
<!-- section:stats -->
<div class="stats">...</div>
<!-- /section:stats -->

<!-- section:users-table -->
<table>...</table>
<!-- /section:users-table -->
```

### 3.3 Version Chain

Every artifact maintains a version chain. Version numbers are monotonically increasing integers starting at 1. Each non-full update references its `base_version`. Consumers MUST reject updates where `base_version` does not match the current stored version (optimistic concurrency).

```
v1 (full) → v2 (diff, base_version=1) → v3 (section, base_version=2) → v4 (full)
```

A `full` mode envelope resets the chain — no `base_version` is required.

---

## 4. Generation Modes

The `mode` field declares how content is expressed. Producers SHOULD select the most token-efficient mode for the change at hand.

### 4.1 Full (`mode: "full"`)

Complete artifact content in the `content` field. This is the baseline — most expensive, always correct.

**When to use**: initial creation, major rewrites, or when diff overhead exceeds content size.

```json
{
  "protocol": "aap/1.0",
  "id": "report-42",
  "version": 1,
  "format": "text/html",
  "mode": "full",
  "content": "<html><body><h1>Q4 Report</h1>...</body></html>"
}
```

### 4.2 Diff (`mode: "diff"`)

Express changes as operations against a previous version. The `operations` array contains ordered operations applied sequentially.

**When to use**: small, localized changes (value updates, line insertions, deletions).

#### Operation Schema

| Field | Type | Required | Description |
|---|---|---|---|
| `op` | string | YES | `"replace"`, `"insert_before"`, `"insert_after"`, `"delete"` |
| `target` | object | YES | Addressing (see below) |
| `content` | string | no | New content (required for `replace`, `insert_before`, `insert_after`) |

#### Target Addressing

A target identifies where in the artifact the operation applies. Exactly one addressing mode MUST be used:

| Address mode | Fields | Description |
|---|---|---|
| Section | `{"section": "id"}` | Target an entire section by ID |
| Line range | `{"lines": [start, end]}` | Target lines (1-indexed, inclusive) |
| Offset range | `{"offsets": [start, end]}` | Target character offsets (0-indexed, exclusive end) |
| Search | `{"search": "literal text"}` | Target first occurrence of literal text |

**Example** (update a stat card value):

```json
{
  "protocol": "aap/1.0",
  "id": "dashboard-001",
  "version": 2,
  "format": "text/html",
  "mode": "diff",
  "base_version": 1,
  "operations": [
    {
      "op": "replace",
      "target": {"search": "<span class=\"stat-value\">$12,340</span>"},
      "content": "<span class=\"stat-value\">$15,720</span>"
    }
  ]
}
```

### 4.3 Section (`mode: "section"`)

Regenerate only targeted sections. All other sections are preserved from the base version.

**When to use**: one or a few sections need significant changes, but the rest is unchanged.

The `target_sections` array contains objects with:

| Field | Type | Required | Description |
|---|---|---|---|
| `id` | string | YES | Section ID to replace |
| `content` | string | YES | New content for this section |

**Example** (replace the users table):

```json
{
  "protocol": "aap/1.0",
  "id": "dashboard-001",
  "version": 3,
  "format": "text/html",
  "mode": "section",
  "base_version": 2,
  "target_sections": [
    {
      "id": "users-table",
      "content": "<table><tr><th>Name</th><th>Email</th></tr>...</table>"
    }
  ]
}
```

### 4.4 Template (`mode: "template"`)

Define a skeleton with named slots, then fill only the slots. Templates eliminate boilerplate regeneration.

**When to use**: generating variants of a known structure (dashboards with different data, reports with different periods, config files for different environments).

| Field | Type | Required | Description |
|---|---|---|---|
| `template` | string | YES | Template content with `{{slot_name}}` placeholders, or a registered template ID |
| `bindings` | object | YES | Map of slot name to content string |

Slot syntax follows [Mustache](https://mustache.github.io/):

- `{{name}}` — variable substitution (HTML-escaped by default)
- `{{{name}}}` — unescaped substitution
- `{{#items}}...{{/items}}` — iteration
- `{{#condition}}...{{/condition}}` — conditional block
- `{{^condition}}...{{/condition}}` — inverted conditional

**Example** (dashboard with different data):

```json
{
  "protocol": "aap/1.0",
  "id": "dashboard-001",
  "version": 4,
  "format": "text/html",
  "mode": "template",
  "template": "<!DOCTYPE html><html><body><h1>{{title}}</h1><div class='stats'>{{{stats_html}}}</div><div class='table'>{{{table_html}}}</div></body></html>",
  "bindings": {
    "title": "Q1 Dashboard",
    "stats_html": "<div class='stat'><span>Revenue</span><span>$15,720</span></div>",
    "table_html": "<table>...</table>"
  }
}
```

### 4.5 Composite (`mode: "composite"`)

Assemble an artifact from referenced sub-artifacts or external URIs. Enables deduplication of shared components (headers, CSS, boilerplate).

**When to use**: artifacts that share components (common nav bars, shared CSS, reusable code modules).

The `includes` array contains ordered references:

| Field | Type | Description |
|---|---|---|
| `ref` | string | Reference to another artifact: `"artifact_id"` or `"artifact_id:section_id"` |
| `uri` | string | External URI to fetch content from |
| `content` | string | Inline content (fallback if ref/uri unavailable) |
| `hash` | string | Expected `sha256:<hex>` of resolved content |

Exactly one of `ref`, `uri`, or `content` MUST be present per include.

**Example**:

```json
{
  "protocol": "aap/1.0",
  "id": "full-page",
  "version": 1,
  "format": "text/html",
  "mode": "composite",
  "includes": [
    {"ref": "shared-header"},
    {"content": "<main><h1>Page Content</h1></main>"},
    {"ref": "shared-footer"}
  ]
}
```

### 4.6 Content Encoding (Compression)

Any mode MAY compress its content fields using `content_encoding`:

- `"gzip"` — gzip compression (RFC 1952)
- `"zstd"` — Zstandard compression (RFC 8878)

Compressed content MUST be base64-encoded in JSON. The `checksum` field, if present, applies to the **uncompressed** content.

---

## 5. Reprovisioning

Reprovisioning is the act of updating an existing artifact. The producer selects a strategy based on the scope of change.

### 5.1 Section-First Generation (Recommended)

Producers SHOULD emit section markers on the **initial full generation**. This incurs a small overhead (~2% extra tokens for markers) but enables all subsequent updates to use `section` or `diff` mode — typically saving 90-99% of tokens per update.

**Rationale**: the upfront cost of markers is amortized across every future update. After just one `section`-mode update, the total token spend is lower than two full regenerations.

**Guidelines for section placement**:
- Place section boundaries at **independently meaningful blocks** (navigation, stat cards, data tables, forms, sidebars)
- Aim for **5-15 sections** per artifact — too few limits granularity, too many adds overhead
- Each section should be **self-contained**: updating one section should not require changes to another
- Avoid nesting sections deeper than 2 levels

**Cost model** (N = number of future updates):
- Without sections: N full regenerations = N * full_tokens
- With sections: 1 full (with markers) + N section updates = full_tokens * 1.02 + N * section_tokens
- Break-even: 1 update (section_tokens is typically 1-10% of full_tokens)

### 5.2 Parallel Generation

When an artifact has well-defined sections, the initial generation can be **parallelized** — each section is generated by an independent agent/tool call running concurrently, then assembled into the final artifact. This reduces wall-clock latency proportionally to the number of parallel workers without increasing total token cost.

#### 5.2.1 Manifest

A **manifest** declares the artifact structure and section assignments before generation begins. It is an envelope with `mode: "manifest"`:

| Field | Type | Required | Description |
|---|---|---|---|
| `skeleton` | string | YES | Static scaffold with section markers (boilerplate, layout, shared CSS) |
| `section_prompts` | array | YES | Per-section generation instructions |

Each `section_prompt` entry:

| Field | Type | Required | Description |
|---|---|---|---|
| `id` | string | YES | Section ID (matches marker in skeleton) |
| `prompt` | string | YES | Generation instruction for this section |
| `dependencies` | array | no | Section IDs that must complete before this one starts |
| `token_budget` | integer | no | Max tokens for this section |

**Example** (manifest for a dashboard):

```json
{
  "protocol": "aap/1.0",
  "id": "dashboard-001",
  "version": 1,
  "format": "text/html",
  "mode": "manifest",
  "skeleton": "<!DOCTYPE html>\n<html>\n<head>{{head}}</head>\n<body>\n<!-- section:nav --><!-- /section:nav -->\n<!-- section:stats --><!-- /section:stats -->\n<!-- section:users --><!-- /section:users -->\n<!-- section:orders --><!-- /section:orders -->\n</body>\n</html>",
  "section_prompts": [
    {"id": "nav", "prompt": "Generate a navigation bar with logo and user menu"},
    {"id": "stats", "prompt": "Generate 4 stat cards: users, revenue, orders, uptime"},
    {"id": "users", "prompt": "Generate a users table with 50 rows"},
    {"id": "orders", "prompt": "Generate an orders table with 30 rows", "dependencies": ["stats"]}
  ]
}
```

#### 5.2.2 Orchestration Flow

```
                    ┌─── Agent 1 ──▶ nav section ───────┐
Manifest ──parse──▶ ├─── Agent 2 ──▶ stats section ─────┼──▶ Assembler ──▶ Full Artifact
                    ├─── Agent 3 ──▶ users section ──────┤
                    └─── Agent 4 ──▶ orders section ─────┘
                                     (waits for stats)
```

1. **Parse manifest**: extract skeleton and section prompts
2. **Dispatch**: launch one generation per section, respecting `dependencies`
3. **Collect**: each agent returns a section envelope (`mode: "full"`, scoped to its section)
4. **Assemble**: stitch section content into the skeleton at marker positions
5. **Store**: save the assembled artifact as version 1 with all section markers intact

Sections without `dependencies` run concurrently. Sections with dependencies wait for their prerequisites to complete before starting.

#### 5.2.3 Section Results

Each parallel agent returns a **section result** — a lightweight envelope:

```json
{
  "protocol": "aap/1.0",
  "id": "dashboard-001",
  "version": 1,
  "format": "text/html",
  "mode": "full",
  "section_id": "stats",
  "content": "<div class=\"stats\">...</div>",
  "tokens_used": 450
}
```

The `section_id` field identifies which section this result fills. The assembler collects all section results and inserts each between its markers in the skeleton.

#### 5.2.4 Latency and Cost Model

| Strategy | Wall-clock latency | Total tokens | Tool calls |
|---|---|---|---|
| Sequential full | sum(section_times) | full_tokens | 1 |
| Parallel sections | max(section_times) | full_tokens + manifest_overhead | N + 1 |
| Parallel + update | max(section_times) + update_time | full_tokens + section_tokens | N + 2 |

**Manifest overhead** is minimal — the skeleton and prompts are typically 5-10% of the full artifact tokens. The latency win is significant: a 4-section artifact generated in parallel completes in ~25% of sequential wall-clock time.

#### 5.2.5 Parallel Updates

The same pattern applies to updates: when multiple sections need regeneration, dispatch them in parallel:

```json
{
  "protocol": "aap/1.0",
  "id": "dashboard-001",
  "version": 2,
  "format": "text/html",
  "mode": "manifest",
  "base_version": 1,
  "section_prompts": [
    {"id": "stats", "prompt": "Update stat cards with Q2 data"},
    {"id": "orders", "prompt": "Add 10 new order rows"}
  ]
}
```

Only the listed sections are regenerated. The assembler merges results into the existing artifact, preserving unchanged sections from the base version.

### 5.3 Strategy Selection Guide

| Change scope | Recommended mode | Token savings |
|---|---|---|
| Single value change | `diff` (search/replace) | ~95-99% |
| Few lines changed | `diff` (line range) | ~90-98% |
| One section rewritten | `section` | ~80-95% |
| Multiple sections rewritten | `section` | ~50-80% |
| Same structure, different data | `template` | ~90-98% |
| Complete rewrite | `full` | 0% (baseline) |

### 5.4 Version Chain Integrity

1. Each update envelope MUST include `base_version` (except `full` mode)
2. The consumer MUST verify `base_version` matches its current version
3. On mismatch: reject the update, notify the producer of the current version
4. The producer SHOULD re-derive its update against the correct base

### 5.5 Rollback

Consumers SHOULD maintain a configurable version history (default: 10 versions). Rollback replaces the current content with a previous version and increments the version number.

---

## 6. Streaming Protocol

Streaming is orthogonal to generation mode — any mode can be streamed. Streamed payloads are delivered as **JSONL** (one JSON object per line).

### 6.1 Chunk Frame

Each streamed unit is a **chunk frame**:

| Field | Type | Required | Description |
|---|---|---|---|
| `seq` | integer | YES | Monotonically increasing sequence number (starts at 0) |
| `content` | string | YES | Chunk payload |
| `section_id` | string | no | Section being streamed (if applicable) |
| `flush` | boolean | no | Hint to render/apply accumulated content. Default: `false` |
| `final` | boolean | no | `true` on the last chunk. Default: `false` |

The first chunk frame (`seq: 0`) SHOULD include the envelope metadata (all fields except `content`) in an `envelope` field. Subsequent frames carry only chunk data.

**Example** (streaming a full-mode artifact):

```jsonl
{"seq":0,"envelope":{"protocol":"aap/1.0","id":"doc-1","version":1,"format":"text/html","mode":"full"},"content":"<!DOCTYPE html><html>","flush":true,"final":false}
{"seq":1,"content":"<head><title>Report</title></head>","flush":true,"final":false}
{"seq":2,"content":"<body><h1>Q4 Report</h1>","flush":false,"final":false}
{"seq":3,"content":"<p>Revenue increased by 15%.</p></body></html>","flush":true,"final":true}
```

### 6.2 Flush Strategies

Producers SHOULD emit `flush: true` at semantically meaningful boundaries:

| Strategy | Description | Flush overhead | Render quality |
|---|---|---|---|
| **Token-aligned** | Flush every token | High | Smooth but expensive |
| **Syntax-aligned** | Flush at tag/statement boundaries | Low | Clean partial renders |
| **Size-aligned** | Flush every N bytes | Low | May split mid-tag |
| **Adaptive** | Start small (responsiveness), grow chunks over time | Low | Best overall |

**Recommended**: adaptive strategy with syntax-aligned flush points.

### 6.3 Transport

The protocol is transport-agnostic. Reference transports:

| Transport | Description |
|---|---|
| **File write + poll** | Write JSONL to a file; consumer polls for changes |
| **Server-Sent Events** | Each chunk frame is an SSE `data:` line |
| **WebSocket** | Each chunk frame is a WebSocket text message |
| **stdio** | Each chunk frame is a line on stdout |

---

## 7. Token Budgeting

### 7.1 Budget Declaration

The `token_budget` object in the envelope declares constraints:

| Field | Type | Description |
|---|---|---|
| `max_tokens` | integer | Maximum content tokens (excludes envelope overhead) |
| `priority` | string | `"completeness"` (prefer full content), `"brevity"` (prefer concise), `"fidelity"` (prefer accuracy) |
| `max_sections` | integer | Maximum sections to regenerate (for `section` mode) |

### 7.2 Budget Accounting

- **Content tokens**: tokens in the artifact payload (what the user sees)
- **Overhead tokens**: envelope metadata, framing, operation descriptions
- The budget applies to **content tokens only**
- Producers MUST NOT exceed `max_tokens`
- Producers SHOULD select the most token-efficient mode to stay within budget

### 7.3 Reporting

The final envelope (or final chunk frame) MUST include `tokens_used` — the actual content tokens consumed. This enables consumers to track token efficiency over time.

---

## 8. Conformance Levels

Implementations declare their conformance level. Each level is a superset of the previous.

### Level 0 — Basic

- MUST parse and produce valid envelopes
- MUST support `mode: "full"`
- MUST validate `protocol` field

### Level 1 — Incremental

- Level 0, plus:
- MUST support `mode: "diff"` with all addressing modes (section, line, offset, search)
- MUST support `mode: "section"`
- MUST maintain version chain and enforce `base_version` concurrency

### Level 2 — Template

- Level 1, plus:
- MUST support `mode: "template"` with Mustache-subset slot syntax
- MUST support template registration (store and reuse by ID)

### Level 3 — Full Protocol

- Level 2, plus:
- MUST support `mode: "composite"` with ref, uri, and content includes
- MUST support `content_encoding` (gzip and zstd)
- MUST support streaming chunk frames (JSONL)
- MUST support token budgeting (`token_budget` and `tokens_used`)
- MUST support adaptive flush strategy

---

## 9. Security Considerations

- **Content injection**: consumers MUST sanitize artifact content before rendering in privileged contexts (e.g., web browsers)
- **URI resolution**: `composite` mode URIs MUST be validated against an allowlist; arbitrary URI fetch is a server-side request forgery (SSRF) risk
- **Checksum verification**: consumers SHOULD verify `checksum` when present to detect tampering or corruption
- **Token budget enforcement**: producers MUST NOT exceed declared budgets; consumers SHOULD reject payloads that claim to use fewer tokens than they actually contain

---

## 10. IANA Considerations

This specification does not require any IANA registrations. The `format` field uses existing MIME types.

---

## Appendix A: JSON Schemas

Machine-validatable schemas for all protocol structures are provided in the `schemas/` directory:

- [`artifact-envelope.json`](schemas/artifact-envelope.json) — Envelope schema
- [`diff-operation.json`](schemas/diff-operation.json) — Diff operation schema
- [`template-binding.json`](schemas/template-binding.json) — Template binding schema
- [`chunk-frame.json`](schemas/chunk-frame.json) — Streaming chunk frame schema

## Appendix B: Token Savings Reference

Empirical measurements from the reference implementation using a 40KB HTML dashboard artifact:

| Edit scenario | Full tokens | Diff tokens | Savings | Section tokens | Savings | Template tokens | Savings |
|---|---|---|---|---|---|---|---|
| Change 1 stat value | ~10,000 | ~50 | 99.5% | N/A | — | N/A | — |
| Add 5 table rows | ~10,000 | ~300 | 97.0% | ~1,000 | 90.0% | N/A | — |
| Update all CSS colors | ~10,000 | ~700 | 93.0% | ~1,500 | 85.0% | N/A | — |
| New data, same layout | ~10,000 | N/A | — | N/A | — | ~400 | 96.0% |

*Values are approximate; see `ag-aap-bench` for current measurements.*
