| Mode | Scenario | Payload | % of Full | Savings | Apply Time |
|---|---|---:|---:|---:|---:|
| **full** | Full regeneration (baseline) | 8,164 B | 100.0% | — | 1 ns |
| **diff** | 1 value change | 12 B | 0.1% | **99.9%** | 1.5 µs |
| **diff** | 4 value changes | 50 B | 0.6% | **99.4%** | 3.5 µs |
| **section** | 1 section replaced | 441 B | 5.4% | **94.6%** | 1.4 µs |
| **section** | 2 sections replaced | 516 B | 6.3% | **93.7%** | 3.8 µs |
| **template** | 8 slot bindings | 141 B | 1.7% | **98.3%** | 2.6 µs |
| **manifest** | 4 sections assembled | 487 B | 6.0% | **94.0%** | 2.4 µs |
