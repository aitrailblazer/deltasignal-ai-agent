package aapldemo

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/genai"
)

type GeminiSynthesizer struct {
	Model string
}

func (g GeminiSynthesizer) Synthesize(ctx context.Context, req Request, results []ToolResult) (string, error) {
	model := strings.TrimSpace(g.Model)
	if model == "" {
		model = "gemini-2.5-flash"
	}
	payload, err := json.MarshalIndent(map[string]any{
		"request":                    req,
		"bounded_specialist_results": results,
	}, "", "  ")
	if err != nil {
		return "", err
	}
	prompt := strings.Join([]string{
		"You are the review agent in the DeltaSignal Google Cloud AAPL client demo.",
		"Use only the supplied MCP evidence envelopes.",
		"Separate: filing facts, calculations, ATLAS-7 interpretations, scenarios, unresolved claims, and access status.",
		"Never infer a missing ATLAS-7 level. Never interpret payment or entitlement as evidence quality.",
		"Return a concise analyst brief with headings: Established, Interpreted, Missing or Partial, Provenance to Inspect.",
		"Do not provide investment advice, a price target, or an order instruction.",
		"",
		string(payload),
	}, "\n")
	client, err := genai.NewClient(ctx, &genai.ClientConfig{HTTPOptions: genai.HTTPOptions{APIVersion: "v1"}})
	if err != nil {
		return "", fmt.Errorf("create Google Gen AI client: %w", err)
	}
	response, err := client.Models.GenerateContent(ctx, model, genai.Text(prompt), nil)
	if err != nil {
		return "", fmt.Errorf("Gemini synthesis: %w", err)
	}
	return strings.TrimSpace(response.Text()), nil
}
