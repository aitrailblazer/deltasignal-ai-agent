package agent

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

type testSynthesizer struct {
	text string
	err  error
}

func (s testSynthesizer) Synthesize(context.Context, BriefRequest, []string, []SpecialistResult) (string, error) {
	return s.text, s.err
}

func TestCoordinatorBuildBriefUsesDeterministicDemo(t *testing.T) {
	c := Coordinator{
		Tools: DemoToolClient{},
		Clock: func() time.Time {
			return time.Date(2026, 6, 4, 12, 0, 0, 0, time.UTC)
		},
	}

	resp, err := c.BuildBrief(context.Background(), BriefRequest{
		Issuer:   "mstr",
		Question: "What changed?",
	})
	if err != nil {
		t.Fatalf("BuildBrief returned error: %v", err)
	}
	if resp.Issuer != "MSTR" {
		t.Fatalf("issuer = %q, want MSTR", resp.Issuer)
	}
	if resp.Mode != "deterministic-demo" {
		t.Fatalf("mode = %q, want deterministic-demo", resp.Mode)
	}
	if len(resp.Findings) != 3 {
		t.Fatalf("findings len = %d, want 3", len(resp.Findings))
	}
	if !strings.Contains(resp.Brief, "stress-scanner") {
		t.Fatalf("brief did not include stress scanner output: %s", resp.Brief)
	}
	if resp.EvidenceFidelity.Status != "minimal" {
		t.Fatalf("demo evidence fidelity = %#v", resp.EvidenceFidelity)
	}
	if !strings.Contains(resp.Brief, "Evidence fidelity:") {
		t.Fatalf("brief did not include evidence fidelity line: %s", resp.Brief)
	}
}

func TestCoordinatorBuildBriefDefaultsAndSynthesizer(t *testing.T) {
	c := Coordinator{
		Tools:       DemoToolClient{},
		Synthesizer: testSynthesizer{text: "generated brief"},
	}
	resp, err := c.BuildBrief(context.Background(), BriefRequest{})
	if err != nil {
		t.Fatalf("BuildBrief returned error: %v", err)
	}
	if resp.Issuer != "MSTR" || resp.Question == "" || resp.Mode != "vertex-ai-gemini" || !strings.Contains(resp.Brief, "generated brief") {
		t.Fatalf("unexpected synthesized response: %#v", resp)
	}
	if !strings.Contains(resp.Brief, "Evidence fidelity:") {
		t.Fatalf("synthesized brief did not preserve evidence fidelity line: %s", resp.Brief)
	}
}

func TestEnsureEvidenceFidelityLine(t *testing.T) {
	summary := EvidenceFidelitySummary{Status: "minimal"}
	if got := ensureEvidenceFidelityLine("", summary); got != evidenceFidelityLine(summary) {
		t.Fatalf("empty brief got %q", got)
	}
	existing := "brief\nEvidence fidelity: already present"
	if got := ensureEvidenceFidelityLine(existing, summary); got != existing {
		t.Fatalf("existing line should not be changed: %q", got)
	}
}

func TestCoordinatorBuildBriefSynthesizerFallbacks(t *testing.T) {
	for _, synth := range []Synthesizer{
		testSynthesizer{text: "   "},
		testSynthesizer{err: errors.New("boom")},
	} {
		resp, err := (Coordinator{Tools: DemoToolClient{}, Synthesizer: synth}).BuildBrief(context.Background(), BriefRequest{Issuer: "HUT"})
		if err != nil {
			t.Fatalf("BuildBrief returned error: %v", err)
		}
		if resp.Mode != "deterministic-demo" || !strings.Contains(resp.Brief, "DeltaSignal Gemini AI Agent reviewed HUT") {
			t.Fatalf("unexpected fallback response: %#v", resp)
		}
	}
}

func TestCoordinatorBuildBriefToolErrorAndNilTools(t *testing.T) {
	if _, err := (Coordinator{Tools: failingAgentTool{}}).BuildBrief(context.Background(), BriefRequest{Issuer: "HUT"}); err == nil {
		t.Fatal("expected tool error")
	}
	resp, err := (Coordinator{}).BuildBrief(context.Background(), BriefRequest{Issuer: "HUT"})
	if err != nil {
		t.Fatalf("nil tools should fall back to demo: %v", err)
	}
	if resp.Issuer != "HUT" {
		t.Fatalf("issuer = %q", resp.Issuer)
	}
}

func TestToolModeLiveMCPBranches(t *testing.T) {
	if toolMode([]SpecialistResult{{Confidence: "live-mcp"}}) != "live-mcp" {
		t.Fatal("live confidence not detected")
	}
	if toolMode([]SpecialistResult{{Evidence: []Evidence{{Source: "deltasignal-atlas-7-mcp"}}}}) != "live-mcp" {
		t.Fatal("live evidence not detected")
	}
	if toolMode(nil) != "deterministic-demo" {
		t.Fatal("empty findings should be deterministic")
	}
}

func TestCoordinatorBuildBriefCarriesEvidenceFidelity(t *testing.T) {
	stale := false
	c := Coordinator{
		Tools: staticToolClient{result: SpecialistResult{
			Agent:      "evidence-retriever",
			Summary:    "live result",
			Confidence: "live-mcp",
			Evidence: []Evidence{{
				Source:         "deltasignal-atlas-7-mcp",
				Title:          "MCP tool: deltasignal_company_report",
				Observation:    "live result",
				SourceDate:     "2026-06-07",
				ComputedAt:     "2026-06-08T17:37:01Z",
				Stale:          &stale,
				Caveats:        []string{"none"},
				QualityFlags:   []string{"fresh"},
				EvidenceHashes: []string{"hash1"},
				PayloadMode:    "compact",
			}},
		}},
	}
	resp, err := c.BuildBrief(context.Background(), BriefRequest{Issuer: "HUT"})
	if err != nil {
		t.Fatalf("BuildBrief returned error: %v", err)
	}
	if resp.EvidenceFidelity.Status != "provenance-carried" || len(resp.EvidenceFidelity.SourceDates) != 1 || len(resp.EvidenceFidelity.EvidenceHashes) != 1 {
		t.Fatalf("evidence fidelity not carried: %#v", resp.EvidenceFidelity)
	}
	if !strings.Contains(strings.Join(resp.Disclosures, "\n"), "Evidence fidelity: provenance-carried") {
		t.Fatalf("disclosures did not include evidence fidelity: %#v", resp.Disclosures)
	}
}

type failingAgentTool struct{}

func (failingAgentTool) StressSignals(context.Context, string) (SpecialistResult, error) {
	return SpecialistResult{}, errors.New("stress failed")
}

func (failingAgentTool) CompanyEvidence(context.Context, string) (SpecialistResult, error) {
	return SpecialistResult{}, errors.New("company failed")
}

func (failingAgentTool) PeerContext(context.Context, string) (SpecialistResult, error) {
	return SpecialistResult{}, errors.New("peer failed")
}

type staticToolClient struct {
	result SpecialistResult
}

func (s staticToolClient) StressSignals(context.Context, string) (SpecialistResult, error) {
	return s.result, nil
}

func (s staticToolClient) CompanyEvidence(context.Context, string) (SpecialistResult, error) {
	return s.result, nil
}

func (s staticToolClient) PeerContext(context.Context, string) (SpecialistResult, error) {
	return s.result, nil
}
