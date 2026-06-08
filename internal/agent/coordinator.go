package agent

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type Synthesizer interface {
	Synthesize(ctx context.Context, req BriefRequest, plan []string, findings []SpecialistResult) (string, error)
}

type TripCodeSynthesizer interface {
	SynthesizeTripCode(ctx context.Context, req TripCodeResearchRequest, response TripCodeResearchResponse) (string, error)
}

type Coordinator struct {
	Tools       ToolClient
	Synthesizer Synthesizer
	Clock       func() time.Time
}

func (c Coordinator) BuildBrief(ctx context.Context, req BriefRequest) (BriefResponse, error) {
	if strings.TrimSpace(req.Issuer) == "" {
		req.Issuer = "MSTR"
	}
	if strings.TrimSpace(req.Question) == "" {
		req.Question = "What issuer stress or opportunity should an analyst investigate next?"
	}

	tools := c.Tools
	if tools == nil {
		tools = DemoToolClient{}
	}

	plan := []string{
		"Decompose the user request into issuer stress, company evidence, peer context, and review tasks.",
		"Call bounded tools for each specialist lane.",
		"Ground the final answer in returned evidence and label uncertainty.",
		"Return one analyst-ready action brief.",
	}

	findings := make([]SpecialistResult, 0, 3)
	for _, call := range []func(context.Context, string) (SpecialistResult, error){
		tools.StressSignals,
		tools.CompanyEvidence,
		tools.PeerContext,
	} {
		result, err := call(ctx, req.Issuer)
		if err != nil {
			return BriefResponse{}, err
		}
		findings = append(findings, result)
	}

	brief := fallbackBrief(req, findings)
	mode := toolMode(findings)
	if c.Synthesizer != nil {
		if generated, err := c.Synthesizer.Synthesize(ctx, req, plan, findings); err == nil && strings.TrimSpace(generated) != "" {
			brief = generated
			mode = "vertex-ai-gemini"
		}
	}

	now := time.Now().UTC()
	if c.Clock != nil {
		now = c.Clock().UTC()
	}

	return BriefResponse{
		Issuer:      strings.ToUpper(strings.TrimSpace(req.Issuer)),
		Question:    req.Question,
		GeneratedAt: now,
		Mode:        mode,
		Plan:        plan,
		Findings:    findings,
		Brief:       brief,
		NextAction:  "Review the cited evidence packet, then run a live DeltaSignal/SEC refresh before making any external claim or business decision.",
		Disclosures: []string{
			"Competition build: existing DeltaSignal systems are treated as authorized integrations or data sources.",
			"Demo mode uses deterministic fixtures so judges can evaluate the workflow even without private data access.",
			"Production mode should route Gemini through Vertex AI in project startup-ai-deltasignal.",
		},
	}, nil
}

func fallbackBrief(req BriefRequest, findings []SpecialistResult) string {
	parts := []string{
		fmt.Sprintf("DeltaSignal Gemini AI Agent reviewed %s for: %s", strings.ToUpper(req.Issuer), req.Question),
	}
	for _, finding := range findings {
		parts = append(parts, fmt.Sprintf("%s: %s", finding.Agent, finding.Summary))
	}
	parts = append(parts, "Assessment: investigate the issuer only after refreshing live evidence and comparing the stress signal against peers.")
	return strings.Join(parts, "\n")
}

func toolMode(findings []SpecialistResult) string {
	for _, finding := range findings {
		if strings.EqualFold(finding.Confidence, "live-mcp") {
			return "live-mcp"
		}
		for _, evidence := range finding.Evidence {
			if evidence.Source == "deltasignal-atlas-7-mcp" {
				return "live-mcp"
			}
		}
	}
	return "deterministic-demo"
}
