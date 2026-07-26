package liveeval

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

func TestRunGAPFlowWithMockProvider(t *testing.T) {
	root := t.TempDir()
	expDir := filepath.Join(root, "001-html-smoke")
	writeFixture(t, filepath.Join(expDir, "README.md"), "# Experiment: smoke\n\n**Format:** text/html | **Size:** small | **Edits:** 1\n\n**Expected sections:** msg\n")
	writeFixture(t, filepath.Join(expDir, "inputs/base/system.md"), "You produce HTML.")
	writeFixture(t, filepath.Join(expDir, "inputs/base/turn-0.md"), "Create greeting.")
	writeFixture(t, filepath.Join(expDir, "inputs/base/turn-1.md"), "Change hello to world.")
	writeFixture(t, filepath.Join(expDir, "inputs/gap/init-system.md"), "Create marked HTML.")
	writeFixture(t, filepath.Join(expDir, "inputs/gap/maintain-system.md"), "Return a GAP edit envelope.")

	var calls atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		var req map[string]any
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		call := calls.Add(1)
		switch call {
		case 1:
			writeChatResponse(t, w, `<gap:target id="msg">hello</gap:target>`, 10, 8)
		case 2:
			if _, ok := req["response_format"]; !ok {
				t.Fatal("GAP edit request missing response_format")
			}
			writeChatResponse(t, w, `{"protocol":"gap/0.1","id":"001-html-smoke","version":2,"name":"edit","meta":{"format":"text/html"},"content":[{"op":"replace","target":{"type":"id","value":"msg"},"content":"world"}]}`, 12, 6)
		default:
			t.Fatalf("unexpected call %d", call)
		}
	}))
	defer server.Close()

	err := Run(context.Background(), Config{
		ExperimentsDir: root,
		Flow:           "gap",
		Model:          "mock",
		APIBase:        server.URL,
		APIKey:         "test-key",
		Force:          true,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	output, err := os.ReadFile(filepath.Join(expDir, "outputs/gap/turn-1.html"))
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if !strings.Contains(string(output), "world") {
		t.Fatalf("output = %q", output)
	}
	metrics, err := os.ReadFile(filepath.Join(expDir, "metrics.json"))
	if err != nil {
		t.Fatalf("read metrics: %v", err)
	}
	if !strings.Contains(string(metrics), `"apply_success_rate": 1`) {
		t.Fatalf("metrics missing apply success: %s", metrics)
	}
}

func TestRunGAPFlowRepairsInvalidTarget(t *testing.T) {
	root := t.TempDir()
	expDir := filepath.Join(root, "001-html-smoke")
	writeFixture(t, filepath.Join(expDir, "README.md"), "# Experiment: smoke\n\n**Format:** text/html | **Size:** small | **Edits:** 1\n\n**Expected sections:** msg\n")
	writeFixture(t, filepath.Join(expDir, "inputs/base/system.md"), "You produce HTML.")
	writeFixture(t, filepath.Join(expDir, "inputs/base/turn-0.md"), "Create greeting.")
	writeFixture(t, filepath.Join(expDir, "inputs/base/turn-1.md"), "Change hello to world.")
	writeFixture(t, filepath.Join(expDir, "inputs/gap/init-system.md"), "Create marked HTML.")
	writeFixture(t, filepath.Join(expDir, "inputs/gap/maintain-system.md"), "Return a GAP edit envelope.")

	var calls atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		call := calls.Add(1)
		switch call {
		case 1:
			writeChatResponse(t, w, `<gap:target id="msg">hello</gap:target>`, 10, 8)
		case 2:
			messages := req["messages"].([]any)
			last := messages[len(messages)-1].(map[string]any)["content"].(string)
			if !strings.Contains(last, "list_targets() returned") || !strings.Contains(last, "- msg: hello") {
				t.Fatalf("edit request missing target inventory: %s", last)
			}
			writeChatResponse(t, w, `{"protocol":"gap/0.1","id":"001-html-smoke","version":2,"name":"edit","meta":{"format":"text/html"},"content":[{"op":"replace","target":{"type":"id","value":"missing"},"content":"world"}]}`, 12, 6)
		case 3:
			messages := req["messages"].([]any)
			last := messages[len(messages)-1].(map[string]any)["content"].(string)
			if !strings.Contains(last, "Supervisor Tool Error") || !strings.Contains(last, `validate_target("missing") failed`) {
				t.Fatalf("repair request missing validation error: %s", last)
			}
			writeChatResponse(t, w, `{"protocol":"gap/0.1","id":"001-html-smoke","version":2,"name":"edit","meta":{"format":"text/html"},"content":[{"op":"replace","target":{"type":"id","value":"msg"},"content":"world"}]}`, 14, 7)
		default:
			t.Fatalf("unexpected call %d", call)
		}
	}))
	defer server.Close()

	err := Run(context.Background(), Config{
		ExperimentsDir: root,
		Flow:           "gap",
		Model:          "mock",
		APIBase:        server.URL,
		APIKey:         "test-key",
		Force:          true,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	output, err := os.ReadFile(filepath.Join(expDir, "outputs/gap/turn-1.html"))
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if !strings.Contains(string(output), "world") {
		t.Fatalf("output = %q", output)
	}
	metrics, err := os.ReadFile(filepath.Join(expDir, "metrics.json"))
	if err != nil {
		t.Fatalf("read metrics: %v", err)
	}
	if !strings.Contains(string(metrics), `"repair_attempts": 1`) {
		t.Fatalf("metrics missing repair attempt: %s", metrics)
	}
	if !strings.Contains(string(metrics), `"apply_success_rate": 1`) {
		t.Fatalf("metrics missing apply success: %s", metrics)
	}
}

func TestFillDerivedReportsMissEconomics(t *testing.T) {
	reason := "apply failed: missing target msg"
	metrics := Metrics{
		BaseTurn0: &TurnMetrics{InputTokens: 50, OutputTokens: 50},
		GAPTurn0:  &TurnMetrics{InputTokens: 60, OutputTokens: 60},
		DefaultFlow: &FlowMetrics{
			PerTurn: []TurnResult{
				{InputTokens: 100, OutputTokens: 100},
				{InputTokens: 100, OutputTokens: 100},
			},
			TotalInputTokens:  200,
			TotalOutputTokens: 200,
		},
		GAPFlow: &GapFlowMetrics{
			FlowMetrics: FlowMetrics{
				PerTurn: []TurnResult{
					{InputTokens: 20, OutputTokens: 10},
					{
						InputTokens:   20,
						OutputTokens:  5,
						Failed:        true,
						FailureReason: &reason,
						Extra:         map[string]any{"envelope_parsed": true, "apply_succeeded": false},
					},
				},
				TotalInputTokens:  40,
				TotalOutputTokens: 15,
			},
		},
	}

	fillDerived(&metrics)

	if metrics.Reliability == nil || metrics.Reliability.MissCount != 1 {
		t.Fatalf("reliability = %#v, want one miss", metrics.Reliability)
	}
	if metrics.Reliability.ApplyMissCount != 1 || metrics.Reliability.MissRate != 0.5 {
		t.Fatalf("reliability = %#v, want one apply miss at 50%%", metrics.Reliability)
	}
	if metrics.Economics == nil || metrics.Economics.FallbackAdjusted == nil {
		t.Fatalf("economics missing: %#v", metrics.Economics)
	}
	fallback := metrics.Economics.FallbackAdjusted
	if fallback.MissAttemptTotalTokens != 25 || fallback.FallbackRetryTotalTokens != 200 {
		t.Fatalf("fallback = %#v, want miss tax 25 and retry 200", fallback)
	}
	if fallback.TotalTokenSavingsPct != 36.3 {
		t.Fatalf("fallback total savings = %f, want 36.3", fallback.TotalTokenSavingsPct)
	}
	if metrics.Economics.Amortized == nil || metrics.Economics.Amortized.FallbackInitInclusiveTokenSavingsPct != 25 {
		t.Fatalf("amortized = %#v, want fallback init-inclusive savings 25", metrics.Economics.Amortized)
	}
}

func writeFixture(t *testing.T, path string, value string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(value), 0o600); err != nil {
		t.Fatal(err)
	}
}

func writeChatResponse(t *testing.T, w http.ResponseWriter, content string, promptTokens uint64, completionTokens uint64) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(map[string]any{
		"choices": []map[string]any{{
			"message": map[string]string{"content": content},
		}},
		"usage": map[string]any{
			"prompt_tokens":     promptTokens,
			"completion_tokens": completionTokens,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
}
