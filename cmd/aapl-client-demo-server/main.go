package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aitrailblazer/deltasignal-ai-agent/internal/aapldemo"
)

func main() {
	workflow, mode, err := buildWorkflow()
	if err != nil {
		log.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "mode": mode})
	})
	mux.HandleFunc("POST /v1/aapl-evidence-demo", func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Question string `json:"question"`
			AsOfDate string `json:"as_of_date"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10)).Decode(&input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON request"})
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
		defer cancel()
		result, err := workflow.Run(ctx, aapldemo.Request{
			Ticker:   "AAPL",
			Question: input.Question,
			AsOfDate: input.AsOfDate,
			Mode:     mode,
		})
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, result)
	})

	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8080"
	}
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           securityHeaders(mux),
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	log.Printf("AAPL client demo listening on :%s mode=%s", port, mode)
	log.Fatal(server.ListenAndServe())
}

func buildWorkflow() (aapldemo.Workflow, string, error) {
	mode := strings.ToLower(strings.TrimSpace(os.Getenv("AAPL_DEMO_MODE")))
	if mode == "" {
		mode = "fixture"
	}
	var caller aapldemo.ToolCaller
	var now func() time.Time
	switch mode {
	case "fixture":
		fixture, err := aapldemo.NewFixtureClient()
		if err != nil {
			return aapldemo.Workflow{}, "", err
		}
		caller = fixture
		now = func() time.Time { return aapldemo.FixtureGeneratedAt }
	case "live":
		caller = aapldemo.MCPClient{
			Endpoint: os.Getenv("DELTASIGNAL_MCP_ENDPOINT"),
			APIKey:   os.Getenv("DELTASIGNAL_MCP_API_KEY"),
		}
	default:
		return aapldemo.Workflow{}, "", fmt.Errorf("unsupported AAPL_DEMO_MODE %q", mode)
	}
	var synth aapldemo.Synthesizer = aapldemo.DeterministicSynthesizer{}
	if strings.EqualFold(strings.TrimSpace(os.Getenv("AAPL_DEMO_SYNTHESIS")), "gemini") {
		synth = aapldemo.GeminiSynthesizer{Model: os.Getenv("GEMINI_MODEL")}
	}
	return aapldemo.Workflow{Caller: caller, Synthesizer: synth, Now: now}, mode, nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}
