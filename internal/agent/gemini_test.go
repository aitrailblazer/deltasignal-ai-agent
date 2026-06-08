package agent

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestGeminiSynthesizerSynthesizeUsesInjectedGenerator(t *testing.T) {
	var gotModel string
	var gotPrompt string
	s := GeminiSynthesizer{
		Generate: func(_ context.Context, model string, prompt string) (string, error) {
			gotModel = model
			gotPrompt = prompt
			return "brief", nil
		},
	}
	text, err := s.Synthesize(context.Background(), BriefRequest{Issuer: "HUT"}, []string{"plan"}, []SpecialistResult{{Agent: "test", Evidence: []Evidence{{PayloadMode: "compact", SourceDate: "2026-06-07"}}}})
	if err != nil {
		t.Fatalf("Synthesize returned error: %v", err)
	}
	if text != "brief" || gotModel != "gemini-2.5-flash" || !strings.Contains(gotPrompt, "DeltaSignal Gemini AI Agent") || !strings.Contains(gotPrompt, `"issuer": "HUT"`) || !strings.Contains(gotPrompt, "evidence_fidelity") || !strings.Contains(gotPrompt, "source dates") {
		t.Fatalf("unexpected synthesis: text=%q model=%q prompt=%s", text, gotModel, gotPrompt)
	}
}

func TestGeminiSynthesizerSynthesizeTripCodeUsesInjectedGenerator(t *testing.T) {
	var gotModel string
	var gotPrompt string
	s := GeminiSynthesizer{
		Model: "gemini-test",
		Generate: func(_ context.Context, model string, prompt string) (string, error) {
			gotModel = model
			gotPrompt = prompt
			return "trip summary", nil
		},
	}
	text, err := s.SynthesizeTripCode(context.Background(), TripCodeResearchRequest{TripCode: "TF-SUB-X"}, TripCodeResearchResponse{
		TripCode: "TF-SUB-X",
		Packet:   map[string]any{"status": "ready"},
		AgentContext: &AgentContextSnapshot{
			Enabled: true,
			Sources: []AgentContextSource{{
				URL:    "https://aitrailblazer.github.io/deltasignal-atlas-codex-plugin/CLAUDE.md",
				Status: "fetched",
				SHA256: "abc123",
			}},
		},
		ExecutionTrace: []ExecutionTraceStep{{Order: 1, Actor: "google-agent", Action: "loaded context"}},
		Disclosures:    DefaultTripCodeDisclosures(),
	})
	if err != nil {
		t.Fatalf("SynthesizeTripCode returned error: %v", err)
	}
	if text != "trip summary" || gotModel != "gemini-test" || !strings.Contains(gotPrompt, "TF-SUB-X") || !strings.Contains(gotPrompt, "Preserve all boundaries") || !strings.Contains(gotPrompt, "agent_context") || !strings.Contains(gotPrompt, "execution_trace") {
		t.Fatalf("unexpected TripCode synthesis: text=%q model=%q prompt=%s", text, gotModel, gotPrompt)
	}
}

func TestGeminiSynthesizerInjectedGeneratorError(t *testing.T) {
	s := GeminiSynthesizer{
		Generate: func(context.Context, string, string) (string, error) {
			return "", errors.New("boom")
		},
	}
	if _, err := s.Synthesize(context.Background(), BriefRequest{}, nil, nil); err == nil || !strings.Contains(err.Error(), "generate Gemini brief") {
		t.Fatalf("expected Synthesize error, got %v", err)
	}
	if _, err := s.SynthesizeTripCode(context.Background(), TripCodeResearchRequest{}, TripCodeResearchResponse{}); err == nil || !strings.Contains(err.Error(), "generate Gemini TripCode summary") {
		t.Fatalf("expected SynthesizeTripCode error, got %v", err)
	}
}

func TestGeminiSynthesizerSynthesizeTripCodeMarshalError(t *testing.T) {
	s := GeminiSynthesizer{Generate: func(context.Context, string, string) (string, error) {
		t.Fatal("generator should not be called")
		return "", nil
	}}
	_, err := s.SynthesizeTripCode(context.Background(), TripCodeResearchRequest{}, TripCodeResearchResponse{
		Packet: map[string]any{"bad": func() {}},
	})
	if err == nil || !strings.Contains(err.Error(), "marshal TripCode synthesis context") {
		t.Fatalf("expected marshal error, got %v", err)
	}
}

func TestGeminiSynthesizerUsesDefaultGenerator(t *testing.T) {
	oldGenerate := generateGeminiContent
	t.Cleanup(func() { generateGeminiContent = oldGenerate })
	var gotModel string
	var gotPrompt string
	var gotOperation string
	generateGeminiContent = func(_ context.Context, model string, prompt string, operation string) (string, error) {
		gotModel = model
		gotPrompt = prompt
		gotOperation = operation
		return "default generated", nil
	}
	text, err := (GeminiSynthesizer{Model: "gemini-live"}).Synthesize(context.Background(), BriefRequest{Issuer: "HUT"}, nil, nil)
	if err != nil {
		t.Fatalf("Synthesize returned error: %v", err)
	}
	if text != "default generated" || gotModel != "gemini-live" || !strings.Contains(gotPrompt, "HUT") || gotOperation != "generate Gemini brief" {
		t.Fatalf("unexpected default generator call text=%q model=%q op=%q prompt=%s", text, gotModel, gotOperation, gotPrompt)
	}
}

func TestGeminiSynthesizerDefaultGeneratorError(t *testing.T) {
	oldGenerate := generateGeminiContent
	t.Cleanup(func() { generateGeminiContent = oldGenerate })
	generateGeminiContent = func(context.Context, string, string, string) (string, error) {
		return "", errors.New("default failed")
	}
	if _, err := (GeminiSynthesizer{}).Synthesize(context.Background(), BriefRequest{}, nil, nil); err == nil || !strings.Contains(err.Error(), "default failed") {
		t.Fatalf("expected default generator error, got %v", err)
	}
}
