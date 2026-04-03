package main

// Flow execution will be implemented when the `run` command is built.
// The `generate` command creates all experiment inputs without needing an LLM.
//
// The run command will:
// 1. Read inputs from experiment directories
// 2. Send them to the LLM via saige provider interface
// 3. Collect UsageDelta (input_tokens, output_tokens, latency)
// 4. Write outputs and metrics.json
