package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) {
	return 0, errors.New("read failed")
}

func (errorReader) Close() error { return nil }

func TestCallTripCodeSuccessAndHeaders(t *testing.T) {
	var gotKey string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("X-Demo-Key")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"tripcode":"TF-SUB-9DA70A7F98","issuer":"HUT","mode":"live","packet":{"article":{"title":"Hut 8"},"river":{"node_count":10}}}`))
	}))
	defer server.Close()

	resp, err := callTripCode(context.Background(), server.URL, "judge", tripcodeRequest{TripCode: "TF-SUB-9DA70A7F98"})
	if err != nil {
		t.Fatalf("callTripCode returned error: %v", err)
	}
	if gotKey != "judge" || resp.TripCode != "TF-SUB-9DA70A7F98" {
		t.Fatalf("unexpected response/header: key=%q resp=%#v", gotKey, resp)
	}
}

func TestCallTripCodeErrors(t *testing.T) {
	if _, err := callTripCode(context.Background(), "://bad", "", tripcodeRequest{}); err == nil {
		t.Fatal("expected build request error")
	}

	oldClient := demoHTTPClient
	t.Cleanup(func() { demoHTTPClient = oldClient })
	demoHTTPClient = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("network failed")
	})}
	if _, err := callTripCode(context.Background(), "http://example.test", "", tripcodeRequest{}); err == nil {
		t.Fatal("expected network error")
	}

	demoHTTPClient = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: errorReader{}, Header: make(http.Header)}, nil
	})}
	if _, err := callTripCode(context.Background(), "http://example.test", "", tripcodeRequest{}); err == nil {
		t.Fatal("expected read error")
	}

	demoHTTPClient = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusTeapot, Body: io.NopCloser(strings.NewReader("nope")), Header: make(http.Header)}, nil
	})}
	if _, err := callTripCode(context.Background(), "http://example.test", "", tripcodeRequest{}); err == nil || !strings.Contains(err.Error(), "HTTP 418") {
		t.Fatalf("expected HTTP error, got %v", err)
	}

	demoHTTPClient = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("{bad")), Header: make(http.Header)}, nil
	})}
	if _, err := callTripCode(context.Background(), "http://example.test", "", tripcodeRequest{}); err == nil {
		t.Fatal("expected decode error")
	}
}

func TestPrintFunctionsAndExtractors(t *testing.T) {
	oldStdout := demoStdout
	t.Cleanup(func() { demoStdout = oldStdout })
	var out bytes.Buffer
	demoStdout = &out

	resp := tripcodeResponse{
		TripCode:      "TF-SUB-9DA70A7F98",
		Issuer:        "HUT",
		Mode:          "live-mcp-tripcode-gemini",
		GeminiSummary: " summary ",
		Cost:          &costSnapshot{Enabled: true, Source: "local-estimate", RequestCostUSD: 0.03, TrackedSpentUSD: 0.03, BudgetUSD: 500, RemainingUSD: 499.97},
		Packet: map[string]any{
			"research_packet": map[string]any{
				"article": map[string]any{"headline": "Hut 8 Headline"},
				"river":   map[string]any{"node_count": float64(10)},
			},
		},
		Memory: &struct {
			SessionID    string `json:"session_id"`
			Available    bool   `json:"available"`
			Turns        int    `json:"turns"`
			LastTripCode string `json:"last_tripcode"`
			LastIssuer   string `json:"last_issuer"`
			Entries      []struct {
				ArticleTitle        string   `json:"article_title"`
				RiverNodeCount      int      `json:"river_node_count"`
				MonitorItems        []string `json:"monitor_items"`
				WeakenedAssumptions []string `json:"weakened_assumptions"`
			} `json:"entries"`
		}{
			SessionID:    "hut",
			Available:    true,
			Turns:        1,
			LastTripCode: "TF-SUB-9DA70A7F98",
			LastIssuer:   "HUT",
			Entries: []struct {
				ArticleTitle        string   `json:"article_title"`
				RiverNodeCount      int      `json:"river_node_count"`
				MonitorItems        []string `json:"monitor_items"`
				WeakenedAssumptions []string `json:"weakened_assumptions"`
			}{{
				ArticleTitle:        "Hut 8",
				RiverNodeCount:      10,
				MonitorItems:        []string{"power"},
				WeakenedAssumptions: []string{"timing"},
			}},
		},
	}
	printResolved(resp)
	if !strings.Contains(out.String(), "Gemini summary") || !strings.Contains(out.String(), "Cost: request=$0.0300") || !strings.Contains(out.String(), "Remembered River nodes: 10") {
		t.Fatalf("print output missing expected content: %s", out.String())
	}

	out.Reset()
	printCost(tripcodeResponse{Cost: &costSnapshot{Enabled: false}})
	if !strings.Contains(out.String(), "Cost: disabled") {
		t.Fatalf("missing disabled cost output: %s", out.String())
	}

	out.Reset()
	printCost(tripcodeResponse{})
	if out.String() != "" {
		t.Fatalf("nil cost should not print: %s", out.String())
	}

	out.Reset()
	printCost(tripcodeResponse{Cost: &costSnapshot{Enabled: true, RequestCostUSD: 0.01}})
	if !strings.Contains(out.String(), "source=local-estimate") {
		t.Fatalf("missing fallback cost source: %s", out.String())
	}

	out.Reset()
	printMemory(tripcodeResponse{})
	if !strings.Contains(out.String(), "not available") {
		t.Fatalf("missing no-memory output: %s", out.String())
	}

	out.Reset()
	resp.Memory.Entries = nil
	printMemory(resp)
	if !strings.Contains(out.String(), "turns=1") {
		t.Fatalf("missing empty entries output: %s", out.String())
	}

	if extractTitle(map[string]any{"title": " Direct "}) != "Direct" {
		t.Fatal("direct title not extracted")
	}
	if extractTitle(map[string]any{}) != "" {
		t.Fatal("empty title should be empty")
	}
	if extractRiverNodeCount(map[string]any{"river_nodes": []any{1, 2}}) != 2 {
		t.Fatal("river node array count not extracted")
	}
	if extractRiverNodeCount(map[string]any{"river": map[string]any{"node_count": 3}}) != 3 {
		t.Fatal("int node count not extracted")
	}
	if extractRiverNodeCount(map[string]any{}) != 0 || arrayLen("x") != 0 || numericInt("x") != 0 {
		t.Fatal("empty extractors should return zero")
	}
	if envOr("DELTASIGNAL_TEST_MISSING", "fallback") != "fallback" || fallback("", "fallback") != "fallback" || fallback("value", "fallback") != "value" {
		t.Fatal("fallback helpers failed")
	}
}

func TestMainSuccessAndExitf(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "/v1/tripcode") && r.Method == http.MethodPost {
			if strings.Contains(readBody(t, r), `"tripcode"`) {
				_, _ = w.Write([]byte(`{"tripcode":"TF-SUB-9DA70A7F98","issuer":"HUT","mode":"live","packet":{"article":{"title":"Hut 8"},"river":{"node_count":10}},"cost":{"enabled":true,"source":"local-estimate","request_cost_usd":0.03,"tracked_spent_usd":0.03,"budget_usd":500,"remaining_usd":499.97},"memory":{"session_id":"hut-demo","available":true,"turns":1,"last_tripcode":"TF-SUB-9DA70A7F98","last_issuer":"HUT","entries":[{"article_title":"Hut 8","river_node_count":10}]}}`))
				return
			}
			_, _ = w.Write([]byte(`{"tripcode":"TF-SUB-9DA70A7F98","issuer":"HUT","mode":"session-memory","packet":{},"cost":{"enabled":true,"source":"local-estimate","request_cost_usd":0,"tracked_spent_usd":0.03,"budget_usd":500,"remaining_usd":499.97},"memory":{"session_id":"hut-demo","available":true,"turns":1,"last_tripcode":"TF-SUB-9DA70A7F98","last_issuer":"HUT","entries":[{"article_title":"Hut 8","river_node_count":10}]}}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	oldStdout, oldStderr, oldExit, oldClient := demoStdout, demoStderr, demoExit, demoHTTPClient
	t.Cleanup(func() {
		demoStdout, demoStderr, demoExit, demoHTTPClient = oldStdout, oldStderr, oldExit, oldClient
	})
	var out bytes.Buffer
	demoStdout = &out
	demoStderr = io.Discard
	demoExit = func(code int) { t.Fatalf("unexpected exit %d", code) }
	demoHTTPClient = server.Client()
	t.Setenv("DELTASIGNAL_AGENT_URL", server.URL)
	t.Setenv("DELTASIGNAL_DEMO_API_KEY", "judge")
	t.Setenv("DELTASIGNAL_DEMO_TRIPCODE", "TF-SUB-9DA70A7F98")
	t.Setenv("DELTASIGNAL_DEMO_ISSUER", "HUT")
	t.Setenv("DELTASIGNAL_DEMO_SESSION_ID", "hut-demo")
	main()
	if !strings.Contains(out.String(), "Turn 1") || !strings.Contains(out.String(), "Turn 2") || !strings.Contains(out.String(), "Cost: request=$0.0300") {
		t.Fatalf("main output missing turns: %s", out.String())
	}

	var stderr bytes.Buffer
	demoStderr = &stderr
	var gotExit int
	demoExit = func(code int) { gotExit = code }
	exitf("failed: %s", "x")
	if gotExit != 1 || !strings.Contains(stderr.String(), "failed: x") {
		t.Fatalf("exitf got code=%d stderr=%q", gotExit, stderr.String())
	}
}

func readBody(t *testing.T, r *http.Request) string {
	t.Helper()
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("read request body: %v", err)
	}
	return string(raw)
}

func TestMainTurnFailuresExit(t *testing.T) {
	oldStdout, oldStderr, oldExit := demoStdout, demoStderr, demoExit
	t.Cleanup(func() {
		demoStdout, demoStderr, demoExit = oldStdout, oldStderr, oldExit
	})
	demoStdout = io.Discard
	demoStderr = io.Discard
	var gotExit int
	demoExit = func(code int) { gotExit = code }
	t.Setenv("DELTASIGNAL_AGENT_URL", "://bad")
	main()
	if gotExit != 1 {
		t.Fatalf("turn 1 failure exit = %d", gotExit)
	}

	gotExit = 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(readBody(t, r), `"tripcode"`) {
			_, _ = w.Write([]byte(`{"tripcode":"TF-SUB-9DA70A7F98","packet":{}}`))
			return
		}
		http.Error(w, "missing", http.StatusNotFound)
	}))
	defer server.Close()
	t.Setenv("DELTASIGNAL_AGENT_URL", server.URL)
	main()
	if gotExit != 1 {
		t.Fatalf("turn 2 failure exit = %d", gotExit)
	}
}
