package aapldemo

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"
	"time"
)

type recordingCaller struct {
	mu     sync.Mutex
	tools  []string
	result json.RawMessage
	errs   map[string]error
}

func (c *recordingCaller) CallTool(_ context.Context, tool string, _ map[string]any) (json.RawMessage, int, error) {
	c.mu.Lock()
	c.tools = append(c.tools, tool)
	c.mu.Unlock()
	if err := c.errs[tool]; err != nil {
		return nil, http.StatusPaymentRequired, err
	}
	return append(json.RawMessage(nil), c.result...), http.StatusOK, nil
}

func TestEvidencePlanUsesExactlyFourVaultSurfaces(t *testing.T) {
	plan := EvidencePlan(Request{Ticker: "AAPL", AsOfDate: "2026-03-28"})
	got := make([]string, 0, len(plan))
	for _, spec := range plan {
		got = append(got, spec.Tool)
	}
	want := []string{
		"deltasignal_company_fundamentals",
		"deltasignal_atlas7_companyfacts_history",
		"deltasignal_atlas7_point_in_time_history",
		"deltasignal_atlas7_four_level_applicability",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("tools = %v, want %v", got, want)
	}
	if got := plan[2].Arguments["as_of_date_to"]; got != "2026-03-28" {
		t.Fatalf("point-in-time cutoff = %v", got)
	}
}

func TestWorkflowPreservesStableAgentOrderAndPartialAccess(t *testing.T) {
	caller := &recordingCaller{
		result: json.RawMessage(`{"content":[{"type":"text","text":"ok"}]}`),
		errs: map[string]error{
			"deltasignal_atlas7_point_in_time_history": AccessRequiredError{StatusCode: 402, Message: "payment required"},
		},
	}
	fixed := time.Date(2026, 8, 6, 20, 0, 0, 0, time.UTC)
	response, err := (Workflow{
		Caller: caller,
		Now:    func() time.Time { return fixed },
	}).Run(context.Background(), Request{Ticker: "aapl", Mode: "live"})
	if err != nil {
		t.Fatal(err)
	}
	if !response.Partial || !response.AccessNeeded {
		t.Fatalf("partial=%v access=%v", response.Partial, response.AccessNeeded)
	}
	if response.GeneratedAt != fixed {
		t.Fatalf("generated_at = %v", response.GeneratedAt)
	}
	for i, spec := range EvidencePlan(response.Request) {
		if response.Agents[i].Tool != spec.Tool {
			t.Fatalf("agent %d tool=%s want=%s", i, response.Agents[i].Tool, spec.Tool)
		}
	}
	if response.Agents[2].Status != StatusAccessRequired {
		t.Fatalf("status = %s", response.Agents[2].Status)
	}
}

func TestFixtureWorkflowIsDeterministicAndComplete(t *testing.T) {
	fixture, err := NewFixtureClient()
	if err != nil {
		t.Fatal(err)
	}
	fixed := time.Date(2026, 8, 6, 20, 0, 0, 0, time.UTC)
	workflow := Workflow{Caller: fixture, Now: func() time.Time { return fixed }}
	first, err := workflow.Run(context.Background(), Request{Ticker: "AAPL", Mode: "fixture"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := workflow.Run(context.Background(), Request{Ticker: "AAPL", Mode: "fixture"})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatal("fixture responses differ")
	}
	if !first.Partial || first.AccessNeeded || len(first.Agents) != 4 {
		t.Fatalf("unexpected fixture response: partial=%v access=%v agents=%d", first.Partial, first.AccessNeeded, len(first.Agents))
	}
}

func TestMCPClientNeverRetriesOrPaysOn402(t *testing.T) {
	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if got := r.Header.Get("Authorization"); got != "" {
			t.Fatalf("unexpected authorization header: %q", got)
		}
		w.WriteHeader(http.StatusPaymentRequired)
		_, _ = w.Write([]byte(`{"error":"payment required"}`))
	}))
	defer server.Close()

	_, status, err := (MCPClient{Endpoint: server.URL}).CallTool(context.Background(), "example", map[string]any{})
	if status != http.StatusPaymentRequired || err == nil {
		t.Fatalf("status=%d err=%v", status, err)
	}
	var accessErr AccessRequiredError
	if !errors.As(err, &accessErr) {
		t.Fatalf("error type = %T", err)
	}
	if calls != 1 {
		t.Fatalf("calls=%d, want 1", calls)
	}
}

func TestWorkflowRejectsNonAAPL(t *testing.T) {
	_, err := (Workflow{Caller: &recordingCaller{}}).Run(context.Background(), Request{Ticker: "GOOGL"})
	if err == nil {
		t.Fatal("expected bounded-ticker error")
	}
}
