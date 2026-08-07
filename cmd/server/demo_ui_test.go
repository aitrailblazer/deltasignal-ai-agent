package main

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aitrailblazer/deltasignal-ai-agent/internal/agent"
)

func TestDemoUIRoute(t *testing.T) {
	mux := newMux(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		agent.Coordinator{Tools: agent.DemoToolClient{}},
		fakeTripCodeResolver{},
		agent.NewTripCodeMemoryStore(2),
		fakeTripCodeSynthesizer{text: "summary"},
		agent.NewCostTracker(agent.CostTrackerConfig{}),
		nil,
	)

	for _, path := range []string{"/demo/run", "/demo/run/"} {
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, path, nil))
		body := rr.Body.String()
		if rr.Code != http.StatusOK {
			t.Fatalf("%s code = %d", path, rr.Code)
		}
		if got := rr.Header().Get("Content-Type"); !strings.Contains(got, "text/html") {
			t.Fatalf("%s content type = %q", path, got)
		}
		for _, want := range []string{
			"Run The HUT Research-Memory Proof",
			"© 2026 AITrailblazer · DeltaSignal",
			"proofTitle",
			"stageReveal",
			"proofTitleIn",
			"stageIn",
			"DeltaSignal Evidence OS · Research Memory Proof",
			"Second proof after the flagship Apple workflow",
			`href="/demo"`,
			"Overview",
			"btnIcon",
			"showOverview",
			"What to expect",
			"Substack starts the loop",
			"ATLAS-7 is the evidence engine",
			"Gemini orchestrates synthesis",
			"captionText",
			"revealCaption",
			"On screen: the left panel contains the controls",
			"On screen: the response window shows the top of the published Hut 8 Substack article",
			"On screen: the blue response workspace turns into a live request trace",
			"On screen: the JSON response lists the Agent Card fields",
			"On screen: the trace shows a protected request in flight",
			"On screen: the response is a JSON-RPC task result",
			"On screen: the monitor response is the follow-up state",
			"On screen: the usage response shows request accounting",
			"Article Start",
			"Objective",
			"Show a real HUT research workflow",
			"showArticle",
			"actionCard",
			"actionText",
			"describeAction",
			"playTimedDemo",
			"autoplay",
			"revealActionText",
			"What happens next",
			"Health Check",
			"Agent Discovery",
			"/assets/substack-hut-article-top.png",
			"https://deltasignal.substack.com/p/hut-8-the-re-rating-has-a-deadline",
			"LitElement",
			"customElements.define('demo-app'",
			"connectedCallback()",
			"bindControls()",
			"colorizeJSON",
			"json-key",
			"json-brace",
			"json-colon",
			"Live request trace",
			"measured network time",
			"data-scroll=\"page\"",
			"scrollResponse",
			"Render",
			"data-render=\"rich\"",
			"data-render=\"json\"",
			"renderResponse",
			"richRender",
			"Rendered evidence page",
			"/assets/deltasignal-app-icon.png",
			"X-Demo-Key",
			"TF-SUB-9DA70A7F98",
			"/.well-known/agent-card.json",
			"/resolve?tripcode=",
			"monitor_tripcode_thesis",
			"/v1/usage",
			".headerInputGrid{width:100%;grid-template-columns:minmax(0,1fr)}",
			"demo-app{display:block;width:100%;max-width:100vw;height:auto;min-height:100dvh;overflow:visible}",
		} {
			if !strings.Contains(body, want) {
				t.Fatalf("%s body missing %q", path, want)
			}
		}
		if strings.Contains(strings.ToLower(body), "investment advice") {
			t.Fatalf("%s demo UI should not invite investment-advice framing", path)
		}
	}
}

func TestDemoLandingRoute(t *testing.T) {
	mux := newMux(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		agent.Coordinator{Tools: agent.DemoToolClient{}},
		fakeTripCodeResolver{},
		agent.NewTripCodeMemoryStore(2),
		fakeTripCodeSynthesizer{text: "summary"},
		agent.NewCostTracker(agent.CostTrackerConfig{}),
		nil,
	)

	for _, path := range []string{"/demo", "/demo/"} {
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, path, nil))
		body := rr.Body.String()
		if rr.Code != http.StatusOK {
			t.Fatalf("%s code = %d", path, rr.Code)
		}
		if got := rr.Header().Get("Content-Type"); !strings.Contains(got, "text/html") {
			t.Fatalf("%s content type = %q", path, got)
		}
		for _, want := range []string{
			"DeltaSignal · HUT Research Memory Proof",
			"© 2026 AITrailblazer · DeltaSignal",
			"A published thesis becomes an inspectable agent workflow.",
			"Apple is the flagship proof. HUT demonstrates reusable memory across time.",
			"Go To HUT Proof",
			"DEMO · 2 Min Sequence",
			`aria-label="Go to HUT proof"`,
			`aria-label="Start automated two minute demo"`,
			`href="/demo/run"`,
			`href="/demo/run?autoplay=1"`,
			"architectureStage",
			"archIcon",
			"archWire",
			"User",
			"intent + curl",
			"Cloud Run",
			"A2A agent",
			"MCP",
			"ATLAS-7",
			"HUT River",
			"TripCode memory",
			"Evidence",
			"SEC/XBRL refs",
			"TF-SUB-9DA70A7F98",
		} {
			if !strings.Contains(body, want) {
				t.Fatalf("%s body missing %q", path, want)
			}
		}
	}
}

func TestRootLandingRoute(t *testing.T) {
	mux := newMux(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		agent.Coordinator{Tools: agent.DemoToolClient{}},
		fakeTripCodeResolver{},
		agent.NewTripCodeMemoryStore(2),
		fakeTripCodeSynthesizer{text: "summary"},
		agent.NewCostTracker(agent.CostTrackerConfig{}),
		nil,
	)

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	body := rr.Body.String()
	if rr.Code != http.StatusOK {
		t.Fatalf("root code = %d", rr.Code)
	}
	if got := rr.Header().Get("Content-Type"); !strings.Contains(got, "text/html") {
		t.Fatalf("root content type = %q", got)
	}
	for _, want := range []string{
		"DeltaSignal Evidence OS · Apple Reference First",
		"Evidence agents can inspect—not just repeat.",
		"Open Apple Reference",
		`href="https://aitrailblazer.github.io/deltasignal-ai-agent/client-demo/google-cloud-aapl/"`,
		"Explore HUT Memory Proof",
		`href="/demo"`,
		"SEC/XBRL · ATLAS-7 · MCP · Google Cloud Agents",
		"Source · Apple SEC/XBRL",
		"Deliver · Evidence Packet",
		`meta name="description"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("root body missing %q", want)
		}
	}
}

func TestDemoUILogoRoute(t *testing.T) {
	mux := newMux(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		agent.Coordinator{Tools: agent.DemoToolClient{}},
		fakeTripCodeResolver{},
		agent.NewTripCodeMemoryStore(2),
		fakeTripCodeSynthesizer{text: "summary"},
		agent.NewCostTracker(agent.CostTrackerConfig{}),
		nil,
	)

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/assets/deltasignal-app-icon.png", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("logo route code = %d", rr.Code)
	}
	if got := rr.Header().Get("Content-Type"); !strings.Contains(got, "image/png") {
		t.Fatalf("logo content type = %q", got)
	}
	if rr.Body.Len() == 0 {
		t.Fatal("logo route returned empty body")
	}
}

func TestDemoUISubstackArticleImageRoute(t *testing.T) {
	mux := newMux(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		agent.Coordinator{Tools: agent.DemoToolClient{}},
		fakeTripCodeResolver{},
		agent.NewTripCodeMemoryStore(2),
		fakeTripCodeSynthesizer{text: "summary"},
		agent.NewCostTracker(agent.CostTrackerConfig{}),
		nil,
	)

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/assets/substack-hut-article-top.png", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("article image route code = %d", rr.Code)
	}
	if got := rr.Header().Get("Content-Type"); !strings.Contains(got, "image/png") {
		t.Fatalf("article image content type = %q", got)
	}
	if rr.Body.Len() == 0 {
		t.Fatal("article image route returned empty body")
	}
}

func TestDemoUIPipelineImageRoute(t *testing.T) {
	mux := newMux(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		agent.Coordinator{Tools: agent.DemoToolClient{}},
		fakeTripCodeResolver{},
		agent.NewTripCodeMemoryStore(2),
		fakeTripCodeSynthesizer{text: "summary"},
		agent.NewCostTracker(agent.CostTrackerConfig{}),
		nil,
	)

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/assets/research-to-logic-pipeline.png", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("pipeline image route code = %d", rr.Code)
	}
	if got := rr.Header().Get("Content-Type"); !strings.Contains(got, "image/png") {
		t.Fatalf("pipeline image content type = %q", got)
	}
	if rr.Body.Len() == 0 {
		t.Fatal("pipeline image route returned empty body")
	}
}
