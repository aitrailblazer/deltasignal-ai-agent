package agent

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"google.golang.org/genai"
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

func TestGenerateGeminiContentUsesClientFactory(t *testing.T) {
	oldNew := newGeminiClient
	t.Cleanup(func() { newGeminiClient = oldNew })
	newGeminiClient = func(context.Context) (geminiContentClient, error) {
		return fakeGeminiClient{text: "live text"}, nil
	}
	text, err := generateGeminiContent(context.Background(), "model", "prompt", "op")
	if err != nil || text != "live text" {
		t.Fatalf("generateGeminiContent text=%q err=%v", text, err)
	}
	newGeminiClient = func(context.Context) (geminiContentClient, error) {
		return nil, errors.New("client failed")
	}
	if _, err := generateGeminiContent(context.Background(), "model", "prompt", "op"); err == nil || !strings.Contains(err.Error(), "client failed") {
		t.Fatalf("expected client error, got %v", err)
	}
	newGeminiClient = func(context.Context) (geminiContentClient, error) {
		return fakeGeminiClient{err: errors.New("generate failed")}, nil
	}
	if _, err := generateGeminiContent(context.Background(), "model", "prompt", "op"); err == nil || !strings.Contains(err.Error(), "op") {
		t.Fatalf("expected operation error, got %v", err)
	}
}

func TestGoogleGenAIClientAndFactory(t *testing.T) {
	oldGenerate := googleGenerateContent
	oldNew := genaiNewClient
	t.Cleanup(func() {
		googleGenerateContent = oldGenerate
		genaiNewClient = oldNew
	})
	var gotModel string
	var gotPrompt string
	googleGenerateContent = func(_ context.Context, _ *genai.Client, model string, prompt string) (string, error) {
		gotModel = model
		gotPrompt = prompt
		return "google text", nil
	}
	text, err := (googleGenAIClient{client: &genai.Client{}}).GenerateContentText(context.Background(), "gemini", "hello")
	if err != nil || text != "google text" || gotModel != "gemini" || gotPrompt != "hello" {
		t.Fatalf("google client text=%q model=%q prompt=%q err=%v", text, gotModel, gotPrompt, err)
	}
	googleGenerateContent = func(context.Context, *genai.Client, string, string) (string, error) {
		return "", errors.New("google failed")
	}
	if _, err := (googleGenAIClient{client: &genai.Client{}}).GenerateContentText(context.Background(), "gemini", "hello"); err == nil {
		t.Fatal("expected google generate error")
	}

	genaiNewClient = func(context.Context, *genai.ClientConfig) (*genai.Client, error) {
		return &genai.Client{}, nil
	}
	if client, err := newGeminiClient(context.Background()); err != nil || client == nil {
		t.Fatalf("newGeminiClient success client=%#v err=%v", client, err)
	}
	genaiNewClient = func(context.Context, *genai.ClientConfig) (*genai.Client, error) {
		return nil, errors.New("new failed")
	}
	if _, err := newGeminiClient(context.Background()); err == nil || !strings.Contains(err.Error(), "create Gemini client") {
		t.Fatalf("expected newGeminiClient error, got %v", err)
	}
}

func TestDefaultGoogleGenerateContent(t *testing.T) {
	t.Setenv("GOOGLE_API_KEY", "test-api-key")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "generateContent") {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"candidates":[{"content":{"parts":[{"text":"server text"}],"role":"model"}}]}`))
	}))
	defer server.Close()
	client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		HTTPOptions: genai.HTTPOptions{BaseURL: server.URL, APIVersion: "v1"},
	})
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	text, err := googleGenerateContent(context.Background(), client, "gemini-test", "prompt")
	if err != nil || text != "server text" {
		t.Fatalf("googleGenerateContent text=%q err=%v", text, err)
	}

	errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":{"message":"server failed"}}`))
	}))
	defer errorServer.Close()
	errorClient, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		HTTPOptions: genai.HTTPOptions{BaseURL: errorServer.URL, APIVersion: "v1"},
	})
	if err != nil {
		t.Fatalf("NewClient error client returned error: %v", err)
	}
	if _, err := googleGenerateContent(context.Background(), errorClient, "gemini-test", "prompt"); err == nil {
		t.Fatal("expected googleGenerateContent server error")
	}
}

type fakeGeminiClient struct {
	text string
	err  error
}

func (c fakeGeminiClient) GenerateContentText(context.Context, string, string) (string, error) {
	return c.text, c.err
}
