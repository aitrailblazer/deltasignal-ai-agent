package agent

import (
	"context"
	"fmt"
	"strings"
)

type ToolClient interface {
	StressSignals(ctx context.Context, issuer string) (SpecialistResult, error)
	CompanyEvidence(ctx context.Context, issuer string) (SpecialistResult, error)
	PeerContext(ctx context.Context, issuer string) (SpecialistResult, error)
}

type DemoToolClient struct{}

func (DemoToolClient) StressSignals(ctx context.Context, issuer string) (SpecialistResult, error) {
	issuer = normalizeIssuer(issuer)
	return SpecialistResult{
		Agent:      "stress-scanner",
		Summary:    fmt.Sprintf("%s shows a demo stress profile with elevated volatility, financing sensitivity, and catalyst dependence.", issuer),
		Confidence: "demo-medium",
		Evidence: []Evidence{
			{
				Source:      "demo-deltasignal",
				Title:       "Stress signal fixture",
				Observation: "Fixture represents the intended DeltaSignal MCP stress-scan output shape for the competition demo.",
			},
		},
	}, nil
}

func (DemoToolClient) CompanyEvidence(ctx context.Context, issuer string) (SpecialistResult, error) {
	issuer = normalizeIssuer(issuer)
	return SpecialistResult{
		Agent:      "evidence-retriever",
		Summary:    fmt.Sprintf("%s requires SEC-grounded validation before any investment or operating conclusion is promoted.", issuer),
		Confidence: "demo-high",
		Evidence: []Evidence{
			{
				Source:      "sec-companyfacts",
				Title:       "SEC evidence fixture",
				Observation: "Public-company claims should be grounded in filings, company facts, and dated source records.",
			},
		},
	}, nil
}

func (DemoToolClient) PeerContext(ctx context.Context, issuer string) (SpecialistResult, error) {
	issuer = normalizeIssuer(issuer)
	return SpecialistResult{
		Agent:      "peer-analyst",
		Summary:    fmt.Sprintf("%s should be evaluated against peer stress, liquidity, and catalyst profiles rather than in isolation.", issuer),
		Confidence: "demo-medium",
		Evidence: []Evidence{
			{
				Source:      "demo-peer-ranking",
				Title:       "Peer ranking fixture",
				Observation: "Fixture represents a peer-comparison tool result for judge-safe deterministic demos.",
			},
		},
	}, nil
}

func normalizeIssuer(issuer string) string {
	issuer = strings.TrimSpace(issuer)
	if issuer == "" {
		return "the issuer"
	}
	return strings.ToUpper(issuer)
}
