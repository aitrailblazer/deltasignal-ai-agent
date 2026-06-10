package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aitrailblazer/deltasignal-ai-agent/internal/agent"
)

type failingToolClient struct{}

func (failingToolClient) StressSignals(context.Context, string) (agent.SpecialistResult, error) {
	return agent.SpecialistResult{}, errors.New("stress failed")
}

func (failingToolClient) CompanyEvidence(context.Context, string) (agent.SpecialistResult, error) {
	return agent.SpecialistResult{}, errors.New("company failed")
}

func (failingToolClient) PeerContext(context.Context, string) (agent.SpecialistResult, error) {
	return agent.SpecialistResult{}, errors.New("peer failed")
}

type fakeTripCodeResolver struct {
	err error
}

func (r fakeTripCodeResolver) ResolveTripCodeResearchPacket(_ context.Context, req agent.TripCodeResearchRequest) (agent.TripCodeResearchResponse, error) {
	if r.err != nil {
		return agent.TripCodeResearchResponse{}, r.err
	}
	return agent.TripCodeResearchResponse{
		TripCode:    strings.ToUpper(req.TripCode),
		Issuer:      strings.ToUpper(req.Issuer),
		GeneratedAt: time.Date(2026, 6, 7, 12, 0, 0, 0, time.UTC),
		Mode:        "live-mcp-tripcode",
		Packet: map[string]any{
			"article": map[string]any{"title": "Hut 8"},
			"river":   map[string]any{"nodes": []any{map[string]any{"tripcode": "TF-SUB-1"}}},
		},
		Disclosures: agent.DefaultTripCodeDisclosures(),
	}, nil
}

type fakeTripCodeSynthesizer struct {
	text string
	err  error
}

func (s fakeTripCodeSynthesizer) SynthesizeTripCode(context.Context, agent.TripCodeResearchRequest, agent.TripCodeResearchResponse) (string, error) {
	return s.text, s.err
}

type errReadCloser struct{}

func (errReadCloser) Read([]byte) (int, error) {
	return 0, errors.New("read failed")
}

func (errReadCloser) Close() error {
	return nil
}

func TestToolClientFromEnvDefaultsToDemo(t *testing.T) {
	t.Setenv("DELTASIGNAL_USE_LIVE_MCP", "")
	t.Setenv("MCP_API_KEY", "")
	t.Setenv("DELTASIGNAL_API_KEY", "")

	if _, ok := ToolClientFromEnv().(agent.DemoToolClient); !ok {
		t.Fatalf("ToolClientFromEnv returned %T, want DemoToolClient", ToolClientFromEnv())
	}
}

func TestToolClientFromEnvEnablesLiveMCP(t *testing.T) {
	t.Setenv("DELTASIGNAL_USE_LIVE_MCP", "true")
	t.Setenv("MCP_API_KEY", "test-key")
	t.Setenv("DELTASIGNAL_ATLAS_BASE_URL", "https://example.test")
	t.Setenv("DELTASIGNAL_MCP_STRESS_TOOL", "stress")
	t.Setenv("DELTASIGNAL_MCP_COMPANY_TOOL", "company")
	t.Setenv("DELTASIGNAL_MCP_PEER_TOOL", "peer")
	t.Setenv("DELTASIGNAL_MCP_TRIPCODE_TOOL", "tripcode")

	client, ok := ToolClientFromEnv().(agent.AtlasMCPToolClient)
	if !ok {
		t.Fatalf("ToolClientFromEnv returned %T, want AtlasMCPToolClient", ToolClientFromEnv())
	}
	if client.BaseURL != "https://example.test" {
		t.Fatalf("BaseURL = %q, want https://example.test", client.BaseURL)
	}
	if client.APIKey != "test-key" {
		t.Fatalf("APIKey not loaded from MCP_API_KEY")
	}
	if client.StressTool != "stress" || client.CompanyTool != "company" || client.PeerTool != "peer" {
		t.Fatalf("tool overrides not loaded: %#v", client)
	}
	if client.TripCodeTool != "tripcode" {
		t.Fatalf("TripCodeTool = %q, want tripcode", client.TripCodeTool)
	}
}

func TestToolClientFromEnvFallsBackWhenLiveMCPKeyMissing(t *testing.T) {
	t.Setenv("DELTASIGNAL_USE_LIVE_MCP", "true")
	t.Setenv("MCP_API_KEY", "")
	t.Setenv("DELTASIGNAL_API_KEY", "")

	if _, ok := ToolClientFromEnv().(agent.DemoToolClient); !ok {
		t.Fatalf("ToolClientFromEnv returned %T, want DemoToolClient", ToolClientFromEnv())
	}
}

func TestTripCodeResolverFromEnvRequiresKey(t *testing.T) {
	t.Setenv("MCP_API_KEY", "")
	t.Setenv("DELTASIGNAL_API_KEY", "")
	t.Setenv("DELTASIGNAL_ENABLE_TRIPCODE_FALLBACK", "")

	if TripCodeResolverFromEnv() != nil {
		t.Fatal("TripCodeResolverFromEnv returned resolver without API key")
	}
}

func TestTripCodeResolverFromEnvCanUseFallbackWithoutKey(t *testing.T) {
	t.Setenv("MCP_API_KEY", "")
	t.Setenv("DELTASIGNAL_API_KEY", "")
	t.Setenv("DELTASIGNAL_ENABLE_TRIPCODE_FALLBACK", "true")

	if _, ok := TripCodeResolverFromEnv().(agent.DemoTripCodeResolver); !ok {
		t.Fatalf("TripCodeResolverFromEnv returned %T, want DemoTripCodeResolver", TripCodeResolverFromEnv())
	}
}

func TestTripCodeResolverFromEnvUsesMCPClient(t *testing.T) {
	t.Setenv("MCP_API_KEY", "test-key")
	t.Setenv("DELTASIGNAL_ATLAS_BASE_URL", "https://example.test")
	t.Setenv("DELTASIGNAL_MCP_TRIPCODE_TOOL", "tripcode")
	t.Setenv("DELTASIGNAL_ENABLE_TRIPCODE_FALLBACK", "")

	resolver, ok := TripCodeResolverFromEnv().(agent.AtlasMCPToolClient)
	if !ok {
		t.Fatalf("TripCodeResolverFromEnv returned %T, want AtlasMCPToolClient", TripCodeResolverFromEnv())
	}
	if resolver.BaseURL != "https://example.test" || resolver.APIKey != "test-key" || resolver.TripCodeTool != "tripcode" {
		t.Fatalf("resolver env not loaded: %#v", resolver)
	}
}

func TestTripCodeResolverFromEnvWrapsMCPWithFallback(t *testing.T) {
	t.Setenv("MCP_API_KEY", "test-key")
	t.Setenv("DELTASIGNAL_ENABLE_TRIPCODE_FALLBACK", "true")

	resolver, ok := TripCodeResolverFromEnv().(agent.FallbackTripCodeResolver)
	if !ok {
		t.Fatalf("TripCodeResolverFromEnv returned %T, want FallbackTripCodeResolver", TripCodeResolverFromEnv())
	}
	if resolver.Primary == nil || resolver.Fallback == nil {
		t.Fatalf("fallback resolver not fully configured: %#v", resolver)
	}
}

func TestCostTrackerFromEnv(t *testing.T) {
	t.Setenv("DELTASIGNAL_COST_TRACKING", "true")
	t.Setenv("DELTASIGNAL_GOOGLE_CREDIT_BUDGET_USD", "500")
	t.Setenv("DELTASIGNAL_ESTIMATED_BRIEF_COST_USD", "0.02")
	t.Setenv("DELTASIGNAL_ESTIMATED_TRIPCODE_COST_USD", "0.03")
	t.Setenv("DELTASIGNAL_ESTIMATED_SESSION_MEMORY_COST_USD", "0.004")
	t.Setenv("DELTASIGNAL_COST_SOURCE", "test-estimate")
	t.Setenv("DELTASIGNAL_OFFICIAL_BILLING_AVAILABLE", "true")
	t.Setenv("DELTASIGNAL_OFFICIAL_BILLING_SOURCE", "manual-console")
	t.Setenv("DELTASIGNAL_BILLING_PROJECT_ID", "startup-ai-deltasignal")
	t.Setenv("DELTASIGNAL_BILLING_ACCOUNT_NAME", "billingAccounts/018530-615AB8-696753")
	t.Setenv("DELTASIGNAL_BILLING_ENABLED", "true")
	t.Setenv("DELTASIGNAL_OFFICIAL_BILLING_SPENT_USD", "12.34")
	t.Setenv("DELTASIGNAL_OFFICIAL_BILLING_CREDIT_BUDGET_USD", "500")
	t.Setenv("DELTASIGNAL_OFFICIAL_BILLING_UPDATED_AT", "2026-06-08T11:00:00Z")

	tracker := CostTrackerFromEnv()
	snapshot := tracker.Record("brief")
	if snapshot == nil || snapshot.Source != "test-estimate" || snapshot.RequestCostUSD != 0.02 || snapshot.RemainingUSD != 499.98 {
		t.Fatalf("unexpected cost snapshot: %#v", snapshot)
	} else if snapshot.OfficialBilling == nil || !snapshot.OfficialBilling.Available || snapshot.OfficialBilling.Source != "manual-console" || snapshot.OfficialBilling.BillingAccountName != "billingAccounts/018530-615AB8-696753" || !snapshot.OfficialBilling.BillingEnabled || snapshot.OfficialBilling.RemainingUSD < 487.65 || snapshot.OfficialBilling.RemainingUSD > 487.67 {
		t.Fatalf("unexpected official billing snapshot: %#v", snapshot.OfficialBilling)
	}
}

func TestEnvFloatFallbacks(t *testing.T) {
	t.Setenv("BAD_FLOAT", "nope")
	if got := envFloat("MISSING_FLOAT", 1.5); got != 1.5 {
		t.Fatalf("missing float = %f", got)
	}
	if got := envFloat("BAD_FLOAT", 1.5); got != 1.5 {
		t.Fatalf("bad float = %f", got)
	}
	t.Setenv("GOOD_FLOAT", "2.75")
	if got := envFloat("GOOD_FLOAT", 1.5); got != 2.75 {
		t.Fatalf("good float = %f", got)
	}
	if got := envBool("MISSING_BOOL", true); !got {
		t.Fatal("missing bool fallback not used")
	}
	t.Setenv("BAD_BOOL", "nope")
	if got := envBool("BAD_BOOL", true); !got {
		t.Fatal("bad bool fallback not used")
	}
	t.Setenv("GOOD_BOOL", "false")
	if got := envBool("GOOD_BOOL", true); got {
		t.Fatal("good bool not parsed")
	}
}

func TestAuthorizedDemoRequestAllowsWhenUnset(t *testing.T) {
	t.Setenv("DELTASIGNAL_DEMO_API_KEY", "")
	req := httptest.NewRequest(http.MethodPost, "/v1/brief", nil)
	if !authorizedDemoRequest(req) {
		t.Fatal("authorizedDemoRequest denied request when demo key was unset")
	}
}

func TestAuthorizedDemoRequestRequiresDemoKeyWhenSet(t *testing.T) {
	t.Setenv("DELTASIGNAL_DEMO_API_KEY", "judge-key")

	req := httptest.NewRequest(http.MethodPost, "/v1/brief", nil)
	if authorizedDemoRequest(req) {
		t.Fatal("authorizedDemoRequest allowed request without key")
	}

	req.Header.Set("X-Demo-Key", "judge-key")
	if !authorizedDemoRequest(req) {
		t.Fatal("authorizedDemoRequest denied matching X-Demo-Key")
	}
}

func TestRunReturnsZeroWhenServerStopsCleanly(t *testing.T) {
	oldListen := listenAndServe
	oldOutput := outputWriter
	t.Cleanup(func() {
		listenAndServe = oldListen
		outputWriter = oldOutput
	})
	outputWriter = io.Discard
	listenAndServe = func(addr string, handler http.Handler) error {
		if addr != ":8080" {
			t.Fatalf("addr = %q, want :8080", addr)
		}
		if handler == nil {
			t.Fatal("handler is nil")
		}
		return nil
	}
	t.Setenv("PORT", "")
	if code := run(); code != 0 {
		t.Fatalf("run code = %d, want 0", code)
	}
}

func TestRunEnablesGeminiBranch(t *testing.T) {
	oldListen := listenAndServe
	oldOutput := outputWriter
	t.Cleanup(func() {
		listenAndServe = oldListen
		outputWriter = oldOutput
	})
	outputWriter = io.Discard
	listenAndServe = func(string, http.Handler) error { return nil }
	t.Setenv("DELTASIGNAL_USE_GEMINI", "true")
	t.Setenv("GEMINI_MODEL", "gemini-test")
	if code := run(); code != 0 {
		t.Fatalf("run code = %d, want 0", code)
	}
}

func TestRunReturnsOneWhenServerFails(t *testing.T) {
	oldListen := listenAndServe
	oldOutput := outputWriter
	t.Cleanup(func() {
		listenAndServe = oldListen
		outputWriter = oldOutput
	})
	outputWriter = io.Discard
	listenAndServe = func(addr string, handler http.Handler) error {
		if addr != ":9999" {
			t.Fatalf("addr = %q, want :9999", addr)
		}
		return errors.New("bind failed")
	}
	t.Setenv("PORT", "9999")
	if code := run(); code != 1 {
		t.Fatalf("run code = %d, want 1", code)
	}
}

func TestMainCallsExitProcess(t *testing.T) {
	oldListen := listenAndServe
	oldExit := exitProcess
	oldOutput := outputWriter
	t.Cleanup(func() {
		listenAndServe = oldListen
		exitProcess = oldExit
		outputWriter = oldOutput
	})
	outputWriter = io.Discard
	listenAndServe = func(string, http.Handler) error { return nil }
	var gotCode int
	exitProcess = func(code int) { gotCode = code }
	main()
	if gotCode != 0 {
		t.Fatalf("exit code = %d, want 0", gotCode)
	}
}

func TestMuxHealthAndBriefRoutes(t *testing.T) {
	mux := newMux(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		agent.Coordinator{Tools: agent.DemoToolClient{}},
		nil,
		agent.NewTripCodeMemoryStore(1),
		nil,
		agent.NewCostTracker(agent.CostTrackerConfig{Enabled: true, BudgetUSD: 1, BriefCostUSD: 0.05}),
		nil,
	)

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), "deltasignal-ai-agent") {
		t.Fatalf("health response = %d %s", rr.Code, rr.Body.String())
	}

	t.Setenv("DELTASIGNAL_DEMO_API_KEY", "key")
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/v1/brief", strings.NewReader(`{}`)))
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("brief unauthorized code = %d", rr.Code)
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/brief", strings.NewReader(`{bad`))
	req.Header.Set("X-Demo-Key", "key")
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("brief bad JSON code = %d", rr.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/v1/brief", strings.NewReader(`{"issuer":"hut"}`))
	req.Header.Set("Authorization", "Bearer key")
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), `"issuer":"HUT"`) || !strings.Contains(rr.Body.String(), `"request_cost_usd":0.05`) {
		t.Fatalf("brief success response = %d %s", rr.Code, rr.Body.String())
	}
}

func TestMuxBriefToolFailure(t *testing.T) {
	mux := newMux(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		agent.Coordinator{Tools: failingToolClient{}},
		nil,
		agent.NewTripCodeMemoryStore(1),
		nil,
		nil,
		nil,
	)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/v1/brief", strings.NewReader(`{"issuer":"HUT"}`)))
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("brief failure code = %d", rr.Code)
	}
}

func TestMuxProductLoopRoute(t *testing.T) {
	mux := newMux(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		agent.Coordinator{Tools: agent.DemoToolClient{}},
		nil,
		agent.NewTripCodeMemoryStore(1),
		nil,
		agent.NewCostTracker(agent.CostTrackerConfig{Enabled: true, BudgetUSD: 1}),
		nil,
	)
	t.Setenv("DELTASIGNAL_DEMO_API_KEY", "key")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/v1/product-loop", strings.NewReader(`{}`)))
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("product-loop unauthorized code = %d", rr.Code)
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/product-loop", strings.NewReader(`{bad`))
	req.Header.Set("X-Demo-Key", "key")
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("product-loop bad JSON code = %d", rr.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/v1/product-loop", strings.NewReader(`{"objective":"Ship ADK loop","workflow_type":"TripCode Monitor","risk_level":"high","allow_parallel_builders":true}`))
	req.Header.Set("X-Demo-Key", "key")
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	body := rr.Body.String()
	if rr.Code != http.StatusOK || !strings.Contains(body, `"objective":"Ship ADK loop"`) || !strings.Contains(body, `"workflow_type":"tripcode_monitor"`) || !strings.Contains(body, `"human_approval_required":true`) || !strings.Contains(body, `"parallel-builders-with-one-worktree-per-task"`) || !strings.Contains(body, `"request_kind":"product-loop"`) {
		t.Fatalf("product-loop success response = %d %s", rr.Code, body)
	}
}

func TestMuxTripCodeRoutes(t *testing.T) {
	memory := agent.NewTripCodeMemoryStore(2)
	mux := newMux(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		agent.Coordinator{Tools: agent.DemoToolClient{}},
		fakeTripCodeResolver{},
		memory,
		fakeTripCodeSynthesizer{text: "gemini summary"},
		agent.NewCostTracker(agent.CostTrackerConfig{
			Enabled:              true,
			BudgetUSD:            1,
			TripCodeCostUSD:      0.10,
			SessionMemoryCostUSD: 0.01,
		}),
		nil,
	)
	t.Setenv("DELTASIGNAL_DEMO_API_KEY", "key")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/v1/tripcode", strings.NewReader(`{}`)))
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("tripcode unauthorized code = %d", rr.Code)
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/tripcode", strings.NewReader(`{bad`))
	req.Header.Set("X-Demo-Key", "key")
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("tripcode bad JSON code = %d", rr.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/v1/tripcode", strings.NewReader(`{}`))
	req.Header.Set("X-Demo-Key", "key")
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("tripcode missing input code = %d", rr.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/v1/tripcode", strings.NewReader(`{"session_id":"missing"}`))
	req.Header.Set("X-Demo-Key", "key")
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("tripcode missing memory code = %d", rr.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/v1/tripcode", strings.NewReader(`{"session_id":"hut","tripcode":"tf-sub-9da70a7f98","issuer":"hut"}`))
	req.Header.Set("X-Demo-Key", "key")
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), `"gemini_summary":"gemini summary"`) || !strings.Contains(rr.Body.String(), `"mode":"live-mcp-tripcode-gemini"`) || !strings.Contains(rr.Body.String(), `"remaining_usd":0.9`) {
		t.Fatalf("tripcode success response = %d %s", rr.Code, rr.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/v1/tripcode", strings.NewReader(`{"session_id":"hut","question":"what changed?"}`))
	req.Header.Set("X-Demo-Key", "key")
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), `"mode":"session-memory"`) || !strings.Contains(rr.Body.String(), `"tracked_spent_usd":0.11`) {
		t.Fatalf("tripcode memory response = %d %s", rr.Code, rr.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/v1/usage", nil)
	req.Header.Set("X-Demo-Key", "key")
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), `"remaining_usd":0.89`) {
		t.Fatalf("usage response = %d %s", rr.Code, rr.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/v1/usage", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("usage unauthorized code = %d", rr.Code)
	}
}

func TestMuxResolveJudgeRoutes(t *testing.T) {
	oldFetchContext := agent.FetchDefaultAgentContext
	t.Cleanup(func() { agent.FetchDefaultAgentContext = oldFetchContext })
	agent.FetchDefaultAgentContext = func(context.Context) (agent.AgentContextSnapshot, error) {
		return agent.AgentContextSnapshot{
			Enabled: true,
			Purpose: "test context",
			Sources: []agent.AgentContextSource{{
				URL:    "https://aitrailblazer.github.io/deltasignal-atlas-codex-plugin/CLAUDE.md",
				Status: "fetched",
				Bytes:  123,
				SHA256: "abc123",
			}},
		}, nil
	}
	memory := agent.NewTripCodeMemoryStore(2)
	mux := newMux(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		agent.Coordinator{Tools: agent.DemoToolClient{}},
		fakeTripCodeResolver{},
		memory,
		fakeTripCodeSynthesizer{text: "judge gemini summary"},
		agent.NewCostTracker(agent.CostTrackerConfig{
			Enabled:              true,
			BudgetUSD:            1,
			TripCodeCostUSD:      0.10,
			SessionMemoryCostUSD: 0.01,
		}),
		nil,
	)
	t.Setenv("DELTASIGNAL_DEMO_API_KEY", "judge-key")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/resolve?tripcode=TF-SUB-9DA70A7F98", nil))
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("resolve unauthorized code = %d", rr.Code)
	}

	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/resolve", nil))
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("resolve post unauthorized code = %d", rr.Code)
	}

	req := httptest.NewRequest(http.MethodGet, "/resolve?tripcode=tf-sub-9da70a7f98&session_id=demo2&issuer=hut&payload_mode=compact&include_article_body=true&include_filing_evidence=true&include_prior_articles=true&include_thesis_map=true&include_agent_context=true", nil)
	req.Header.Set("Authorization", "Bearer judge-key")
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), `"tripcode":"TF-SUB-9DA70A7F98"`) || !strings.Contains(rr.Body.String(), `"issuer":"HUT"`) || !strings.Contains(rr.Body.String(), `"gemini_summary":"judge gemini summary"`) || !strings.Contains(rr.Body.String(), `"agent_context"`) || !strings.Contains(rr.Body.String(), `"execution_trace"`) {
		t.Fatalf("resolve get response = %d %s", rr.Code, rr.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/resolve?session_id=demo2", strings.NewReader("Using the previous HUT River, what should I monitor next?"))
	req.Header.Set("X-Demo-Key", "judge-key")
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), `"mode":"session-memory"`) || !strings.Contains(rr.Body.String(), `"question":"Using the previous HUT River, what should I monitor next?"`) {
		t.Fatalf("resolve post text response = %d %s", rr.Code, rr.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/resolve?session_id=demo2", strings.NewReader(`{"question":"Which assumptions weakened?"}`))
	req.Header.Set("X-Demo-Key", "judge-key")
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), `"question":"Which assumptions weakened?"`) || !strings.Contains(rr.Body.String(), `"tracked_spent_usd":0.12`) {
		t.Fatalf("resolve post json response = %d %s", rr.Code, rr.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/resolve", strings.NewReader(`{bad`))
	req.Header.Set("X-Demo-Key", "judge-key")
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("resolve bad json code = %d", rr.Code)
	}
}

func TestMuxTripCodeResolverMissingAndErrors(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	coordinator := agent.Coordinator{Tools: agent.DemoToolClient{}}

	mux := newMux(logger, coordinator, nil, agent.NewTripCodeMemoryStore(1), nil, nil, nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/v1/tripcode", strings.NewReader(`{"tripcode":"TF-SUB-X"}`)))
	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("missing resolver code = %d", rr.Code)
	}

	mux = newMux(logger, coordinator, fakeTripCodeResolver{err: errors.New("resolver failed")}, agent.NewTripCodeMemoryStore(1), nil, nil, nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/v1/tripcode", strings.NewReader(`{"tripcode":"TF-SUB-X"}`)))
	if rr.Code != http.StatusBadGateway {
		t.Fatalf("resolver error code = %d", rr.Code)
	}

	mux = newMux(logger, coordinator, fakeTripCodeResolver{}, agent.NewTripCodeMemoryStore(1), fakeTripCodeSynthesizer{err: errors.New("gemini failed")}, nil, nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/v1/tripcode", strings.NewReader(`{"tripcode":"TF-SUB-X"}`)))
	if rr.Code != http.StatusOK || strings.Contains(rr.Body.String(), "gemini_summary") {
		t.Fatalf("synth error response = %d %s", rr.Code, rr.Body.String())
	}

	oldFetchContext := agent.FetchDefaultAgentContext
	t.Cleanup(func() { agent.FetchDefaultAgentContext = oldFetchContext })
	agent.FetchDefaultAgentContext = func(context.Context) (agent.AgentContextSnapshot, error) {
		return agent.AgentContextSnapshot{}, errors.New("context failed")
	}
	mux = newMux(logger, coordinator, fakeTripCodeResolver{}, agent.NewTripCodeMemoryStore(1), nil, nil, nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/v1/tripcode", strings.NewReader(`{"tripcode":"TF-SUB-X","include_agent_context":true}`)))
	if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), `"agent_context"`) || !strings.Contains(rr.Body.String(), "context failed") {
		t.Fatalf("context error response = %d %s", rr.Code, rr.Body.String())
	}
}

func TestMuxRateLimitsProtectedRoutes(t *testing.T) {
	t.Setenv("DELTASIGNAL_DEMO_API_KEY", "key")
	newLimitedMux := func() *http.ServeMux {
		return newMux(
			slog.New(slog.NewTextHandler(io.Discard, nil)),
			agent.Coordinator{Tools: agent.DemoToolClient{}},
			fakeTripCodeResolver{},
			agent.NewTripCodeMemoryStore(1),
			nil,
			agent.NewCostTracker(agent.CostTrackerConfig{Enabled: true, BudgetUSD: 1, BriefCostUSD: 0.01, TripCodeCostUSD: 0.01}),
			NewRateLimiter(true, 1, time.Minute),
		)
	}
	routes := []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodPost, "/v1/brief", `{"issuer":"HUT"}`},
		{http.MethodPost, "/v1/tripcode", `{"tripcode":"TF-SUB-X"}`},
		{http.MethodGet, "/resolve?tripcode=TF-SUB-X", ""},
		{http.MethodPost, "/resolve", `{"tripcode":"TF-SUB-X"}`},
		{http.MethodGet, "/v1/usage", ""},
		{http.MethodPost, "/v1/product-loop", `{"objective":"test"}`},
	}
	for _, route := range routes {
		mux := newLimitedMux()
		req := httptest.NewRequest(route.method, route.path, strings.NewReader(route.body))
		req.Header.Set("X-Demo-Key", "key")
		if strings.HasPrefix(strings.TrimSpace(route.body), "{") {
			req.Header.Set("Content-Type", "application/json")
		}
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
		if rr.Code < 200 || rr.Code >= 300 {
			t.Fatalf("%s %s first request code = %d body=%s", route.method, route.path, rr.Code, rr.Body.String())
		}

		req = httptest.NewRequest(route.method, route.path, strings.NewReader(route.body))
		req.Header.Set("X-Demo-Key", "key")
		if strings.HasPrefix(strings.TrimSpace(route.body), "{") {
			req.Header.Set("Content-Type", "application/json")
		}
		rr = httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
		if rr.Code != http.StatusTooManyRequests || !strings.Contains(rr.Body.String(), "rate limit exceeded") || rr.Header().Get("Retry-After") == "" {
			t.Fatalf("%s %s limited response = %d headers=%#v body=%s", route.method, route.path, rr.Code, rr.Header(), rr.Body.String())
		}
	}
}

func TestTripCodeMemoryFromEnv(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sessions.json")
	t.Setenv("DELTASIGNAL_MEMORY_MAX_ENTRIES", "3")
	t.Setenv("DELTASIGNAL_MEMORY_FILE", path)
	store := TripCodeMemoryFromEnv()
	if status := store.Status("session"); status.Backend != "file" || !status.Durable || status.EntryLimit != 3 {
		t.Fatalf("memory from env status = %#v", status)
	}

	blocker := filepath.Join(t.TempDir(), "blocker")
	if err := os.WriteFile(blocker, []byte("not a directory"), 0o600); err != nil {
		t.Fatalf("write blocker: %v", err)
	}
	t.Setenv("DELTASIGNAL_MEMORY_FILE", filepath.Join(blocker, "sessions.json"))
	store = TripCodeMemoryFromEnv()
	if status := store.Status("session"); status.Backend != "file" || status.LastError == "" {
		t.Fatalf("memory from env missing file status = %#v", status)
	}
}

func TestTripCodeRequestHelpers(t *testing.T) {
	values := make(url.Values)
	values.Set("tripcode", "TF-SUB-1")
	values.Set("issuer", "HUT")
	values.Set("session_id", "demo")
	values.Set("question", "what changed")
	values.Set("payload_mode", "full")
	values.Set("include_article_body", "true")
	values.Set("include_filing_evidence", "not-bool")
	values.Set("include_agent_context", "true")
	req := tripCodeRequestFromQuery(values)
	if req.TripCode != "TF-SUB-1" || req.Issuer != "HUT" || req.SessionID != "demo" || req.Question != "what changed" || req.PayloadMode != "full" {
		t.Fatalf("query request fields not loaded: %#v", req)
	}
	if !req.IncludeArticleBody || req.IncludeFilingEvidence || !req.IncludeAgentContext {
		t.Fatalf("query bools not parsed as expected: %#v", req)
	}

	merged := mergeTripCodeRequest(
		agent.TripCodeResearchRequest{TripCode: "base", IncludePriorArticles: true},
		agent.TripCodeResearchRequest{TripCode: "override", Issuer: "HUT", SessionID: "s", Question: "q", PayloadMode: "compact", IncludeThesisMap: true, IncludeAgentContext: true},
	)
	if merged.TripCode != "override" || merged.Issuer != "HUT" || merged.SessionID != "s" || merged.Question != "q" || merged.PayloadMode != "compact" || !merged.IncludePriorArticles || !merged.IncludeThesisMap || !merged.IncludeAgentContext {
		t.Fatalf("merged request unexpected: %#v", merged)
	}

	req, err := tripCodeRequestFromResolvePost(httptest.NewRequest(http.MethodPost, "/resolve?session_id=query-session", nil))
	if err != nil || req.SessionID != "query-session" {
		t.Fatalf("empty resolve post request = %#v err=%v", req, err)
	}

	errReq := httptest.NewRequest(http.MethodPost, "/resolve", nil)
	errReq.Body = errReadCloser{}
	if _, err := tripCodeRequestFromResolvePost(errReq); err == nil {
		t.Fatal("tripCodeRequestFromResolvePost returned nil error for failing body")
	}

	if got := packetTraceEvidence(nil); got != "packet_keys=0" {
		t.Fatalf("packetTraceEvidence(nil) = %q", got)
	}
	if got := agentContextTraceEvidence(agent.AgentContextSnapshot{Sources: []agent.AgentContextSource{{Status: "unavailable"}}}); !strings.Contains(got, "unknown status=unavailable") {
		t.Fatalf("agentContextTraceEvidence empty URL = %q", got)
	}
	fallbackPacket := map[string]any{"live_mcp_error": "unsupported"}
	if actor := resolverTraceActor(fallbackPacket); actor != "google-agent" {
		t.Fatalf("resolverTraceActor fallback = %q", actor)
	}
	if action := resolverTraceAction(fallbackPacket); !strings.Contains(action, "fallback") {
		t.Fatalf("resolverTraceAction fallback = %q", action)
	}
	if mode := tripcodeGeminiMode("demo-tripcode-river", fallbackPacket); mode != "demo-tripcode-river-gemini" {
		t.Fatalf("tripcodeGeminiMode fallback = %q", mode)
	}
	if mode := tripcodeGeminiMode("", fallbackPacket); mode != "demo-tripcode-river-gemini" {
		t.Fatalf("tripcodeGeminiMode empty fallback = %q", mode)
	}
	if actor := resolverTraceActor(map[string]any{}); actor != "mcp" {
		t.Fatalf("resolverTraceActor live = %q", actor)
	}
	if action := resolverTraceAction(map[string]any{}); !strings.Contains(action, "bounded research packet") {
		t.Fatalf("resolverTraceAction live = %q", action)
	}
	if mode := tripcodeGeminiMode("live-mcp-tripcode", map[string]any{}); mode != "live-mcp-tripcode-gemini" {
		t.Fatalf("tripcodeGeminiMode live = %q", mode)
	}
}

func TestWriteJSON(t *testing.T) {
	rr := httptest.NewRecorder()
	writeJSON(rr, http.StatusAccepted, map[string]string{"ok": "true"})
	if rr.Code != http.StatusAccepted || rr.Header().Get("Content-Type") != "application/json" || !bytes.Contains(rr.Body.Bytes(), []byte(`"ok":"true"`)) {
		t.Fatalf("writeJSON response = %d %q %s", rr.Code, rr.Header().Get("Content-Type"), rr.Body.String())
	}
}
