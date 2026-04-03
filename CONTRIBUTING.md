# Contributing

Contributions are welcome. Please follow these guidelines.

## Getting started

```sh
git clone https://github.com/urmzd/aap
cd aap
just build   # compile the Rust binary
just test    # run the smoke test
just bench   # run offline tokenizer benchmarks
```

## Project structure

```
aap/
├── src/
│   ├── main.rs               # CLI entry point, signal handling
│   ├── lib.rs                 # File watcher
│   └── telemetry.rs           # Tracing init, metrics collection, shutdown summary
├── evals/                       # LLM evaluation framework
│   ├── pyproject.toml           # Python dependencies (uv + maturin)
│   └── src/aap_evals/           # Eval CLI and harness
├── benches/watcher.rs         # Criterion benchmarks
├── justfile                   # Task recipes
└── .github/workflows/ci.yml   # CI (Rust build + test)
```

## Making changes

- **Rust binary** (`src/`): file watcher, apply engine, telemetry. Keep dependencies light.
- **Telemetry** (`src/telemetry.rs`): structured logging via `tracing`, metrics summary on shutdown.
- **Evals** (`evals/`): LLM evaluation framework for measuring token efficiency. See `evals/README.md`.
- **New recipes**: add them to `justfile` with a comment describing what they do.

## Pull requests

1. Fork the repo and create a branch from `main`.
2. Make your changes and ensure `just test` passes.
3. Open a pull request with a clear description of what changed and why.

## Code style

- Rust: `cargo fmt` before committing.

## License

By contributing you agree that your contributions will be licensed under the [Apache 2.0 License](LICENSE).
