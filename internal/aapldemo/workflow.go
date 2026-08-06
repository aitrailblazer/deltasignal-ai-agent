package aapldemo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

var DefaultDisclosures = []string{
	"Filing facts, calculations, ATLAS-7 interpretations, scenarios, and unresolved claims are separate evidence classes.",
	"Missing evidence stays missing. Payment or entitlement changes access, not evidence existence.",
	"Level applicability is not a directional investment signal.",
	"Delta Signal outputs are evidence-routing alerts and diligence triage only; they are not investment advice.",
}

type AccessRequiredError struct {
	StatusCode int
	Message    string
}

func (e AccessRequiredError) Error() string {
	if strings.TrimSpace(e.Message) == "" {
		return "DeltaSignal MCP access is required"
	}
	return e.Message
}

type Synthesizer interface {
	Synthesize(ctx context.Context, req Request, results []ToolResult) (string, error)
}

type Workflow struct {
	Caller      ToolCaller
	Synthesizer Synthesizer
	Now         func() time.Time
}

func (w Workflow) Run(ctx context.Context, req Request) (Response, error) {
	if w.Caller == nil {
		return Response{}, fmt.Errorf("AAPL demo tool caller is not configured")
	}
	req.Ticker = strings.ToUpper(strings.TrimSpace(req.Ticker))
	if req.Ticker == "" {
		req.Ticker = "AAPL"
	}
	if req.Ticker != "AAPL" {
		return Response{}, fmt.Errorf("this bounded demo supports AAPL only")
	}
	if strings.TrimSpace(req.Question) == "" {
		req.Question = "What can the filing evidence establish about Apple, and which ATLAS-7 levels remain unsupported?"
	}
	if strings.TrimSpace(req.Mode) == "" {
		req.Mode = "fixture"
	}

	specs := EvidencePlan(req)
	results := make([]ToolResult, len(specs))
	var wg sync.WaitGroup
	for i, spec := range specs {
		wg.Add(1)
		go func(index int, item ToolSpec) {
			defer wg.Done()
			results[index] = runTool(ctx, w.Caller, item)
		}(i, spec)
	}
	wg.Wait()

	response := Response{
		Request:     req,
		GeneratedAt: w.now(),
		Agents:      results,
		Disclosures: append([]string(nil), DefaultDisclosures...),
	}
	for _, result := range results {
		if result.Status != StatusAvailable {
			response.Partial = true
		}
		if result.Status == StatusAvailable && hasEvidenceGap(result.Data) {
			response.Partial = true
		}
		if result.Status == StatusAccessRequired {
			response.AccessNeeded = true
		}
	}

	synth := w.Synthesizer
	if synth == nil {
		synth = DeterministicSynthesizer{}
	}
	text, err := synth.Synthesize(ctx, req, results)
	if err != nil {
		response.Partial = true
		response.Synthesis = DeterministicSynthesizer{}.mustSynthesize(req, results) +
			"\n\nGemini synthesis unavailable: " + err.Error()
		return response, nil
	}
	response.Synthesis = text
	return response, nil
}

func EvidencePlan(req Request) []ToolSpec {
	tickerArgs := func() map[string]any { return map[string]any{"ticker": req.Ticker} }
	pointArgs := tickerArgs()
	pointArgs["mode"] = "compact"
	pointArgs["limit"] = 25
	if strings.TrimSpace(req.AsOfDate) != "" {
		pointArgs["as_of_date_to"] = req.AsOfDate
	}
	return []ToolSpec{
		{
			Agent:     "filing-facts-agent",
			Tool:      "deltasignal_company_fundamentals",
			Arguments: tickerArgs(),
			Purpose:   "Retrieve filing-backed financial facts and filing provenance.",
		},
		{
			Agent: "companyfacts-history-agent",
			Tool:  "deltasignal_atlas7_companyfacts_history",
			Arguments: map[string]any{
				"ticker": req.Ticker,
				"mode":   "summary",
				"limit":  25,
			},
			Purpose: "Inspect the materialized CompanyFacts inventory and evidence pointers.",
		},
		{
			Agent:     "point-in-time-agent",
			Tool:      "deltasignal_atlas7_point_in_time_history",
			Arguments: pointArgs,
			Purpose:   "Check what was knowable by the requested date without look-ahead.",
		},
		{
			Agent: "applicability-agent",
			Tool:  "deltasignal_atlas7_four_level_applicability",
			Arguments: map[string]any{
				"ticker": req.Ticker,
				"mode":   "full",
				"limit":  1,
			},
			Purpose: "Report which ATLAS-7 levels are evidence-backed, partial, or missing.",
		},
	}
}

func runTool(ctx context.Context, caller ToolCaller, spec ToolSpec) ToolResult {
	result := ToolResult{
		Agent:     spec.Agent,
		Tool:      spec.Tool,
		Purpose:   spec.Purpose,
		Arguments: spec.Arguments,
		Status:    StatusUnavailable,
	}
	data, status, err := caller.CallTool(ctx, spec.Tool, spec.Arguments)
	result.HTTPStatus = status
	if err == nil {
		result.Status = StatusAvailable
		result.Data = data
		return result
	}
	var accessErr AccessRequiredError
	if errors.As(err, &accessErr) {
		result.Status = StatusAccessRequired
		result.Error = accessErr.Error()
		if result.HTTPStatus == 0 {
			result.HTTPStatus = accessErr.StatusCode
		}
		return result
	}
	result.Error = err.Error()
	return result
}

func (w Workflow) now() time.Time {
	if w.Now != nil {
		return w.Now().UTC()
	}
	return time.Now().UTC()
}

func hasEvidenceGap(data json.RawMessage) bool {
	normalized := strings.ToLower(string(data))
	return strings.Contains(normalized, `\"status\":\"missing\"`) ||
		strings.Contains(normalized, `\"status\":\"partial\"`) ||
		strings.Contains(normalized, `"status":"missing"`) ||
		strings.Contains(normalized, `"status":"partial"`)
}

type DeterministicSynthesizer struct{}

func (DeterministicSynthesizer) Synthesize(_ context.Context, req Request, results []ToolResult) (string, error) {
	return DeterministicSynthesizer{}.mustSynthesize(req, results), nil
}

func (DeterministicSynthesizer) mustSynthesize(req Request, results []ToolResult) string {
	available, access, missing := 0, 0, 0
	for _, result := range results {
		switch result.Status {
		case StatusAvailable:
			available++
		case StatusAccessRequired:
			access++
		default:
			missing++
		}
	}
	return fmt.Sprintf(
		"%s evidence workflow completed %d of %d bounded specialist calls. "+
			"%d call(s) require access and %d call(s) are unavailable. "+
			"Inspect each raw MCP evidence envelope before drawing a conclusion; unsupported levels and unresolved attribution must remain explicit.",
		req.Ticker, available, len(results), access, missing,
	)
}

func CompactJSON(value any) string {
	raw, _ := json.Marshal(value)
	return string(raw)
}
