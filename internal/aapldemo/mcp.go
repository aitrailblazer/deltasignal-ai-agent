package aapldemo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const DefaultMCPEndpoint = "https://api.aitrailblazer.net/mcp"

type MCPClient struct {
	Endpoint   string
	APIKey     string
	HTTPClient *http.Client
}

func (c MCPClient) CallTool(ctx context.Context, tool string, arguments map[string]any) (json.RawMessage, int, error) {
	payload, err := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      "google-cloud-aapl-client-demo",
		"method":  "tools/call",
		"params": map[string]any{
			"name":      tool,
			"arguments": arguments,
		},
	})
	if err != nil {
		return nil, 0, fmt.Errorf("encode MCP request: %w", err)
	}
	endpoint := strings.TrimSpace(c.Endpoint)
	if endpoint == "" {
		endpoint = DefaultMCPEndpoint
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, 0, fmt.Errorf("build MCP request: %w", err)
	}
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("Content-Type", "application/json")
	if key := strings.TrimSpace(c.APIKey); key != "" {
		req.Header.Set("X-API-Key", key)
	}
	client := c.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 25 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("call MCP tool %s: %w", tool, err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("read MCP tool %s: %w", tool, err)
	}
	if resp.StatusCode == http.StatusPaymentRequired || resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return raw, resp.StatusCode, AccessRequiredError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("DeltaSignal MCP access required for %s (HTTP %d); no payment was attempted", tool, resp.StatusCode),
		}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return raw, resp.StatusCode, fmt.Errorf("MCP tool %s returned HTTP %d", tool, resp.StatusCode)
	}

	var envelope struct {
		Result json.RawMessage `json:"result"`
		Error  *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return raw, resp.StatusCode, fmt.Errorf("decode MCP tool %s: %w", tool, err)
	}
	if envelope.Error != nil {
		return raw, resp.StatusCode, fmt.Errorf("MCP tool %s error %d: %s", tool, envelope.Error.Code, envelope.Error.Message)
	}
	if len(envelope.Result) == 0 {
		return raw, resp.StatusCode, fmt.Errorf("MCP tool %s returned no result", tool)
	}
	return envelope.Result, resp.StatusCode, nil
}
