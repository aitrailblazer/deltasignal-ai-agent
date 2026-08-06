package aapldemo

import (
	"context"
	"encoding/json"
	"time"
)

const (
	StatusAvailable      = "available"
	StatusAccessRequired = "access_required"
	StatusUnavailable    = "unavailable"
)

// Request is the bounded client question sent to the AAPL evidence workflow.
type Request struct {
	Ticker   string `json:"ticker"`
	Question string `json:"question"`
	AsOfDate string `json:"as_of_date,omitempty"`
	Mode     string `json:"mode"`
}

// ToolSpec defines one evidence surface used by the public client demo.
type ToolSpec struct {
	Agent     string         `json:"agent"`
	Tool      string         `json:"tool"`
	Arguments map[string]any `json:"arguments"`
	Purpose   string         `json:"purpose"`
}

// ToolResult preserves the MCP evidence envelope instead of flattening it into prose.
type ToolResult struct {
	Agent      string          `json:"agent"`
	Tool       string          `json:"tool"`
	Purpose    string          `json:"purpose"`
	Arguments  map[string]any  `json:"arguments"`
	Status     string          `json:"status"`
	HTTPStatus int             `json:"http_status,omitempty"`
	Data       json.RawMessage `json:"data,omitempty"`
	Error      string          `json:"error,omitempty"`
}

// Response is safe for deterministic fixture snapshots and live Cloud Run output.
type Response struct {
	Request      Request      `json:"request"`
	GeneratedAt  time.Time    `json:"generated_at"`
	Agents       []ToolResult `json:"agents"`
	Synthesis    string       `json:"synthesis"`
	Disclosures  []string     `json:"disclosures"`
	Partial      bool         `json:"partial"`
	AccessNeeded bool         `json:"access_needed"`
}

// ToolCaller is implemented by the live MCP client and deterministic fixture client.
type ToolCaller interface {
	CallTool(ctx context.Context, tool string, arguments map[string]any) (json.RawMessage, int, error)
}
