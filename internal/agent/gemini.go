package agent

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

func (g GeminiSynthesizer) Synthesize(ctx context.Context, req BriefRequest, plan []string, findings []SpecialistResult) (string, error) {
	model := strings.TrimSpace(g.Model)
	if model == "" {
		model = "gemini-2.5-flash"
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		HTTPOptions: genai.HTTPOptions{APIVersion: "v1"},
	})
	if err != nil {
		return "", fmt.Errorf("create Gemini client: %w", err)
	}

	payload, err := json.MarshalIndent(struct {
		Request  BriefRequest       `json:"request"`
		Plan     []string           `json:"plan"`
		Findings []SpecialistResult `json:"findings"`
	}{req, plan, findings}, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal synthesis context: %w", err)
	}

	prompt := "You are DeltaSignal AI Agent. Create a concise B2B issuer intelligence brief. Use only the provided evidence. Include: key signal, evidence, uncertainty, and next action.\n\n" + string(payload)
	resp, err := client.Models.GenerateContent(ctx, model, genai.Text(prompt), nil)
	if err != nil {
		return "", fmt.Errorf("generate Gemini brief: %w", err)
	}
	return resp.Text(), nil
}
