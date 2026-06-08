package agent

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type mcpRoundTripFunc func(*http.Request) (*http.Response, error)

func (f mcpRoundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

type mcpErrorReader struct{}

func (mcpErrorReader) Read([]byte) (int, error) {
	return 0, errors.New("read failed")
}

func (mcpErrorReader) Close() error { return nil }

func TestAtlasMCPToolClientCallsMCPWithAPIKey(t *testing.T) {
	var gotKey string
	var gotMethod string
	var gotTool string
	var gotTicker string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("X-API-Key")
		var body struct {
			Method string `json:"method"`
			Params struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments"`
			} `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		gotMethod = body.Method
		gotTool = body.Params.Name
		gotTicker, _ = body.Params.Arguments["ticker"].(string)

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"deltasignal-ai-agent","result":{"content":[{"type":"text","text":"HUT live evidence packet"}]}}`))
	}))
	defer server.Close()

	client := AtlasMCPToolClient{
		BaseURL:     server.URL,
		APIKey:      "test-key",
		CompanyTool: "deltasignal_company_report",
	}
	result, err := client.CompanyEvidence(t.Context(), "hut")
	if err != nil {
		t.Fatalf("CompanyEvidence returned error: %v", err)
	}
	if gotKey != "test-key" {
		t.Fatalf("X-API-Key = %q, want test-key", gotKey)
	}
	if gotMethod != "tools/call" {
		t.Fatalf("method = %q, want tools/call", gotMethod)
	}
	if gotTool != "deltasignal_company_report" {
		t.Fatalf("tool = %q, want deltasignal_company_report", gotTool)
	}
	if gotTicker != "HUT" {
		t.Fatalf("ticker = %q, want HUT", gotTicker)
	}
	if result.Confidence != "live-mcp" {
		t.Fatalf("confidence = %q, want live-mcp", result.Confidence)
	}
	if result.Summary != "HUT live evidence packet" {
		t.Fatalf("summary = %q, want live evidence text", result.Summary)
	}
	if result.Evidence[0].Source != "deltasignal-atlas-7-mcp" {
		t.Fatalf("evidence source = %q, want live MCP source", result.Evidence[0].Source)
	}
}

func TestAtlasMCPToolClientReportsMCPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"deltasignal-ai-agent","error":{"code":-32000,"message":"payment_required"}}`))
	}))
	defer server.Close()

	client := AtlasMCPToolClient{
		BaseURL: server.URL,
		APIKey:  "test-key",
	}
	if _, err := client.PeerContext(t.Context(), "HUT"); err == nil {
		t.Fatal("PeerContext returned nil error, want MCP error")
	}
}

func TestAtlasMCPToolClientResolvesTripCodeResearchPacket(t *testing.T) {
	var gotKey string
	var gotTool string
	var gotTripCode string
	var gotIssuer string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("X-API-Key")
		var body struct {
			Params struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments"`
			} `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		gotTool = body.Params.Name
		gotTripCode, _ = body.Params.Arguments["tripcode"].(string)
		gotIssuer, _ = body.Params.Arguments["issuer"].(string)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"jsonrpc": "2.0",
			"id":      "deltasignal-ai-agent",
			"result": map[string]any{
				"content": []map[string]any{{"type": "text", "text": "resolved HUT TripCode research packet"}},
				"structuredContent": map[string]any{
					"data": map[string]any{
						"tool":     "deltasignal_resolve_tripcode_research_packet",
						"status":   "ready",
						"tripcode": "TF-SUB-9DA70A7F98",
						"research_packet": map[string]any{
							"article": map[string]any{"tripcode": "TF-SUB-9DA70A7F98"},
							"river":   map[string]any{"node_count": float64(10)},
						},
					},
				},
			},
		})
	}))
	defer server.Close()

	client := AtlasMCPToolClient{
		BaseURL:      server.URL,
		APIKey:       "test-key",
		TripCodeTool: "deltasignal_resolve_tripcode_research_packet",
	}
	result, err := client.ResolveTripCodeResearchPacket(t.Context(), TripCodeResearchRequest{
		TripCode: "tf-sub-9da70a7f98",
		Issuer:   "hut",
	})
	if err != nil {
		t.Fatalf("ResolveTripCodeResearchPacket returned error: %v", err)
	}
	if gotKey != "test-key" {
		t.Fatalf("X-API-Key = %q, want test-key", gotKey)
	}
	if gotTool != DefaultTripCodeResearchPacketTool {
		t.Fatalf("tool = %q, want %q", gotTool, DefaultTripCodeResearchPacketTool)
	}
	if gotTripCode != "TF-SUB-9DA70A7F98" || gotIssuer != "HUT" {
		t.Fatalf("arguments = tripcode %q issuer %q", gotTripCode, gotIssuer)
	}
	if result.Mode != "live-mcp-tripcode" {
		t.Fatalf("mode = %q, want live-mcp-tripcode", result.Mode)
	}
	if result.TripCode != "TF-SUB-9DA70A7F98" {
		t.Fatalf("tripcode = %q", result.TripCode)
	}
	if result.Packet["tool"] != "deltasignal_resolve_tripcode_research_packet" {
		t.Fatalf("unexpected packet: %#v", result.Packet)
	}
	if result.Packet["status"] != "ready" {
		t.Fatalf("structured data not preserved: %#v", result.Packet)
	}
	if result.Packet["raw_text"] != nil {
		t.Fatalf("structured packet should not collapse to raw_text: %#v", result.Packet)
	}
	if result.Packet["mcp_text_summary"] != "resolved HUT TripCode research packet" {
		t.Fatalf("text summary not preserved: %#v", result.Packet)
	}
	if len(result.Disclosures) == 0 {
		t.Fatal("expected evidence boundary disclosures")
	}
}

func TestAtlasMCPToolClientStressSignalsAndDefaults(t *testing.T) {
	var gotTool string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Params struct {
				Name string `json:"name"`
			} `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		gotTool = body.Params.Name
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","result":{"summary":"stress summary"}}`))
	}))
	defer server.Close()

	client := AtlasMCPToolClient{BaseURL: server.URL, APIKey: "key"}
	result, err := client.StressSignals(t.Context(), "hut")
	if err != nil {
		t.Fatalf("StressSignals returned error: %v", err)
	}
	if gotTool != DefaultStressTool || result.Summary != "stress summary" {
		t.Fatalf("unexpected stress result tool=%q result=%#v", gotTool, result)
	}
	if client.companyTool() != DefaultCompanyTool || client.peerTool() != DefaultPeerTool || client.tripCodeTool() != DefaultTripCodeResearchPacketTool {
		t.Fatal("default tool names not returned")
	}
	if (AtlasMCPToolClient{StressTool: " custom "}).stressTool() != "custom" {
		t.Fatal("custom stress tool not trimmed")
	}
	if (AtlasMCPToolClient{PeerTool: " peer-custom "}).peerTool() != "peer-custom" {
		t.Fatal("custom peer tool not trimmed")
	}
}

func TestAtlasMCPToolClientResolveTripCodeErrors(t *testing.T) {
	client := AtlasMCPToolClient{APIKey: "key"}
	if _, err := client.ResolveTripCodeResearchPacket(t.Context(), TripCodeResearchRequest{TripCode: "bad"}); err == nil {
		t.Fatal("expected invalid TripCode error")
	}
	if _, err := (AtlasMCPToolClient{}).ResolveTripCodeResearchPacket(t.Context(), TripCodeResearchRequest{TripCode: "TF-SUB-X"}); err == nil {
		t.Fatal("expected missing API key error")
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","result":{}}`))
	}))
	defer server.Close()
	client = AtlasMCPToolClient{BaseURL: server.URL, APIKey: "key"}
	if _, err := client.ResolveTripCodeResearchPacket(t.Context(), TripCodeResearchRequest{TripCode: "TF-SUB-X"}); err == nil {
		t.Fatal("expected empty packet error")
	}

	var gotArgs map[string]any
	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Params struct {
				Arguments map[string]any `json:"arguments"`
			} `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		gotArgs = body.Params.Arguments
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","result":{"structuredContent":{"status":"ready"}}}`))
	}))
	defer server2.Close()
	client = AtlasMCPToolClient{BaseURL: server2.URL, APIKey: "key"}
	result, err := client.ResolveTripCodeResearchPacket(t.Context(), TripCodeResearchRequest{
		TripCode:              "TF-SUB-X",
		PayloadMode:           "full",
		IncludeFilingEvidence: true,
		IncludePriorArticles:  true,
		IncludeThesisMap:      true,
	})
	if err != nil {
		t.Fatalf("ResolveTripCodeResearchPacket returned error: %v", err)
	}
	if result.Issuer != "" || gotArgs["payload_mode"] != "full" || gotArgs["issuer"] != nil || gotArgs["include_thesis_map"] != true {
		t.Fatalf("unexpected args/result args=%#v result=%#v", gotArgs, result)
	}

	client = AtlasMCPToolClient{
		BaseURL: "http://example.test",
		APIKey:  "key",
		HTTPClient: &http.Client{Transport: mcpRoundTripFunc(func(*http.Request) (*http.Response, error) {
			return nil, errors.New("network failed")
		})},
	}
	if _, err := client.ResolveTripCodeResearchPacket(t.Context(), TripCodeResearchRequest{TripCode: "TF-SUB-X"}); err == nil {
		t.Fatal("expected callTool error")
	}
}

func TestAtlasMCPToolClientCallLaneErrors(t *testing.T) {
	if _, err := (AtlasMCPToolClient{}).callLane(t.Context(), "agent", "tool", nil); err == nil {
		t.Fatal("expected missing API key error")
	}
	if _, err := (AtlasMCPToolClient{APIKey: "key"}).callLane(t.Context(), "agent", "", nil); err == nil {
		t.Fatal("expected missing tool name error")
	}
	client := AtlasMCPToolClient{
		BaseURL: "http://example.test",
		APIKey:  "key",
		HTTPClient: &http.Client{Transport: mcpRoundTripFunc(func(*http.Request) (*http.Response, error) {
			return nil, errors.New("network failed")
		})},
	}
	if _, err := client.callLane(t.Context(), "agent", "tool", nil); err == nil {
		t.Fatal("expected call error")
	}
}

func TestAtlasMCPToolClientCallToolErrors(t *testing.T) {
	client := AtlasMCPToolClient{
		BaseURL: "http://example.test",
		APIKey:  "key",
	}
	if _, err := client.callTool(t.Context(), "tool", map[string]any{"bad": func() {}}); err == nil {
		t.Fatal("expected marshal error")
	}
	client.BaseURL = "://bad"
	if _, err := client.callTool(t.Context(), "tool", nil); err == nil {
		t.Fatal("expected request build error")
	}

	client = AtlasMCPToolClient{BaseURL: "http://example.test", APIKey: "key", HTTPClient: &http.Client{Transport: mcpRoundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: mcpErrorReader{}, Header: make(http.Header)}, nil
	})}}
	if _, err := client.callTool(t.Context(), "tool", nil); err == nil {
		t.Fatal("expected read error")
	}

	client.HTTPClient = &http.Client{Transport: mcpRoundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusPaymentRequired, Body: io.NopCloser(strings.NewReader("payment_required")), Header: make(http.Header)}, nil
	})}
	if _, err := client.callTool(t.Context(), "tool", nil); err == nil || !strings.Contains(err.Error(), "HTTP 402") {
		t.Fatalf("expected HTTP status error, got %v", err)
	}

	client.HTTPClient = &http.Client{Transport: mcpRoundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("{bad")), Header: make(http.Header)}, nil
	})}
	if _, err := client.callTool(t.Context(), "tool", nil); err == nil {
		t.Fatal("expected decode error")
	}

	client.HTTPClient = &http.Client{Transport: mcpRoundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"jsonrpc":"2.0","error":{"code":-1,"message":"bad"}}`)), Header: make(http.Header)}, nil
	})}
	if _, err := client.callTool(t.Context(), "tool", nil); err == nil || !strings.Contains(err.Error(), "error -1") {
		t.Fatalf("expected JSON-RPC error, got %v", err)
	}

	client.HTTPClient = &http.Client{Transport: mcpRoundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"jsonrpc":"2.0"}`)), Header: make(http.Header)}, nil
	})}
	if _, err := client.callTool(t.Context(), "tool", nil); err == nil {
		t.Fatal("expected no result error")
	}
}

func TestAtlasMCPHelpers(t *testing.T) {
	if (AtlasMCPToolClient{}).endpointURL() != "https://api.aitrailblazer.net/mcp" {
		t.Fatalf("default endpoint = %s", (AtlasMCPToolClient{}).endpointURL())
	}
	if got := (AtlasMCPToolClient{BaseURL: "http://host/base", EndpointPath: "mcp"}).endpointURL(); got != "http://host/base/mcp" {
		t.Fatalf("joined endpoint = %s", got)
	}
	if got := (AtlasMCPToolClient{BaseURL: "http://host/mcp", EndpointPath: "/mcp"}).endpointURL(); got != "http://host/mcp" {
		t.Fatalf("deduped endpoint = %s", got)
	}
	if got := (AtlasMCPToolClient{BaseURL: "://bad", EndpointPath: "mcp"}).endpointURL(); got != "://bad/mcp" {
		t.Fatalf("fallback endpoint = %s", got)
	}

	if summarizeMCPResult(map[string]any{"brief": " short brief "}) != "short brief" {
		t.Fatal("brief summary failed")
	}
	if !strings.Contains(summarizeMCPResult(map[string]any{"structuredContent": map[string]any{"a": "b"}}), `"a": "b"`) {
		t.Fatal("structured summary failed")
	}
	if !strings.Contains(summarizeMCPResult(map[string]any{"x": "y"}), `"x": "y"`) {
		t.Fatal("fallback JSON summary failed")
	}

	if parseMCPResultData(map[string]any{"content": []any{"bad", map[string]any{"text": ""}, map[string]any{"text": "not-json"}}})["raw_text"] != "not-json" {
		t.Fatal("raw text parse failed")
	}
	if parseMCPResultData(map[string]any{"content": []any{map[string]any{"text": `{"json_status":"ready"}`}}})["json_status"] != "ready" {
		t.Fatal("JSON text parse failed")
	}
	if parseMCPResultData(map[string]any{"structuredContent": map[string]any{"ok": true}})["ok"] != true {
		t.Fatal("structured parse failed")
	}
	nested := parseMCPResultData(map[string]any{
		"content": []any{map[string]any{"text": "human summary"}},
		"structuredContent": map[string]any{
			"data": map[string]any{"status": "ready"},
		},
	})
	if nested["status"] != "ready" || nested["mcp_text_summary"] != "human summary" || nested["raw_text"] != nil {
		t.Fatalf("nested structured data parse failed: %#v", nested)
	}
	preserved := parseMCPResultData(map[string]any{
		"content":           []any{map[string]any{"text": "new summary"}},
		"structuredContent": map[string]any{"mcp_text_summary": "existing summary"},
	})
	if preserved["mcp_text_summary"] != "existing summary" {
		t.Fatalf("existing text summary overwritten: %#v", preserved)
	}
	if parseMCPResultData(map[string]any{"structuredContent": map[string]any{}, "fallback": true})["fallback"] != true {
		t.Fatal("empty structured content should fall back to result")
	}
	if mcpContentText(map[string]any{"content": []any{"bad", map[string]any{"text": ""}, map[string]any{"text": " a "}, map[string]any{"text": "b"}}}) != "a\nb" {
		t.Fatal("content text extraction failed")
	}
	if summarizeJSON(func() {}) != "" {
		t.Fatal("summarizeJSON should return empty on marshal failure")
	}
	if compactText(" a  b ", 0) != "a b" || compactText("abcdef", 1) != "a" || compactText("abcdef", 4) != "abc..." {
		t.Fatal("compactText branches failed")
	}
}
