package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aitrailblazer/deltasignal-ai-agent/internal/aapldemo"
)

func main() {
	mode := flag.String("mode", "fixture", "fixture or live")
	question := flag.String("question", "What can the filing evidence establish about Apple, and which ATLAS-7 levels remain unsupported?", "bounded AAPL evidence question")
	asOf := flag.String("as-of", "", "optional point-in-time cutoff (YYYY-MM-DD)")
	synthesis := flag.String("synthesis", "deterministic", "deterministic or gemini")
	flag.Parse()

	var caller aapldemo.ToolCaller
	var now func() time.Time
	switch strings.ToLower(strings.TrimSpace(*mode)) {
	case "fixture":
		fixture, err := aapldemo.NewFixtureClient()
		if err != nil {
			fatal(err)
		}
		caller = fixture
		now = func() time.Time { return aapldemo.FixtureGeneratedAt }
	case "live":
		caller = aapldemo.MCPClient{
			Endpoint: os.Getenv("DELTASIGNAL_MCP_ENDPOINT"),
			APIKey:   os.Getenv("DELTASIGNAL_MCP_API_KEY"),
		}
	default:
		fatal(fmt.Errorf("unsupported mode %q", *mode))
	}

	var synth aapldemo.Synthesizer = aapldemo.DeterministicSynthesizer{}
	if strings.EqualFold(strings.TrimSpace(*synthesis), "gemini") {
		synth = aapldemo.GeminiSynthesizer{Model: os.Getenv("GEMINI_MODEL")}
	}
	workflow := aapldemo.Workflow{Caller: caller, Synthesizer: synth, Now: now}
	response, err := workflow.Run(context.Background(), aapldemo.Request{
		Ticker:   "AAPL",
		Question: *question,
		AsOfDate: *asOf,
		Mode:     strings.ToLower(strings.TrimSpace(*mode)),
	})
	if err != nil {
		fatal(err)
	}
	out, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		fatal(err)
	}
	fmt.Println(string(out))
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "aapl-client-demo:", err)
	os.Exit(1)
}
