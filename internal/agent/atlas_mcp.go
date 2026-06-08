package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultAtlasBaseURL     = "https://api.aitrailblazer.net"
	defaultAtlasMCPEndpoint = "/mcp"

	DefaultStressTool                 = "deltasignal_pressure_board"
	DefaultCompanyTool                = "deltasignal_company_report"
	DefaultPeerTool                   = "deltasignal_peer_ranking"
	DefaultTripCodeResearchPacketTool = "deltasignal_resolve_tripcode_research_packet"
)

type AtlasMCPToolClient struct {
	BaseURL      string
	EndpointPath string
	APIKey       string
	HTTPClient   *http.Client
	StressTool   string
	CompanyTool  string
	PeerTool     string
	TripCodeTool string
}

func (c AtlasMCPToolClient) StressSignals(ctx context.Context, issuer string) (SpecialistResult, error) {
	return c.callLane(ctx, "stress-scanner", c.stressTool(), map[string]any{
		"output_mode": "compact",
	})
}

func (c AtlasMCPToolClient) CompanyEvidence(ctx context.Context, issuer string) (SpecialistResult, error) {
	ticker := normalizeIssuer(issuer)
	return c.callLane(ctx, "evidence-retriever", c.companyTool(), map[string]any{
		"ticker":      ticker,
		"output_mode": "compact",
	})
}

func (c AtlasMCPToolClient) PeerContext(ctx context.Context, issuer string) (SpecialistResult, error) {
	ticker := normalizeIssuer(issuer)
	return c.callLane(ctx, "peer-analyst", c.peerTool(), map[string]any{
		"ticker": ticker,
	})
}

func (c AtlasMCPToolClient) ResolveTripCodeResearchPacket(ctx context.Context, req TripCodeResearchRequest) (TripCodeResearchResponse, error) {
	tripcode := strings.ToUpper(strings.TrimSpace(req.TripCode))
	if !strings.HasPrefix(tripcode, "TF-SUB-") {
		return TripCodeResearchResponse{}, fmt.Errorf("tripcode must start with TF-SUB-")
	}
	if strings.TrimSpace(c.APIKey) == "" {
		return TripCodeResearchResponse{}, fmt.Errorf("DeltaSignal MCP API key is not configured")
	}

	payloadMode := strings.TrimSpace(req.PayloadMode)
	if payloadMode == "" {
		payloadMode = "compact"
	}

	args := map[string]any{
		"tripcode":                tripcode,
		"payload_mode":            payloadMode,
		"include_article_body":    req.IncludeArticleBody,
		"include_filing_evidence": true,
		"include_prior_articles":  true,
		"include_thesis_map":      true,
	}
	if strings.TrimSpace(req.Issuer) != "" {
		args["issuer"] = normalizeIssuer(req.Issuer)
	}
	if req.IncludeFilingEvidence {
		args["include_filing_evidence"] = true
	}
	if req.IncludePriorArticles {
		args["include_prior_articles"] = true
	}
	if req.IncludeThesisMap {
		args["include_thesis_map"] = true
	}

	result, err := c.callTool(ctx, c.tripCodeTool(), args)
	if err != nil {
		return TripCodeResearchResponse{}, err
	}
	packet := parseMCPResultData(result)
	if len(packet) == 0 {
		return TripCodeResearchResponse{}, fmt.Errorf("DeltaSignal MCP returned an empty TripCode packet")
	}

	return TripCodeResearchResponse{
		TripCode:    tripcode,
		Issuer:      normalizeOptionalIssuer(req.Issuer),
		GeneratedAt: time.Now().UTC(),
		Mode:        "live-mcp-tripcode",
		Packet:      packet,
		Disclosures: DefaultTripCodeDisclosures(),
	}, nil
}

func normalizeOptionalIssuer(issuer string) string {
	if strings.TrimSpace(issuer) == "" {
		return ""
	}
	return normalizeIssuer(issuer)
}

func (c AtlasMCPToolClient) callLane(ctx context.Context, agentName, toolName string, args map[string]any) (SpecialistResult, error) {
	if strings.TrimSpace(c.APIKey) == "" {
		return SpecialistResult{}, fmt.Errorf("DeltaSignal MCP API key is not configured")
	}
	if strings.TrimSpace(toolName) == "" {
		return SpecialistResult{}, fmt.Errorf("DeltaSignal MCP tool name is not configured")
	}

	result, err := c.callTool(ctx, toolName, args)
	if err != nil {
		return SpecialistResult{}, err
	}
	summary := summarizeMCPResult(result)
	evidence := enrichEvidenceFromMCP(Evidence{
		Source:      "deltasignal-atlas-7-mcp",
		Title:       "MCP tool: " + toolName,
		Observation: compactText(summary, 900),
		URL:         c.endpointURL(),
	}, result, args)
	return SpecialistResult{
		Agent:      agentName,
		Summary:    summary,
		Confidence: "live-mcp",
		Evidence:   []Evidence{evidence},
	}, nil
}

func (c AtlasMCPToolClient) callTool(ctx context.Context, toolName string, args map[string]any) (map[string]any, error) {
	body := map[string]any{
		"jsonrpc": "2.0",
		"id":      "deltasignal-ai-agent",
		"method":  "tools/call",
		"params": map[string]any{
			"name":      toolName,
			"arguments": args,
		},
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal MCP request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpointURL(), bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("build MCP request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", strings.TrimSpace(c.APIKey))

	client := c.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call DeltaSignal MCP %s: %w", toolName, err)
	}
	defer resp.Body.Close()

	limited := io.LimitReader(resp.Body, 2<<20)
	raw, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("read DeltaSignal MCP response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("DeltaSignal MCP %s returned HTTP %d: %s", toolName, resp.StatusCode, compactText(string(raw), 500))
	}

	var envelope struct {
		Result map[string]any `json:"result"`
		Error  *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return nil, fmt.Errorf("decode DeltaSignal MCP response: %w", err)
	}
	if envelope.Error != nil {
		return nil, fmt.Errorf("DeltaSignal MCP %s error %d: %s", toolName, envelope.Error.Code, envelope.Error.Message)
	}
	if envelope.Result == nil {
		return nil, fmt.Errorf("DeltaSignal MCP %s returned no result", toolName)
	}
	return envelope.Result, nil
}

func (c AtlasMCPToolClient) endpointURL() string {
	base := strings.TrimSpace(c.BaseURL)
	if base == "" {
		base = defaultAtlasBaseURL
	}
	path := strings.TrimSpace(c.EndpointPath)
	if path == "" {
		path = defaultAtlasMCPEndpoint
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	u, err := url.Parse(base)
	if err != nil {
		return strings.TrimRight(base, "/") + path
	}
	if strings.HasSuffix(strings.TrimRight(u.Path, "/"), strings.TrimRight(path, "/")) {
		return u.String()
	}
	u.Path = strings.TrimRight(u.Path, "/") + path
	return u.String()
}

func (c AtlasMCPToolClient) stressTool() string {
	if strings.TrimSpace(c.StressTool) != "" {
		return strings.TrimSpace(c.StressTool)
	}
	return DefaultStressTool
}

func (c AtlasMCPToolClient) companyTool() string {
	if strings.TrimSpace(c.CompanyTool) != "" {
		return strings.TrimSpace(c.CompanyTool)
	}
	return DefaultCompanyTool
}

func (c AtlasMCPToolClient) peerTool() string {
	if strings.TrimSpace(c.PeerTool) != "" {
		return strings.TrimSpace(c.PeerTool)
	}
	return DefaultPeerTool
}

func (c AtlasMCPToolClient) tripCodeTool() string {
	if strings.TrimSpace(c.TripCodeTool) != "" {
		return strings.TrimSpace(c.TripCodeTool)
	}
	return DefaultTripCodeResearchPacketTool
}

func summarizeMCPResult(result map[string]any) string {
	if text := strings.TrimSpace(mcpContentText(result)); text != "" {
		return compactText(text, 1600)
	}
	for _, key := range []string{"summary", "brief", "natural_brief", "status", "title"} {
		if text, ok := result[key].(string); ok && strings.TrimSpace(text) != "" {
			return compactText(text, 1600)
		}
	}
	if structured, ok := result["structuredContent"]; ok {
		if text := summarizeJSON(structured); text != "" {
			return compactText(text, 1600)
		}
	}
	return compactText(summarizeJSON(result), 1600)
}

func parseMCPResultData(result map[string]any) map[string]any {
	text := strings.TrimSpace(mcpContentText(result))
	if structured, ok := result["structuredContent"].(map[string]any); ok {
		packet := structured
		if data, ok := structured["data"].(map[string]any); ok {
			packet = data
		}
		if len(packet) > 0 {
			packet = cloneMap(packet)
			if text != "" {
				if _, exists := packet["mcp_text_summary"]; !exists {
					packet["mcp_text_summary"] = compactText(text, 1600)
				}
			}
			return packet
		}
	}

	content, ok := result["content"].([]any)
	if ok {
		for _, item := range content {
			entry, ok := item.(map[string]any)
			if !ok {
				continue
			}
			text, ok := entry["text"].(string)
			if !ok || strings.TrimSpace(text) == "" {
				continue
			}
			var parsed map[string]any
			if err := json.Unmarshal([]byte(text), &parsed); err == nil {
				return parsed
			}
			return map[string]any{"raw_text": text}
		}
	}
	return result
}

func cloneMap(input map[string]any) map[string]any {
	out := make(map[string]any, len(input))
	for key, value := range input {
		out[key] = value
	}
	return out
}

func mcpContentText(result map[string]any) string {
	content, ok := result["content"].([]any)
	if !ok {
		return ""
	}
	var parts []string
	for _, item := range content {
		entry, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if text, ok := entry["text"].(string); ok && strings.TrimSpace(text) != "" {
			parts = append(parts, strings.TrimSpace(text))
		}
	}
	return strings.Join(parts, "\n")
}

func summarizeJSON(v any) string {
	raw, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return ""
	}
	return string(raw)
}

func compactText(text string, limit int) string {
	text = strings.Join(strings.Fields(strings.TrimSpace(text)), " ")
	if limit <= 0 || len(text) <= limit {
		return text
	}
	if limit <= 1 {
		return text[:limit]
	}
	return strings.TrimSpace(text[:limit-1]) + "..."
}
