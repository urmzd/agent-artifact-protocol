<p align="center">
  <h1 align="center">Generative Artifact Protocol</h1>
  <p align="center">
    Token-efficient artifact updates and streaming for LLMs — 90-99% output token reduction per edit.
    <br /><br />
    <a href="https://crates.io/crates/generative-artifact-protocol">Crates.io</a>
    &middot;
    <a href="https://github.com/urmzd/generative-artifact-protocol/issues">Report Bug</a>
    &middot;
    <a href="spec/gap.md">Specification</a>
  </p>
</p>

<p align="center">
  <a href="https://github.com/urmzd/generative-artifact-protocol/actions/workflows/ci.yml"><img src="https://github.com/urmzd/generative-artifact-protocol/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://crates.io/crates/generative-artifact-protocol"><img src="https://img.shields.io/crates/v/generative-artifact-protocol" alt="crates.io"></a>
</p>

> **Warning**: This project is `v0` — the protocol, schemas, and APIs are subject to breaking changes without notice until a formal release.

An open standard protocol — **[GAP](spec/gap.md)** — that lets LLMs declare, diff, and reprovision text artifacts with minimal token expenditure. Includes a Rust reference implementation of the apply engine plus a Python evaluation framework for measuring token efficiency against real LLM runs.

## Features

- **Envelope system** — three operation types (`synthesize`, `edit`, `handle`) for full generation, targeted updates, and lightweight references
- **Stateless apply engine** — pure function, no I/O, ~2μs per edit; portable to browsers (WASM), IDEs, CLIs, or service backends
- **ID-based targeting** — `<gap:target id="ID">` markers and JSON Pointer paths eliminate hallucinated search strings
- **Format-agnostic** — works with HTML, Python, JavaScript, JSON, YAML, Rust, Go, SVG, and more
- **90-99% output token reduction** per edit, translating to 43-86% total cost savings ([cost model](spec/gap.md#811-cost-model))
- **SSE transport binding** — wire format for streaming with reconnection support ([GAP-SSE](spec/gap-sse.md))
- **Evaluation framework** — 89 experiment datasets measuring token efficiency and reliability against real LLM runs

## Install

**Rust crate:**

```sh
cargo add generative-artifact-protocol
```

**From source (full workspace):**

```sh
git clone https://github.com/urmzd/generative-artifact-protocol
cd generative-artifact-protocol
```

Requires [Rust](https://rustup.rs/) (stable), [uv](https://github.com/astral-sh/uv) (for evals), and optionally [just](https://github.com/casey/just) (for recipes).

## Quick Start

```sh
# Build the Rust library
just build

# Run tests
just test

# Run criterion benchmarks (apply engine speed)
just bench

# Sync workspace — build FFI via maturin + Python packages
just bind

# Run LLM evaluations
just run count=5 model="gemini-2.0-flash" provider="google"

# Generate report from experiment metrics
just report
```

## Usage

### How it works

```
LLM ──produces──▶ envelope ──apply──▶ (artifact, handle)
                                 ▲
                           gap (stateless, ~2μs)
```

1. An LLM produces an artifact envelope (JSON) — either a `synthesize` envelope (full content with target markers) or an `edit` envelope (targeted changes by ID or JSON Pointer).
2. The apply engine resolves the envelope against the current artifact state to produce the updated artifact and a lightweight handle.
3. The orchestrator holds handles; the resolved artifact is stored and consumed by downstream tools — browsers, IDEs, etc.

### Apply engine

The core of the library is a single stateless function:

```rust
pub fn apply(artifact: Option<&Artifact>, envelope: &Envelope) -> Result<(Artifact, Envelope)>
```

| Envelope | Direction | Description |
|---|---|---|
| **synthesize** | input | Complete artifact content (baseline or reset) with `<gap:target>` markers |
| **edit** | input | Targeted changes via ID (`<gap:target>` markers) or JSON Pointer |
| **handle** | output | Lightweight reference returned after every synthesize or edit |

### Recipes

| Recipe | Description |
|---|---|
| `just build` | Compile the Rust library |
| `just test` | Run Rust unit tests |
| `just bench` | Criterion micro-benchmarks (apply engine speed) |
| `just bind` | Sync workspace — build FFI via maturin + Python packages |
| `just run [count] [model] [id] [provider]` | Run conversation benchmark experiments (base vs GAP flows) |
| `just report` | Generate markdown report from experiment metrics |

### Cost model

GAP saves tokens by replacing full artifact regeneration with small diff envelopes. The savings vary with the model's tokenizer, output/input price ratio, and whether a cheaper model handles diffs. See the [full derivation in the spec](spec/gap.md#811-cost-model).

The maintain context reads the full artifact ($S$ input tokens) and produces an edit envelope ($d$ output tokens, where $d$ is typically 1–5% of $S$). The apply engine resolves the edit at zero token cost (CPU, ~2μs).

- **Output token reduction:** $d$ instead of $S$ per edit (95–99% fewer output tokens)
- **Context flattening:** each edit reads only the current artifact ($S$), not all prior versions ($k \cdot S$ at edit $k$)
- **Model asymmetry:** the maintain context can use a cheaper model, multiplying savings further

**Example** (2,000-token artifact, 30-token edit, $r = p_{\text{out}}/p_{\text{in}} = 4\text{x}$):

| After $N$ edits | Naive conversation | GAP | Total savings |
|---:|---:|---:|---:|
| 1 | \$0.071 | \$0.039 | 45% |
| 5 | \$0.304 | \$0.070 | 77% |
| 10 | \$0.763 | \$0.107 | 86% |

### Payload benchmarks

Payload size and apply time for each envelope type, measured against an 8 KB HTML dashboard fixture.

> **Note:** "Payload savings" measures **byte reduction** — a proxy for output token reduction but not identical (tokenizers vary). See [cost model](spec/gap.md#711-cost-model) for the full derivation.

<!-- embed-src src="benches/results.md" -->
| Envelope | Scenario | Payload | % of Full | Payload savings | Apply Time |
|---|---|---:|---:|---:|---:|
| **synthesize** | Full generation (baseline) | 8,164 B | 100.0% | — | 1 ns |
| **edit** | 1 value replace (ID targeting) | 12 B | 0.1% | **99.9%** | 1.5 µs |
| **edit** | 4 value replaces (ID targeting) | 50 B | 0.6% | **99.4%** | 3.5 µs |
| **edit** | 1 section replace (ID targeting) | 441 B | 5.4% | **94.6%** | 1.4 µs |
| **edit** | 2 section replaces (ID targeting) | 516 B | 6.3% | **93.7%** | 3.8 µs |
<!-- /embed-src -->

## License

This project is dual-licensed:

- **Code** (`src/`, `evals/`, `benches/`, build files) — [Apache License 2.0](LICENSE)
- **Specification & docs** (`spec/`, `assets/`, documentation) — [CC-BY 4.0](LICENSE-CC-BY-4.0)

See [NOTICE](NOTICE) for details. Attribution is required under both licenses.
