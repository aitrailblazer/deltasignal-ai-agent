package main

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aitrailblazer/deltasignal-ai-agent/internal/agent"
)

var errTestResolver = errors.New("resolver failed")

func TestA2AAgentCardRoutes(t *testing.T) {
	t.Setenv("DELTASIGNAL_PUBLIC_BASE_URL", "https://demo.example")
	t.Setenv("DELTASIGNAL_A2A_PROTOCOL_VERSION", "0.3.0")
	t.Setenv("DELTASIGNAL_AGENT_VERSION", "2.0.0")
	mux := newMux(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		agent.Coordinator{Tools: agent.DemoToolClient{}},
		fakeTripCodeResolver{},
		agent.NewTripCodeMemoryStore(1),
		nil,
		nil,
		nil,
	)
	for _, path := range []string{"/.well-known/agent-card.json", "/a2a/app/.well-known/agent-card.json"} {
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, path, nil))
		if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), `"protocolVersion":"0.3.0"`) || !strings.Contains(rr.Body.String(), `"url":"https://demo.example/a2a"`) || !strings.Contains(rr.Body.String(), `"resolve-tripcode-research-packet"`) {
			t.Fatalf("agent card %s = %d %s", path, rr.Code, rr.Body.String())
		}
	}
}

func TestBuildA2AAgentCardFallbackBaseURL(t *testing.T) {
	t.Setenv("DELTASIGNAL_PUBLIC_BASE_URL", "")
	req := httptest.NewRequest(http.MethodGet, "http://internal/.well-known/agent-card.json", nil)
	req.Host = "agent.example"
	req.Header.Set("X-Forwarded-Proto", "https")
	card := buildA2AAgentCard(req)
	if card.URL != "https://agent.example/a2a" || card.Name == "" || len(card.Skills) != 1 || card.Capabilities["streaming"] {
		t.Fatalf("card = %#v", card)
	}
	req.Host = ""
	req.Header.Set("X-Forwarded-Proto", "")
	if got := publicBaseURL(req); got != "http://localhost" {
		t.Fatalf("fallback base URL = %q", got)
	}
}

func TestA2AMessageRouteSuccessAndSessionMemory(t *testing.T) {
	memory := agent.NewTripCodeMemoryStore(2)
	mux := newMux(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		agent.Coordinator{Tools: agent.DemoToolClient{}},
		fakeTripCodeResolver{},
		memory,
		fakeTripCodeSynthesizer{text: "a2a gemini"},
		agent.NewCostTracker(agent.CostTrackerConfig{Enabled: true, BudgetUSD: 1, TripCodeCostUSD: 0.05}),
		nil,
	)
	body := `{"jsonrpc":"2.0","id":"a2a-1","method":"message/send","params":{"message":{"parts":[{"kind":"text","text":"Resolve tf-sub-9da70a7f98 for the HUT River"}]},"metadata":{"session_id":"a2a-session"}}}`
	req := httptest.NewRequest(http.MethodPost, "/a2a", strings.NewReader(body))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), `"state":"completed"`) || !strings.Contains(rr.Body.String(), `"tripcode":"TF-SUB-9DA70A7F98"`) || !strings.Contains(rr.Body.String(), `"a2a gemini"`) {
		t.Fatalf("a2a success = %d %s", rr.Code, rr.Body.String())
	}
	if snap := memory.Snapshot("a2a-session"); !snap.Available || snap.Turns != 1 {
		t.Fatalf("memory snapshot = %#v", snap)
	}
}

func TestA2AMessageRouteErrorsAndRateLimit(t *testing.T) {
	t.Setenv("DELTASIGNAL_DEMO_API_KEY", "key")
	limited := newMux(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		agent.Coordinator{Tools: agent.DemoToolClient{}},
		fakeTripCodeResolver{},
		agent.NewTripCodeMemoryStore(1),
		nil,
		nil,
		NewRateLimiter(true, 1, 0),
	)

	req := httptest.NewRequest(http.MethodPost, "/a2a", strings.NewReader(`{}`))
	rr := httptest.NewRecorder()
	limited.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized = %d %s", rr.Code, rr.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/a2a", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"message/send","params":{"text":"Resolve TF-SUB-9DA70A7F98","session_id":"s"}}`))
	req.Header.Set("X-Demo-Key", "key")
	rr = httptest.NewRecorder()
	limited.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("first limited request = %d %s", rr.Code, rr.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/a2a", strings.NewReader(`{"jsonrpc":"2.0","id":2,"method":"message/send","params":{"text":"Resolve TF-SUB-9DA70A7F98"}}`))
	req.Header.Set("X-Demo-Key", "key")
	rr = httptest.NewRecorder()
	limited.ServeHTTP(rr, req)
	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("rate limit = %d %s", rr.Code, rr.Body.String())
	}

	t.Setenv("DELTASIGNAL_DEMO_API_KEY", "")
	open := newMux(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		agent.Coordinator{Tools: agent.DemoToolClient{}},
		fakeTripCodeResolver{err: errTestResolver},
		agent.NewTripCodeMemoryStore(1),
		nil,
		nil,
		nil,
	)
	for _, tc := range []struct {
		body string
		want int
	}{
		{`{bad`, http.StatusBadRequest},
		{`{"jsonrpc":"2.0","id":"x","method":"unknown"}`, http.StatusBadRequest},
		{`{"jsonrpc":"2.0","id":"x","method":"message/send","params":{"text":"No code here"}}`, http.StatusOK},
		{`{"jsonrpc":"2.0","id":"x","method":"resolve_tripcode","params":{"text":"Resolve TF-SUB-9DA70A7F98"}}`, http.StatusBadGateway},
	} {
		rr = httptest.NewRecorder()
		open.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/a2a", strings.NewReader(tc.body)))
		if rr.Code != tc.want {
			t.Fatalf("body %s code = %d want %d response=%s", tc.body, rr.Code, tc.want, rr.Body.String())
		}
	}
}

func TestA2AHelpers(t *testing.T) {
	resp, status, mode := handleA2AMessage(
		t.Context(),
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		a2aRPCRequest{ID: 7, Method: "tasks/send", Params: json.RawMessage(`{"text":"Resolve TF-SUB-ABC123","session_id":"direct"}`)},
		fakeTripCodeResolver{},
		agent.NewTripCodeMemoryStore(1),
		nil,
		nil,
		beginRequest(httptest.NewRequest(http.MethodPost, "/a2a", nil), "/a2a"),
		nil,
	)
	if status != http.StatusOK || mode != "a2a-tripcode" || resp.JSONRPC != "2.0" {
		t.Fatalf("direct handle = status %d mode %s response %#v", status, mode, resp)
	}
	text, session := a2aTextAndSession(json.RawMessage(`{"message":{"parts":[{"kind":"text","text":"First"},{"type":"text","text":"Second"}]},"metadata":{"session_id":"s1"}}`))
	if text != "First\nSecond" || session != "s1" {
		t.Fatalf("message params = %q %q", text, session)
	}
	text, session = a2aTextAndSession(json.RawMessage(`{"text":"Resolve TF-SUB-ABC123","session_id":"s2"}`))
	if text != "Resolve TF-SUB-ABC123" || session != "s2" {
		t.Fatalf("fallback params = %q %q", text, session)
	}
	if got := stringMetadata(map[string]any{"session_id": 7}, "session_id"); got != "" {
		t.Fatalf("non-string metadata = %q", got)
	}
	if got := stringMetadata(nil, "session_id"); got != "" {
		t.Fatalf("nil metadata = %q", got)
	}
	if got := firstTripCode("resolve tf-sub-abc123 now"); got != "TF-SUB-ABC123" {
		t.Fatalf("first tripcode = %q", got)
	}
	if got := firstTripCode("no code"); got != "" {
		t.Fatalf("missing tripcode = %q", got)
	}
	if taskID(nil) == taskID("missing-id") {
		t.Fatal("nil task id should hash a sentinel JSON value")
	}
	if !strings.HasPrefix(taskID(make(chan int)), "a2a-") {
		t.Fatal("unmarshalable task id did not produce fallback")
	}
	if !strings.Contains(a2aRegisterCurl("https://agent.example/card.json"), "--registration-type a2a") {
		t.Fatal("register curl missing a2a flag")
	}
	if got := a2aMarketplaceCardPath(""); !strings.Contains(got, "<marketplace-agent-card-bucket>") {
		t.Fatalf("empty bucket card path = %q", got)
	}
	if got := a2aMarketplaceCardPath("gs://bucket"); got != "gs://bucket/deltasignal-gemini-ai-agent/agent-card.json" {
		t.Fatalf("gs bucket card path = %q", got)
	}
	if got := a2aMarketplaceCardPath("plain bucket"); got != "gs://plain%20bucket/deltasignal-gemini-ai-agent/agent-card.json" {
		t.Fatalf("plain bucket card path = %q", got)
	}
	if metadata := a2aMetadata(); metadata["non_advice"] != true || metadata["track"] != "Track 3 Path" {
		t.Fatalf("metadata = %#v", metadata)
	}
}
