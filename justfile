build:
    cargo build

test:
    cargo test

# Rust criterion micro-benchmarks (apply engine speed)
bench:
    cargo bench

# Sync workspace — builds Rust FFI via maturin + all Python packages
bind:
    uv sync

# Run conversation benchmark experiments (base vs GAP flows)
run count="0" model="" id="" provider="google" fallback="":
    uv run gap-evals run --provider {{provider}} $(if [ "{{model}}" != "" ]; then echo "--model {{model}}"; fi) $(if [ "{{count}}" != "0" ]; then echo "--count {{count}}"; fi) $(if [ "{{id}}" != "" ]; then echo "--id {{id}}"; fi) $(if [ "{{fallback}}" != "" ]; then echo "--fallback {{fallback}}"; fi)

# Generate markdown report from experiment metrics
report:
    uv run gap-evals report

# ── Go eval-cli ──────────────────────────────────────────────────────────

# Build Rust cdylib (C FFI target) + Go eval-cli binary
build-go: build
    cd apps/eval-cli && go build -ldflags "-X github.com/urmzd/generative-artifact-protocol/eval-cli/cmd.Version=$(git describe --tags --always 2>/dev/null || echo dev)" -o ../../target/gap-eval-cli .

# Run Go tests for eval-cli
test-go: build
    cd apps/eval-cli && go test ./...

# Run experiments using the Go CLI
run-go count="0" model="" id="" provider="google" fallback="":
    ./target/gap-eval-cli run --provider {{provider}} --experiments-dir libs/evals/data/experiments $(if [ "{{model}}" != "" ]; then echo "--model {{model}}"; fi) $(if [ "{{count}}" != "0" ]; then echo "--count {{count}}"; fi) $(if [ "{{id}}" != "" ]; then echo "--id {{id}}"; fi) $(if [ "{{fallback}}" != "" ]; then echo "--fallback {{fallback}}"; fi)

# Generate report using the Go CLI
report-go:
    ./target/gap-eval-cli report --experiments-dir libs/evals/data/experiments
