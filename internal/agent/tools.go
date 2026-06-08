package agent

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type ToolClient interface {
	StressSignals(ctx context.Context, issuer string) (SpecialistResult, error)
	CompanyEvidence(ctx context.Context, issuer string) (SpecialistResult, error)
	PeerContext(ctx context.Context, issuer string) (SpecialistResult, error)
}

type TripCodeResolver interface {
	ResolveTripCodeResearchPacket(ctx context.Context, req TripCodeResearchRequest) (TripCodeResearchResponse, error)
}

type FallbackTripCodeResolver struct {
	Primary  TripCodeResolver
	Fallback TripCodeResolver
}

func (r FallbackTripCodeResolver) ResolveTripCodeResearchPacket(ctx context.Context, req TripCodeResearchRequest) (TripCodeResearchResponse, error) {
	if r.Primary != nil {
		resp, err := r.Primary.ResolveTripCodeResearchPacket(ctx, req)
		if err == nil {
			return resp, nil
		}
		if r.Fallback == nil {
			return TripCodeResearchResponse{}, err
		}
		resp, fallbackErr := r.Fallback.ResolveTripCodeResearchPacket(ctx, req)
		if fallbackErr != nil {
			return TripCodeResearchResponse{}, fmt.Errorf("live TripCode resolver failed: %v; fallback resolver failed: %w", err, fallbackErr)
		}
		if resp.Packet == nil {
			resp.Packet = map[string]any{}
		}
		resp.Packet["live_mcp_error"] = compactText(err.Error(), 500)
		return resp, nil
	}
	if r.Fallback == nil {
		return TripCodeResearchResponse{}, fmt.Errorf("TripCode resolver is not configured")
	}
	return r.Fallback.ResolveTripCodeResearchPacket(ctx, req)
}

type DemoToolClient struct{}

type DemoTripCodeResolver struct{}

func (DemoTripCodeResolver) ResolveTripCodeResearchPacket(ctx context.Context, req TripCodeResearchRequest) (TripCodeResearchResponse, error) {
	tripcode := strings.ToUpper(strings.TrimSpace(req.TripCode))
	if !strings.HasPrefix(tripcode, "TF-SUB-") {
		return TripCodeResearchResponse{}, fmt.Errorf("tripcode must start with TF-SUB-")
	}
	issuer := normalizeIssuer(firstNonEmpty(req.Issuer, "HUT"))
	title := "Hut 8: The Re-Rating Has A Deadline"
	if tripcode != "TF-SUB-9DA70A7F98" {
		title = "DeltaSignal TripCode Research Packet"
	}
	packet := map[string]any{
		"article": map[string]any{
			"tripcode":       tripcode,
			"title":          title,
			"primary_issuer": issuer,
			"source":         "Delta Signal Substack research memory fixture",
			"thesis":         "A credible AI-infrastructure pivot meets a fragile valuation window.",
		},
		"river": map[string]any{
			"river_tripcode": "TF-RIVER-HUT-DEMO",
			"issuer":         issuer,
			"node_count":     10,
			"status":         "ready_with_demo_fixture",
			"prior_article_tripcodes": []string{
				"TF-SUB-6D75A87AD9",
				"TF-SUB-CAC4B8F9F4",
				"TF-SUB-8872A03DD7",
				"TF-SUB-09AC89D63C",
				"TF-SUB-C87F255122",
				"TF-SUB-D9F21CEF3F",
				"TF-SUB-A09241130B",
				"TF-SUB-314C30330B",
				"TF-SUB-EC6C5BF5F2",
			},
		},
		"thesis_map": map[string]any{
			"what_changed": []string{
				"Hut 8 moved from a miner balance-sheet story toward an AI-infrastructure execution story.",
				"The re-rating now depends on contracted compute revenue, site energization, and financing discipline.",
				"Prior crypto-treasury optionality is less central than proof that AI capacity can convert into durable cash flow.",
				"The deadline is evidence timing: valuation support weakens if filings do not show progress before expectations reset.",
			},
			"confirmed_signals": []string{
				"Prior DeltaSignal coverage correctly centered the crypto plus AI pivot and balance-sheet optionality.",
			},
			"weakened_assumptions": []string{
				"Bitcoin treasury value alone is not enough to carry the thesis.",
				"AI infrastructure announcements require proof of power, contracts, capex control, and operating conversion.",
				"Re-rating durability depends on filing-backed execution rather than narrative momentum.",
			},
			"execution_bridge_risks": []string{
				"Power, interconnect, and energization timing can lag market expectations.",
				"Debt, dilution, or capex timing can absorb upside before AI revenue scales.",
			},
			"proof_milestones_2027": []string{
				"Signed or expanded AI infrastructure contracts with clear economics.",
				"Filing evidence that AI-related revenue and margins are materializing.",
				"Capex and financing path that does not overwhelm operating cash flow.",
			},
			"invalidation_checklist": []string{
				"Delayed energization, weak contract economics, rising financing stress, or filings that fail to show AI revenue conversion.",
			},
			"monitor_next": []string{
				"Next 10-Q and 8-K updates.",
				"Power contract disclosures.",
				"AI hosting or colocation contract terms.",
				"Debt, dilution, and liquidity changes.",
				"Bitcoin treasury changes.",
				"Capex timing and site-level milestones.",
				"Peer AI-infrastructure pricing.",
				"Management language around 2027 utilization.",
			},
			"scenario_map": map[string]any{
				"bull": "AI infrastructure revenue arrives before valuation support decays.",
				"base": "Execution progresses, but financing and power timing keep the re-rating uneven.",
				"bear": "Narrative outruns filings, and capital costs absorb the AI pivot upside.",
			},
		},
		"evidence_boundary": map[string]any{
			"article_memory":  "Fixture reconstructs DeltaSignal article/River memory for judge demo continuity.",
			"filing_evidence": "Live TF-XBRL and TF-DS evidence should be treated as missing unless returned by ATLAS-7.",
			"non_advice":      "Diligence triage only; no investment advice, price target, or order instruction.",
		},
	}
	return TripCodeResearchResponse{
		TripCode:    tripcode,
		Issuer:      issuer,
		GeneratedAt: time.Now().UTC(),
		Mode:        "demo-tripcode-river",
		Packet:      packet,
		Disclosures: DefaultTripCodeDisclosures(),
	}, nil
}

func (DemoToolClient) StressSignals(ctx context.Context, issuer string) (SpecialistResult, error) {
	issuer = normalizeIssuer(issuer)
	return SpecialistResult{
		Agent:      "stress-scanner",
		Summary:    fmt.Sprintf("%s shows a demo stress profile with elevated volatility, financing sensitivity, and catalyst dependence.", issuer),
		Confidence: "demo-medium",
		Evidence: []Evidence{
			{
				Source:      "deltasignal-atlas-7-demo",
				Title:       "ATLAS-7 pressure-board fixture",
				Observation: "Fixture represents the intended DeltaSignal ATLAS-7 MCP pressure-board/top-stressed output shape for the competition demo.",
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
				Source:      "deltasignal-atlas-7-demo",
				Title:       "ATLAS-7 company-fundamentals fixture",
				Observation: "Public-company claims should be grounded in SEC/XBRL filings, company facts, and dated source records.",
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
				Source:      "deltasignal-atlas-7-demo",
				Title:       "ATLAS-7 peer-ranking fixture",
				Observation: "Fixture represents a peer-ranking tool result for judge-safe deterministic demos.",
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

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
