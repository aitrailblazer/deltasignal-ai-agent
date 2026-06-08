package main

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aitrailblazer/deltasignal-ai-agent/internal/agent"
)

var (
	listenAndServe           = http.ListenAndServe
	exitProcess              = os.Exit
	outputWriter   io.Writer = os.Stdout
)

func main() {
	exitProcess(run())
}

func run() int {
	logger := slog.New(slog.NewJSONHandler(outputWriter, nil))
	coordinator := agent.Coordinator{
		Tools: ToolClientFromEnv(),
	}
	tripcodeResolver := TripCodeResolverFromEnv()
	tripcodeMemory := agent.NewTripCodeMemoryStore(20)
	costTracker := CostTrackerFromEnv()
	var tripcodeSynthesizer agent.TripCodeSynthesizer
	if strings.EqualFold(os.Getenv("DELTASIGNAL_USE_GEMINI"), "true") {
		gemini := agent.GeminiSynthesizer{Model: os.Getenv("GEMINI_MODEL")}
		coordinator.Synthesizer = gemini
		tripcodeSynthesizer = gemini
	}

	mux := newMux(logger, coordinator, tripcodeResolver, tripcodeMemory, tripcodeSynthesizer, costTracker)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	logger.Info("starting DeltaSignal Gemini AI Agent", "port", port)
	if err := listenAndServe(":"+port, mux); err != nil {
		logger.Error("server stopped", "error", err)
		return 1
	}
	return 0
}

func newMux(
	logger *slog.Logger,
	coordinator agent.Coordinator,
	tripcodeResolver agent.TripCodeResolver,
	tripcodeMemory *agent.TripCodeMemoryStore,
	tripcodeSynthesizer agent.TripCodeSynthesizer,
	costTracker *agent.CostTracker,
) *http.ServeMux {
	mux := http.NewServeMux()
	healthHandler := func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"ok":      true,
			"service": "deltasignal-ai-agent",
			"time":    time.Now().UTC(),
		})
	}
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("GET /healthz", healthHandler)
	mux.HandleFunc("POST /v1/brief", func(w http.ResponseWriter, r *http.Request) {
		if !authorizedDemoRequest(r) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing or invalid demo key"})
			return
		}
		var req agent.BriefRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON request"})
			return
		}
		resp, err := coordinator.BuildBrief(r.Context(), req)
		if err != nil {
			logger.Error("build brief failed", "error", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to build brief"})
			return
		}
		resp.Cost = costTracker.Record("brief")
		writeJSON(w, http.StatusOK, resp)
	})
	mux.HandleFunc("POST /v1/tripcode", func(w http.ResponseWriter, r *http.Request) {
		if !authorizedDemoRequest(r) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing or invalid demo key"})
			return
		}
		var req agent.TripCodeResearchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON request"})
			return
		}
		writeTripCodeHTTPResult(w, resolveTripCodeHTTP(r.Context(), logger, req, tripcodeResolver, tripcodeMemory, tripcodeSynthesizer, costTracker))
	})
	mux.HandleFunc("GET /resolve", func(w http.ResponseWriter, r *http.Request) {
		if !authorizedDemoRequest(r) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing or invalid demo key"})
			return
		}
		req := tripCodeRequestFromQuery(r.URL.Query())
		writeTripCodeHTTPResult(w, resolveTripCodeHTTP(r.Context(), logger, req, tripcodeResolver, tripcodeMemory, tripcodeSynthesizer, costTracker))
	})
	mux.HandleFunc("POST /resolve", func(w http.ResponseWriter, r *http.Request) {
		if !authorizedDemoRequest(r) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing or invalid demo key"})
			return
		}
		req, err := tripCodeRequestFromResolvePost(r)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid resolve request"})
			return
		}
		writeTripCodeHTTPResult(w, resolveTripCodeHTTP(r.Context(), logger, req, tripcodeResolver, tripcodeMemory, tripcodeSynthesizer, costTracker))
	})
	mux.HandleFunc("GET /v1/usage", func(w http.ResponseWriter, r *http.Request) {
		if !authorizedDemoRequest(r) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing or invalid demo key"})
			return
		}
		writeJSON(w, http.StatusOK, costTracker.Snapshot())
	})
	return mux
}

type tripCodeHTTPResult struct {
	status int
	body   any
}

func resolveTripCodeHTTP(
	ctx context.Context,
	logger *slog.Logger,
	req agent.TripCodeResearchRequest,
	tripcodeResolver agent.TripCodeResolver,
	tripcodeMemory *agent.TripCodeMemoryStore,
	tripcodeSynthesizer agent.TripCodeSynthesizer,
	costTracker *agent.CostTracker,
) tripCodeHTTPResult {
	if strings.TrimSpace(req.TripCode) == "" {
		if strings.TrimSpace(req.SessionID) == "" {
			return tripCodeHTTPResult{status: http.StatusBadRequest, body: map[string]string{"error": "tripcode or session_id is required"}}
		}
		snapshot := tripcodeMemory.Snapshot(req.SessionID)
		if !snapshot.Available {
			return tripCodeHTTPResult{status: http.StatusNotFound, body: map[string]string{"error": "no TripCode session memory found"}}
		}
		resp := agent.NewSessionMemoryResponse(req, snapshot)
		resp.ExecutionTrace = []agent.ExecutionTraceStep{
			traceStep(1, "judge-harness", "Submitted a session-memory follow-up request.", "session_id="+req.SessionID),
			traceStep(2, "google-agent", "Loaded prior TripCode/River context from the in-memory session store.", "turns="+strconv.Itoa(snapshot.Turns)),
			traceStep(3, "google-agent", "Returned the stateful follow-up packet without rerunning the TripCode resolver.", "mode=session-memory"),
		}
		resp.Cost = costTracker.Record("session-memory")
		return tripCodeHTTPResult{status: http.StatusOK, body: resp}
	}
	if tripcodeResolver == nil {
		return tripCodeHTTPResult{status: http.StatusServiceUnavailable, body: map[string]string{
			"error": "live DeltaSignal MCP TripCode resolver is not configured",
		}}
	}
	resp, err := tripcodeResolver.ResolveTripCodeResearchPacket(ctx, req)
	if err != nil {
		logger.Error("resolve tripcode failed", "error", err)
		return tripCodeHTTPResult{status: http.StatusBadGateway, body: map[string]string{"error": err.Error()}}
	}
	resp.ExecutionTrace = []agent.ExecutionTraceStep{
		traceStep(1, "judge-harness", "Submitted a high-level TripCode research goal.", "tripcode="+req.TripCode),
		traceStep(2, "google-agent", "Validated request flags and selected the TripCode/River resolver workflow.", requestTraceEvidence(req)),
		traceStep(3, "google-agent", "Called the configured ATLAS-7 MCP TripCode resolver.", "tool=deltasignal_resolve_tripcode_research_packet"),
		traceStep(4, resolverTraceActor(resp.Packet), resolverTraceAction(resp.Packet), packetTraceEvidence(resp.Packet)),
	}
	if strings.TrimSpace(req.SessionID) != "" {
		snapshot := tripcodeMemory.Remember(req.SessionID, resp)
		resp.Memory = &snapshot
		resp.ExecutionTrace = append(resp.ExecutionTrace, traceStep(5, "google-agent", "Wrote a session memory entry for second-turn River continuity.", "session_id="+req.SessionID+", turns="+strconv.Itoa(snapshot.Turns)))
	}
	if req.IncludeAgentContext {
		contextSnapshot, err := agent.FetchDefaultAgentContext(ctx)
		if err != nil {
			logger.Warn("agent context fetch failed", "error", err)
			contextSnapshot = agent.AgentContextSnapshot{
				Enabled: true,
				Purpose: "Public ATLAS-7 operating guide fetch was requested but failed; do not claim public guide context was used.",
				Sources: []agent.AgentContextSource{{Status: "unavailable", Error: err.Error()}},
			}
		}
		resp.AgentContext = &contextSnapshot
		resp.ExecutionTrace = append(resp.ExecutionTrace, traceStep(nextTraceOrder(resp.ExecutionTrace), "google-agent", "Fetched public ATLAS-7 agent operating guides for autonomous workflow context.", agentContextTraceEvidence(contextSnapshot)))
	}
	if tripcodeSynthesizer != nil {
		if generated, err := tripcodeSynthesizer.SynthesizeTripCode(ctx, req, resp); err == nil && strings.TrimSpace(generated) != "" {
			resp.GeminiSummary = generated
			resp.Mode = tripcodeGeminiMode(resp.Mode, resp.Packet)
			resp.ExecutionTrace = append(resp.ExecutionTrace, traceStep(nextTraceOrder(resp.ExecutionTrace), "gemini", "Synthesized the judge-facing diligence summary from packet, memory, public agent context, and boundaries.", "model=configured Gemini model"))
		} else if err != nil {
			logger.Warn("tripcode Gemini synthesis skipped", "error", err)
			resp.ExecutionTrace = append(resp.ExecutionTrace, traceStep(nextTraceOrder(resp.ExecutionTrace), "google-agent", "Skipped Gemini synthesis after a synthesis error and preserved the raw MCP packet.", "error="+err.Error()))
		}
	}
	resp.Cost = costTracker.Record("tripcode")
	resp.ExecutionTrace = append(resp.ExecutionTrace, traceStep(nextTraceOrder(resp.ExecutionTrace), "google-agent", "Recorded local-estimate Google credit usage for this request.", "request_kind=tripcode"))
	return tripCodeHTTPResult{status: http.StatusOK, body: resp}
}

func writeTripCodeHTTPResult(w http.ResponseWriter, result tripCodeHTTPResult) {
	writeJSON(w, result.status, result.body)
}

func tripCodeRequestFromResolvePost(r *http.Request) (agent.TripCodeResearchRequest, error) {
	req := tripCodeRequestFromQuery(r.URL.Query())
	raw, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return agent.TripCodeResearchRequest{}, err
	}
	if strings.TrimSpace(string(raw)) == "" {
		return req, nil
	}
	if strings.Contains(strings.ToLower(r.Header.Get("Content-Type")), "application/json") {
		var bodyReq agent.TripCodeResearchRequest
		if err := json.Unmarshal(raw, &bodyReq); err != nil {
			return agent.TripCodeResearchRequest{}, err
		}
		return mergeTripCodeRequest(bodyReq, req), nil
	}
	req.Question = strings.TrimSpace(string(raw))
	return req, nil
}

func tripCodeRequestFromQuery(values url.Values) agent.TripCodeResearchRequest {
	return agent.TripCodeResearchRequest{
		TripCode:              values.Get("tripcode"),
		Issuer:                values.Get("issuer"),
		SessionID:             values.Get("session_id"),
		Question:              values.Get("question"),
		PayloadMode:           values.Get("payload_mode"),
		IncludeArticleBody:    queryBool(values, "include_article_body"),
		IncludeFilingEvidence: queryBool(values, "include_filing_evidence"),
		IncludePriorArticles:  queryBool(values, "include_prior_articles"),
		IncludeThesisMap:      queryBool(values, "include_thesis_map"),
		IncludeAgentContext:   queryBool(values, "include_agent_context"),
	}
}

func mergeTripCodeRequest(base, override agent.TripCodeResearchRequest) agent.TripCodeResearchRequest {
	if strings.TrimSpace(override.TripCode) != "" {
		base.TripCode = override.TripCode
	}
	if strings.TrimSpace(override.Issuer) != "" {
		base.Issuer = override.Issuer
	}
	if strings.TrimSpace(override.SessionID) != "" {
		base.SessionID = override.SessionID
	}
	if strings.TrimSpace(override.Question) != "" {
		base.Question = override.Question
	}
	if strings.TrimSpace(override.PayloadMode) != "" {
		base.PayloadMode = override.PayloadMode
	}
	base.IncludeArticleBody = base.IncludeArticleBody || override.IncludeArticleBody
	base.IncludeFilingEvidence = base.IncludeFilingEvidence || override.IncludeFilingEvidence
	base.IncludePriorArticles = base.IncludePriorArticles || override.IncludePriorArticles
	base.IncludeThesisMap = base.IncludeThesisMap || override.IncludeThesisMap
	base.IncludeAgentContext = base.IncludeAgentContext || override.IncludeAgentContext
	return base
}

func traceStep(order int, actor, action, evidence string) agent.ExecutionTraceStep {
	return agent.ExecutionTraceStep{Order: order, Actor: actor, Action: action, Evidence: evidence}
}

func nextTraceOrder(steps []agent.ExecutionTraceStep) int {
	maxOrder := 0
	for _, step := range steps {
		if step.Order > maxOrder {
			maxOrder = step.Order
		}
	}
	return maxOrder + 1
}

func requestTraceEvidence(req agent.TripCodeResearchRequest) string {
	flags := []string{"payload_mode=" + req.PayloadMode}
	if req.IncludeFilingEvidence {
		flags = append(flags, "include_filing_evidence=true")
	}
	if req.IncludePriorArticles {
		flags = append(flags, "include_prior_articles=true")
	}
	if req.IncludeThesisMap {
		flags = append(flags, "include_thesis_map=true")
	}
	if req.IncludeAgentContext {
		flags = append(flags, "include_agent_context=true")
	}
	return strings.Join(flags, ", ")
}

func packetTraceEvidence(packet map[string]any) string {
	if len(packet) == 0 {
		return "packet_keys=0"
	}
	keys := make([]string, 0, len(packet))
	for key := range packet {
		keys = append(keys, key)
	}
	return "packet_keys=" + strings.Join(keys, ",")
}

func resolverTraceActor(packet map[string]any) string {
	if _, ok := packet["live_mcp_error"]; ok {
		return "google-agent"
	}
	return "mcp"
}

func resolverTraceAction(packet map[string]any) string {
	if _, ok := packet["live_mcp_error"]; ok {
		return "Detected live MCP unavailability and used the deterministic HUT fixture fallback while preserving live_mcp_error in the packet."
	}
	return "Returned the bounded research packet with article memory, River context, evidence refs, caveats, and provenance."
}

func tripcodeGeminiMode(mode string, packet map[string]any) string {
	if _, ok := packet["live_mcp_error"]; ok {
		if strings.TrimSpace(mode) == "" {
			return "demo-tripcode-river-gemini"
		}
		return mode + "-gemini"
	}
	return "live-mcp-tripcode-gemini"
}

func agentContextTraceEvidence(snapshot agent.AgentContextSnapshot) string {
	parts := make([]string, 0, len(snapshot.Sources))
	for _, source := range snapshot.Sources {
		label := source.URL
		if label == "" {
			label = "unknown"
		}
		parts = append(parts, label+" status="+source.Status+" bytes="+strconv.Itoa(source.Bytes)+" sha256="+source.SHA256)
	}
	return strings.Join(parts, " | ")
}

func queryBool(values url.Values, key string) bool {
	value := strings.TrimSpace(values.Get(key))
	if value == "" {
		return false
	}
	parsed, err := strconv.ParseBool(value)
	return err == nil && parsed
}

func ToolClientFromEnv() agent.ToolClient {
	if strings.EqualFold(os.Getenv("DELTASIGNAL_USE_LIVE_MCP"), "true") {
		apiKey := firstNonEmpty(os.Getenv("MCP_API_KEY"), os.Getenv("DELTASIGNAL_API_KEY"))
		if strings.TrimSpace(apiKey) != "" {
			return agent.AtlasMCPToolClient{
				BaseURL:      envOr("DELTASIGNAL_ATLAS_BASE_URL", "https://api.aitrailblazer.net"),
				EndpointPath: envOr("DELTASIGNAL_MCP_ENDPOINT_PATH", "/mcp"),
				APIKey:       apiKey,
				StressTool:   envOr("DELTASIGNAL_MCP_STRESS_TOOL", agent.DefaultStressTool),
				CompanyTool:  envOr("DELTASIGNAL_MCP_COMPANY_TOOL", agent.DefaultCompanyTool),
				PeerTool:     envOr("DELTASIGNAL_MCP_PEER_TOOL", agent.DefaultPeerTool),
				TripCodeTool: envOr("DELTASIGNAL_MCP_TRIPCODE_TOOL", agent.DefaultTripCodeResearchPacketTool),
			}
		}
	}
	return agent.DemoToolClient{}
}

func TripCodeResolverFromEnv() agent.TripCodeResolver {
	apiKey := firstNonEmpty(os.Getenv("MCP_API_KEY"), os.Getenv("DELTASIGNAL_API_KEY"))
	fallback := agent.DemoTripCodeResolver{}
	if strings.TrimSpace(apiKey) == "" {
		if strings.EqualFold(os.Getenv("DELTASIGNAL_ENABLE_TRIPCODE_FALLBACK"), "true") {
			return fallback
		}
		return nil
	}
	live := agent.AtlasMCPToolClient{
		BaseURL:      envOr("DELTASIGNAL_ATLAS_BASE_URL", "https://api.aitrailblazer.net"),
		EndpointPath: envOr("DELTASIGNAL_MCP_ENDPOINT_PATH", "/mcp"),
		APIKey:       apiKey,
		TripCodeTool: envOr("DELTASIGNAL_MCP_TRIPCODE_TOOL", agent.DefaultTripCodeResearchPacketTool),
	}
	if strings.EqualFold(os.Getenv("DELTASIGNAL_ENABLE_TRIPCODE_FALLBACK"), "true") {
		return agent.FallbackTripCodeResolver{Primary: live, Fallback: fallback}
	}
	return live
}

func CostTrackerFromEnv() *agent.CostTracker {
	return agent.NewCostTracker(agent.CostTrackerConfig{
		Enabled:              strings.EqualFold(os.Getenv("DELTASIGNAL_COST_TRACKING"), "true"),
		BudgetUSD:            envFloat("DELTASIGNAL_GOOGLE_CREDIT_BUDGET_USD", 0),
		BriefCostUSD:         envFloat("DELTASIGNAL_ESTIMATED_BRIEF_COST_USD", 0),
		TripCodeCostUSD:      envFloat("DELTASIGNAL_ESTIMATED_TRIPCODE_COST_USD", 0),
		SessionMemoryCostUSD: envFloat("DELTASIGNAL_ESTIMATED_SESSION_MEMORY_COST_USD", 0),
		Source:               envOr("DELTASIGNAL_COST_SOURCE", "local-estimate"),
		Billing: agent.BillingTrackerConfig{
			Available:          envBool("DELTASIGNAL_OFFICIAL_BILLING_AVAILABLE", false),
			Source:             envOr("DELTASIGNAL_OFFICIAL_BILLING_SOURCE", "unavailable"),
			ProjectID:          firstNonEmpty(os.Getenv("DELTASIGNAL_BILLING_PROJECT_ID"), os.Getenv("GOOGLE_CLOUD_PROJECT"), os.Getenv("CLOUDSDK_CORE_PROJECT")),
			BillingAccountName: os.Getenv("DELTASIGNAL_BILLING_ACCOUNT_NAME"),
			BillingEnabled:     envBool("DELTASIGNAL_BILLING_ENABLED", false),
			SpentUSD:           envFloat("DELTASIGNAL_OFFICIAL_BILLING_SPENT_USD", 0),
			CreditBudgetUSD:    envFloat("DELTASIGNAL_OFFICIAL_BILLING_CREDIT_BUDGET_USD", 0),
			RemainingUSD:       envFloat("DELTASIGNAL_OFFICIAL_BILLING_REMAINING_USD", 0),
			UpdatedAt:          os.Getenv("DELTASIGNAL_OFFICIAL_BILLING_UPDATED_AT"),
		},
	})
}

func envFloat(key string, fallback float64) float64 {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fallback
	}
	return parsed
}

func envBool(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envOr(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func authorizedDemoRequest(r *http.Request) bool {
	want := strings.TrimSpace(os.Getenv("DELTASIGNAL_DEMO_API_KEY"))
	if want == "" {
		return true
	}
	got := strings.TrimSpace(r.Header.Get("X-Demo-Key"))
	if got == "" {
		got = strings.TrimPrefix(strings.TrimSpace(r.Header.Get("Authorization")), "Bearer ")
	}
	return got == want
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
