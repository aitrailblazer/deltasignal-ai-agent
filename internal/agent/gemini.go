package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/genai"
)

type GeminiSynthesizer struct {
	Model    string
	Generate func(ctx context.Context, model string, prompt string) (string, error)
}

func (g GeminiSynthesizer) Synthesize(ctx context.Context, req BriefRequest, plan []string, findings []SpecialistResult) (string, error) {
	payload, _ := json.MarshalIndent(struct {
		Request          BriefRequest            `json:"request"`
		Plan             []string                `json:"plan"`
		Findings         []SpecialistResult      `json:"findings"`
		EvidenceFidelity EvidenceFidelitySummary `json:"evidence_fidelity"`
	}{req, plan, findings, BuildEvidenceFidelitySummary(findings)}, "", "  ")

	prompt := "You are DeltaSignal Gemini AI Agent. Create a concise B2B issuer intelligence brief. Use only the provided evidence. Include: key signal, evidence, uncertainty, and next action. Preserve evidence_fidelity fields such as source dates, computed timestamps, stale markers, caveats, quality flags, evidence hashes, payload mode, and provenance labels when present. Do not invent missing provenance.\n\n" + string(payload)
	return g.generate(ctx, prompt, "generate Gemini brief")
}

func (g GeminiSynthesizer) SynthesizeTripCode(ctx context.Context, req TripCodeResearchRequest, response TripCodeResearchResponse) (string, error) {
	payload, err := json.MarshalIndent(struct {
		Request        TripCodeResearchRequest `json:"request"`
		TripCode       string                  `json:"tripcode"`
		Issuer         string                  `json:"issuer,omitempty"`
		Packet         map[string]any          `json:"packet"`
		Memory         *TripCodeMemorySnapshot `json:"memory,omitempty"`
		AgentContext   *AgentContextSnapshot   `json:"agent_context,omitempty"`
		ExecutionTrace []ExecutionTraceStep    `json:"execution_trace,omitempty"`
		Boundaries     []string                `json:"boundaries"`
	}{
		Request:        req,
		TripCode:       response.TripCode,
		Issuer:         response.Issuer,
		Packet:         response.Packet,
		Memory:         response.Memory,
		AgentContext:   response.AgentContext,
		ExecutionTrace: response.ExecutionTrace,
		Boundaries:     response.Disclosures,
	}, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal TripCode synthesis context: %w", err)
	}

	prompt := strings.Join([]string{
		"You are DeltaSignal Gemini AI Agent.",
		"Use only the supplied TripCode research packet, session memory, and boundaries.",
		"If agent_context is present, treat it as public operating guidance fetched by the cloud agent. Do not claim a guide was read unless its source status is fetched.",
		"Use execution_trace only as observable workflow proof. Do not expose hidden chain-of-thought.",
		"Return a compact diligence summary with: what changed, what was confirmed or weakened, current execution bridge risk, and what to monitor next.",
		"Preserve all boundaries. Do not provide investment advice, price targets, recommendations, or order instructions.",
		"",
		string(payload),
	}, "\n")
	return g.generate(ctx, prompt, "generate Gemini TripCode summary")
}

func (g GeminiSynthesizer) generate(ctx context.Context, prompt, operation string) (string, error) {
	model := g.model()
	if g.Generate != nil {
		text, err := g.Generate(ctx, model, prompt)
		if err != nil {
			return "", fmt.Errorf("%s: %w", operation, err)
		}
		return text, nil
	}
	return generateGeminiContent(ctx, model, prompt, operation)
}

func (g GeminiSynthesizer) model() string {
	model := strings.TrimSpace(g.Model)
	if model == "" {
		return "gemini-2.5-flash"
	}
	return model
}

var generateGeminiContent = func(ctx context.Context, model string, prompt string, operation string) (string, error) {
	client, err := newGeminiClient(ctx)
	if err != nil {
		return "", err
	}
	resp, err := client.Models.GenerateContent(ctx, model, genai.Text(prompt), nil)
	if err != nil {
		return "", fmt.Errorf("%s: %w", operation, err)
	}
	return resp.Text(), nil
}

var newGeminiClient = func(ctx context.Context) (*genai.Client, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		HTTPOptions: genai.HTTPOptions{APIVersion: "v1"},
	})
	if err != nil {
		return nil, fmt.Errorf("create Gemini client: %w", err)
	}
	return client, nil
}
