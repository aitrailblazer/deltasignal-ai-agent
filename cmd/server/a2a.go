package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/aitrailblazer/deltasignal-ai-agent/internal/agent"
)

var tripCodePattern = regexp.MustCompile(`(?i)\bTF-SUB-[A-Z0-9]{6,}\b`)

type a2aAgentCard struct {
	ProtocolVersion    string            `json:"protocolVersion"`
	Name               string            `json:"name"`
	Description        string            `json:"description"`
	URL                string            `json:"url"`
	Version            string            `json:"version"`
	Capabilities       map[string]bool   `json:"capabilities"`
	Skills             []a2aSkill        `json:"skills"`
	DefaultInputModes  []string          `json:"defaultInputModes"`
	DefaultOutputModes []string          `json:"defaultOutputModes"`
	Provider           map[string]string `json:"provider,omitempty"`
}

type a2aSkill struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Examples    []string `json:"examples,omitempty"`
}

type a2aRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type a2aRPCResponse struct {
	JSONRPC string       `json:"jsonrpc"`
	ID      any          `json:"id,omitempty"`
	Result  any          `json:"result,omitempty"`
	Error   *a2aRPCError `json:"error,omitempty"`
}

type a2aRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type a2aMessageParams struct {
	Message  a2aMessage     `json:"message"`
	Metadata map[string]any `json:"metadata"`
}

type a2aMessage struct {
	Parts []a2aPart `json:"parts"`
}

type a2aPart struct {
	Kind string `json:"kind"`
	Type string `json:"type"`
	Text string `json:"text"`
}

func registerA2ARoutes(
	mux *http.ServeMux,
	logger *slog.Logger,
	tripcodeResolver agent.TripCodeResolver,
	tripcodeMemory *agent.TripCodeMemoryStore,
	tripcodeSynthesizer agent.TripCodeSynthesizer,
	costTracker *agent.CostTracker,
	rateLimiter *RateLimiter,
) {
	card := func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, buildA2AAgentCard(r))
	}
	mux.HandleFunc("GET /.well-known/agent-card.json", card)
	mux.HandleFunc("GET /a2a/app/.well-known/agent-card.json", card)
	mux.HandleFunc("POST /a2a", func(w http.ResponseWriter, r *http.Request) {
		rt := beginRequest(r, "/a2a")
		if !authorizedDemoRequest(r) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing or invalid demo key"})
			rt.Log(logger, http.StatusUnauthorized, "a2a-unauthorized")
			return
		}
		rate := rateLimiter.Allow(r)
		writeRateLimitHeaders(w, rate)
		if !rate.Allowed {
			writeJSON(w, http.StatusTooManyRequests, map[string]any{"error": "rate limit exceeded", "rate_limit": rate})
			rt.Log(logger, http.StatusTooManyRequests, "a2a-rate-limited")
			return
		}
		var rpc a2aRPCRequest
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&rpc); err != nil {
			writeJSON(w, http.StatusBadRequest, a2aRPCResponse{JSONRPC: "2.0", Error: &a2aRPCError{Code: -32700, Message: "invalid JSON-RPC request"}})
			rt.Log(logger, http.StatusBadRequest, "a2a-bad-json")
			return
		}
		result, status, mode := handleA2AMessage(r.Context(), logger, rpc, tripcodeResolver, tripcodeMemory, tripcodeSynthesizer, costTracker, rt, &rate)
		writeJSON(w, status, result)
		rt.Log(logger, status, mode)
	})
}

func buildA2AAgentCard(r *http.Request) a2aAgentCard {
	base := publicBaseURL(r)
	return a2aAgentCard{
		ProtocolVersion: envOr("DELTASIGNAL_A2A_PROTOCOL_VERSION", "0.3.0"),
		Name:            "DeltaSignal Gemini AI Agent",
		Description:     "Issuer diligence agent that resolves DeltaSignal TripCodes into River memory, evidence boundaries, thesis deltas, and monitor-next workflows.",
		URL:             strings.TrimRight(base, "/") + "/a2a",
		Version:         envOr("DELTASIGNAL_AGENT_VERSION", "1.0.0"),
		Capabilities: map[string]bool{
			"streaming":         false,
			"pushNotifications": false,
		},
		Skills: []a2aSkill{
			{
				ID:          "resolve-tripcode-research-packet",
				Name:        "Resolve TripCode research packet",
				Description: "Resolve a TF-SUB TripCode into article memory, River continuity, filing evidence references, thesis evolution, caveats, and monitor-next items.",
				Tags:        []string{"finance", "issuer-diligence", "tripcode", "research-memory", "sec-xbrl"},
				Examples:    []string{"Resolve TF-SUB-9DA70A7F98 and show what changed across the HUT River."},
			},
			{
				ID:          "monitor-tripcode-thesis",
				Name:        "Monitor TripCode thesis",
				Description: "Use a TF-SUB TripCode as a post-publication thesis-monitor baseline for confirmed signals, weakened assumptions, stale evidence, invalidation checks, and monitor-next actions.",
				Tags:        []string{"finance", "issuer-monitoring", "tripcode", "river-memory", "evidence-boundary"},
				Examples:    []string{"Monitor TF-SUB-9DA70A7F98 for weakened assumptions and next evidence checks."},
			},
		},
		DefaultInputModes:  []string{"text/plain", "application/json"},
		DefaultOutputModes: []string{"application/json"},
		Provider: map[string]string{
			"organization": "AITrailblazer",
			"url":          "https://aitrailblazer.com",
		},
	}
}

func publicBaseURL(r *http.Request) string {
	if configured := strings.TrimSpace(osGetenv("DELTASIGNAL_PUBLIC_BASE_URL")); configured != "" {
		return strings.TrimRight(configured, "/")
	}
	scheme := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto"))
	if scheme == "" {
		scheme = "http"
	}
	host := strings.TrimSpace(r.Host)
	if host == "" {
		host = "localhost"
	}
	return scheme + "://" + host
}

var osGetenv = getenv

func handleA2AMessage(
	ctx context.Context,
	logger *slog.Logger,
	rpc a2aRPCRequest,
	tripcodeResolver agent.TripCodeResolver,
	tripcodeMemory *agent.TripCodeMemoryStore,
	tripcodeSynthesizer agent.TripCodeSynthesizer,
	costTracker *agent.CostTracker,
	rt requestRuntime,
	rate *agent.RateLimitSnapshot,
) (a2aRPCResponse, int, string) {
	if rpc.JSONRPC == "" {
		rpc.JSONRPC = "2.0"
	}
	if rpc.Method != "message/send" && rpc.Method != "tasks/send" && rpc.Method != "resolve_tripcode" && rpc.Method != "monitor_tripcode_thesis" {
		return a2aRPCResponse{JSONRPC: "2.0", ID: rpc.ID, Error: &a2aRPCError{Code: -32601, Message: "unsupported A2A method"}}, http.StatusBadRequest, "a2a-unsupported-method"
	}
	text, sessionID := a2aTextAndSession(rpc.Params)
	tripcode := firstTripCode(text)
	monitor := rpc.Method == "monitor_tripcode_thesis" || containsMonitorIntent(text)
	if tripcode == "" {
		if strings.TrimSpace(sessionID) == "" {
			return a2aRPCResponse{JSONRPC: "2.0", ID: rpc.ID, Result: a2aInputRequiredTask(rpc.ID, "Send a TF-SUB TripCode, for example TF-SUB-9DA70A7F98.")}, http.StatusOK, "a2a-input-required"
		}
		req := agent.TripCodeResearchRequest{
			SessionID: sessionID,
			Question:  text,
		}
		result := resolveTripCodeHTTP(ctx, logger, req, tripcodeResolver, tripcodeMemory, tripcodeSynthesizer, costTracker, rt, rate)
		if result.status < 200 || result.status > 299 {
			return a2aRPCResponse{JSONRPC: "2.0", ID: rpc.ID, Error: &a2aRPCError{Code: -32000, Message: "TripCode session follow-up failed"}}, result.status, "a2a-session-memory-error"
		}
		if monitor {
			return a2aRPCResponse{JSONRPC: "2.0", ID: rpc.ID, Result: a2aMonitorTask(rpc.ID, "session memory", result.body)}, http.StatusOK, "a2a-monitor-session-memory"
		}
		return a2aRPCResponse{JSONRPC: "2.0", ID: rpc.ID, Result: a2aCompletedTask(rpc.ID, "session memory", result.body)}, http.StatusOK, "a2a-session-memory"
	}
	req := agent.TripCodeResearchRequest{
		TripCode:              tripcode,
		SessionID:             sessionID,
		Question:              text,
		PayloadMode:           "compact",
		IncludeFilingEvidence: true,
		IncludePriorArticles:  true,
		IncludeThesisMap:      true,
	}
	result := resolveTripCodeHTTP(ctx, logger, req, tripcodeResolver, tripcodeMemory, tripcodeSynthesizer, costTracker, rt, rate)
	if result.status < 200 || result.status > 299 {
		return a2aRPCResponse{JSONRPC: "2.0", ID: rpc.ID, Error: &a2aRPCError{Code: -32000, Message: "TripCode resolution failed"}}, result.status, "a2a-resolve-error"
	}
	if monitor {
		return a2aRPCResponse{JSONRPC: "2.0", ID: rpc.ID, Result: a2aMonitorTask(rpc.ID, tripcode, result.body)}, http.StatusOK, "a2a-monitor-tripcode"
	}
	return a2aRPCResponse{JSONRPC: "2.0", ID: rpc.ID, Result: a2aCompletedTask(rpc.ID, tripcode, result.body)}, http.StatusOK, "a2a-tripcode"
}

func a2aTextAndSession(raw json.RawMessage) (string, string) {
	var params a2aMessageParams
	if len(raw) > 0 && json.Unmarshal(raw, &params) == nil {
		parts := make([]string, 0, len(params.Message.Parts))
		for _, part := range params.Message.Parts {
			if strings.TrimSpace(part.Text) != "" {
				parts = append(parts, strings.TrimSpace(part.Text))
			}
		}
		if len(parts) > 0 || stringMetadata(params.Metadata, "session_id") != "" {
			return strings.Join(parts, "\n"), stringMetadata(params.Metadata, "session_id")
		}
	}
	var fallback map[string]any
	_ = json.Unmarshal(raw, &fallback)
	return stringMetadata(fallback, "text"), stringMetadata(fallback, "session_id")
}

func stringMetadata(values map[string]any, key string) string {
	if values == nil {
		return ""
	}
	if got, ok := values[key].(string); ok {
		return strings.TrimSpace(got)
	}
	return ""
}

func firstTripCode(text string) string {
	found := tripCodePattern.FindString(text)
	return strings.ToUpper(strings.TrimSpace(found))
}

func containsMonitorIntent(text string) bool {
	lower := strings.ToLower(text)
	for _, needle := range []string{"monitor", "watch", "recheck", "post-publication", "stale evidence", "invalidation"} {
		if strings.Contains(lower, needle) {
			return true
		}
	}
	return false
}

func a2aInputRequiredTask(id any, message string) map[string]any {
	return map[string]any{
		"id": taskID(id),
		"status": map[string]any{
			"state": "input-required",
			"message": map[string]any{
				"role":  "agent",
				"parts": []map[string]string{{"kind": "text", "text": message}},
			},
		},
		"metadata": a2aMetadata(),
	}
}

func a2aCompletedTask(id any, tripcode string, body any) map[string]any {
	return a2aTaskWithArtifact(
		id,
		"deltasignal_research_packet",
		"Resolved DeltaSignal TripCode research packet for "+tripcode+".",
		body,
	)
}

func a2aMonitorTask(id any, label string, body any) map[string]any {
	return a2aTaskWithArtifact(
		id,
		"deltasignal_thesis_monitor_baseline",
		"Post-publication DeltaSignal thesis monitor baseline for "+label+".",
		a2aMonitorBaseline(label, body),
	)
}

func a2aTaskWithArtifact(id any, name, description string, body any) map[string]any {
	return map[string]any{
		"id": taskID(id),
		"status": map[string]any{
			"state": "completed",
		},
		"artifacts": []map[string]any{{
			"name":        name,
			"description": description,
			"parts":       []map[string]any{{"kind": "data", "data": body}},
		}},
		"metadata": a2aMetadata(),
	}
}

func a2aMonitorBaseline(label string, body any) map[string]any {
	return map[string]any{
		"monitor_baseline": map[string]any{
			"status": "ready",
			"skill":  "monitor-tripcode-thesis",
			"input":  strings.TrimSpace(label),
			"scope":  "post-publication thesis monitoring",
			"required_checks": []string{
				"confirmed signals",
				"weakened assumptions",
				"stale or missing evidence",
				"invalidation criteria",
				"monitor-next actions",
			},
			"evidence_boundary": "Monitor output preserves the underlying research packet and does not upgrade TripCode/River memory into SEC/XBRL evidence.",
			"non_advice":        true,
		},
		"research_packet": body,
	}
}

func taskID(id any) string {
	if id == nil {
		return "a2a-" + shortHash("missing-id")
	}
	encoded, err := json.Marshal(id)
	if err != nil {
		return "a2a-" + shortHash("unmarshalable-id")
	}
	return "a2a-" + shortHash(string(encoded))
}

func shortHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])[:12]
}

func a2aMetadata() map[string]any {
	return map[string]any{
		"track":          "Track 3 Path",
		"marketplace":    "agent-card-ready",
		"gemini_surface": "Gemini Enterprise A2A registration candidate",
		"boundary":       "TripCodes are proprietary resolver keys; official filing identifiers remain SEC accessions, CIKs, filing forms, periods, and XBRL concepts.",
		"non_advice":     true,
	}
}

func a2aRegisterCurl(agentCardURL string) string {
	escapedCard := strings.ReplaceAll(agentCardURL, `"`, `\"`)
	return `agents-cli publish gemini-enterprise --registration-type a2a --agent-card-url "` + escapedCard + `" --display-name "DeltaSignal Gemini AI Agent"`
}

func a2aMarketplaceCardPath(bucket string) string {
	bucket = strings.Trim(strings.TrimSpace(bucket), "/")
	if bucket == "" {
		return "gs://<marketplace-agent-card-bucket>/deltasignal-gemini-ai-agent/agent-card.json"
	}
	if strings.HasPrefix(bucket, "gs://") {
		return strings.TrimRight(bucket, "/") + "/deltasignal-gemini-ai-agent/agent-card.json"
	}
	return "gs://" + url.PathEscape(bucket) + "/deltasignal-gemini-ai-agent/agent-card.json"
}
