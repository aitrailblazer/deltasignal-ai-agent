package agent

import (
	"context"
	"strings"
	"testing"
	"time"
)

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
}
