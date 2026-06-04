package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aitrailblazer/deltasignal-ai-agent/internal/agent"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	coordinator := agent.Coordinator{
		Tools: DemoToolClientFromEnv(),
	}
	if strings.EqualFold(os.Getenv("DELTASIGNAL_USE_GEMINI"), "true") {
		coordinator.Synthesizer = agent.GeminiSynthesizer{Model: os.Getenv("GEMINI_MODEL")}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"ok":      true,
			"service": "deltasignal-ai-agent",
			"time":    time.Now().UTC(),
		})
	})
	mux.HandleFunc("POST /v1/brief", func(w http.ResponseWriter, r *http.Request) {
		var req agent.BriefRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON request"})
			return
		}
		resp, err := coordinator.BuildBrief(r.Context(), req)
		if err != nil {
			logger.Error("build brief failed", "error", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to build brief"})
			return
		}
		writeJSON(w, http.StatusOK, resp)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	logger.Info("starting DeltaSignal AI Agent", "port", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func DemoToolClientFromEnv() agent.ToolClient {
	return agent.DemoToolClient{}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
