# aap

> **Warning**: This project is `v0` — the protocol, schemas, and APIs are subject to breaking changes without notice until a formal release.

An open standard for token-efficient artifact generation, updates, and streaming — the **[Agent-Artifact Protocol (AAP)](spec/aap.md)**. The protocol defines how LLMs can declare, diff, and reprovision text artifacts with minimal token expenditure — 90-99% output token reduction per update, translating to 43-86% total cost savings depending on the model's pricing (see [cost model](spec/aap.md#811-cost-model)).

Includes a Rust reference implementation with a versioned artifact store, apply engine, and CLI tool for resolving protocol envelopes.

## How it works

1. An LLM produces an artifact envelope (JSON) declaring content, diffs, section updates, templates, or composites.
2. The apply engine resolves the envelope against a versioned store to produce the final text artifact.
3. The resolved artifact (HTML, SVG, source code, config, etc.) can be consumed by any downstream tool — browsers, PDF generators, IDEs, etc.

```
LLM ──produces──▶ envelope.json ──resolve──▶ artifact content
                                    ▲
                              aap (Rust apply engine)
```

> AAP produces text artifacts; rendering is a consumer responsibility.

## Requirements

- [Rust](https://rustup.rs/) (stable)
- [uv](https://github.com/astral-sh/uv) (Python package manager)
- [just](https://github.com/casey/just) (optional, for recipes)

## Quick start

```sh
# Build the binary
just build

# Resolve an envelope file
just resolve path/to/envelope.json

# Watch a file and resolve on changes
just watch path/to/artifact.html

# Stream a pre-built HTML dashboard
just demo

# Stream via a real LLM (requires ollama)
just demo-llm

# Stream via a HuggingFace tokenizer (gpt2 by default)
just demo-hf

# Run offline tokenizer benchmarks (no server needed)
just bench

# Run Rust criterion benchmarks (watcher, broadcast throughput)
just bench-rust
```

## CLI usage

```sh
aap <input> [--watch] [--output <file>]
```

- `<input>` — the file to read (or watch).
- `--watch` — keep watching and resolve on changes (runs until Ctrl+C).
- `--output` — write resolved content to a file instead of stdout.

When the input contains a protocol envelope (JSON with `"protocol": "aap/0.1"`), the binary resolves the envelope (applying diffs, section updates, templates, or composites) using the versioned artifact store. Plain files are passed through unchanged.

## Observability

The Rust binary emits structured log lines to stderr via `tracing` and prints a metrics summary on shutdown.

### Structured logging

Logs use `tracing-subscriber` (compact format with timestamps). Control verbosity with the `RUST_LOG` environment variable (default: `aap=info`).

```sh
# Show debug-level spans
RUST_LOG=aap=debug aap input.html
```

### Tracing spans

| Span | Location |
|---|---|
| `file_watcher` | File polling loop |

### Metrics summary

On Ctrl+C (watch mode) the binary prints a summary table to stderr:

```
── Metrics Summary ───────────────────────────────────
envelope.apply_count          5
envelope.apply_duration_ms    avg=0.3        min=0.1        max=0.8
watcher.changes_detected      5
watcher.poll_duration_ms      avg=0.1        min=0.0        max=0.3
broadcast.lag_count           0
───────────────────────────────────────────────────────
```

## Recipes

| Recipe | Description |
|---|---|
| `just build` | Compile the Rust binary |
| `just install` | Install the binary via `cargo install` |
| `just run [file]` | Compile and run the binary on a file |
| `just resolve <file>` | Resolve an envelope and print to stdout |
| `just watch [file]` | Watch a file and resolve on changes |
| `just demo` | Stream a pre-built HTML dashboard |
| `just demo-llm [model]` | Live ollama LLM streaming (default: gemma3) |
| `just demo-hf [tokenizer]` | HuggingFace tokenizer streaming |
| `just bench` | Offline Python tokenizer benchmarks |
| `just bench-rust` | Rust criterion benchmarks (watcher, broadcast) |
| `just bench-protocol` | Regenerate AAP benchmark table and embed into README |
| `just test` | Run Rust unit tests |

## Tools package

The Python scripts live under `tools/` and are structured as a proper package (`aap`) with console entry points:

| Entry point | Description |
|---|---|
| `ag-demo` | Stream a pre-built HTML dashboard in fixed chunks |
| `ag-ollama` | Stream a live LLM response via ollama |
| `ag-stream` | Generic file streaming utility |
| `ag-hf-stream` | Stream via a HuggingFace tokenizer |
| `ag-bench` | Offline benchmark: tokenize time, token count, throughput |
| `ag-realtime` | Real-time streaming dashboard |
| `ag-aap-demo` | AAP lifecycle demo (full → diff → section → template) |
| `ag-aap-bench` | Token savings benchmark across AAP generation modes |
| `ag-parallel-demo` | Parallel manifest generation demo (concurrent sections + assembly) |

Install and run any entry point with:

```sh
uv run --project tools ag-bench
```

## Cost model

AAP saves tokens by replacing full artifact regeneration with small diff envelopes. The savings are real but **LLM-dependent** — they vary with the model's tokenizer, output/input price ratio, and whether a cheaper model handles diffs. See the [full derivation in the spec](spec/aap.md#811-cost-model).

**The mechanism:** the maintain context reads the full artifact (S input tokens) and produces a diff envelope (d output tokens, where d is typically 1-5% of S). The apply engine resolves the diff at zero token cost (CPU, ~2μs). The orchestrator never reads the artifact at all — it holds only lightweight handles.

- **Output token reduction:** d instead of S per edit (95-99% fewer output tokens)
- **Context flattening:** no conversation history accumulates — each edit reads only the current artifact (S), not all prior versions (k·S at edit k in a naive conversation)
- **Model asymmetry:** the maintain context can use a cheaper model, multiplying savings further

**Concrete example** (2,000-token artifact, 30-token diff, r = p_out/p_in = 4×):

| After N edits | Naive conversation | AAP | Total savings |
|---:|---:|---:|---:|
| 1 | $0.071 | $0.039 | 45% |
| 5 | $0.304 | $0.070 | 77% |
| 10 | $0.763 | $0.107 | 86% |

At r = 1 (equal pricing), the same scenario yields ~49% savings after 10 edits. At r = 5, it reaches ~87%. The output token reduction is constant — what changes is how much of total cost it represents.

## AAP payload benchmarks

Payload size and apply time for each [Agent-Artifact Protocol (AAP)](spec/aap.md) generation mode, measured against an 8 KB HTML dashboard fixture. Regenerate with `cargo run --release --bin bench-table > benches/results.md`.

> **Note:** The "Payload savings" column measures **byte reduction** in the envelope payload — a proxy for output token reduction but not identical (tokenizers vary). Actual cost savings depend on the model's output/input price ratio; see [cost model](spec/aap.md#811-cost-model) for the full derivation.

<!-- embed-src src="benches/results.md" -->
| Mode | Scenario | Payload | % of Full | Payload savings | Apply Time |
|---|---|---:|---:|---:|---:|
| **full** | Full regeneration (baseline) | 8,164 B | 100.0% | — | 1 ns |
| **diff** | 1 value change | 12 B | 0.1% | **99.9%** | 1.5 µs |
| **diff** | 4 value changes | 50 B | 0.6% | **99.4%** | 3.5 µs |
| **section** | 1 section replaced | 441 B | 5.4% | **94.6%** | 1.4 µs |
| **section** | 2 sections replaced | 516 B | 6.3% | **93.7%** | 3.8 µs |
| **template** | 8 slot bindings | 141 B | 1.7% | **98.3%** | 2.6 µs |
| **manifest** | 4 sections assembled | 487 B | 6.0% | **94.0%** | 2.4 µs |
<!-- /embed-src -->

## Tokenizer benchmarks (example)

```
Tokenizer                    Tokens  Avg ch/tok    Tok ms   Tokens/sec
────────────────────────────────────────────────────────────────────────
gpt2                         27,300         2.4      14.2       1,917k
bert-base-uncased            33,628         1.9      16.1       2,094k
Fixed 30-char chunks          2,169        30.0       0.1      23,220k
```

## License

This project is dual-licensed:

- **Code** (`src/`, `tools/`, `benches/`, build files) — [Apache License 2.0](LICENSE)
- **Specification & docs** (`spec/`, `assets/`, documentation) — [CC-BY 4.0](LICENSE-CC-BY-4.0)

See [NOTICE](NOTICE) for details. Attribution is required under both licenses.
