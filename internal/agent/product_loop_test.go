package agent

import (
	"strings"
	"testing"
	"time"
)

func TestBuildProductLoopDefaults(t *testing.T) {
	resp := BuildProductLoop(ProductLoopRequest{}, time.Time{})
	if resp.Objective == "" || resp.Status != "planned" || resp.WorkflowType != "daily_product_engineering_triage" || resp.EvidenceMode != "source_of_truth_first" || resp.RiskLevel != "medium" {
		t.Fatalf("defaults not applied: %#v", resp)
	}
	if !resp.HumanApprovalRequired || resp.Parallelism != "single-isolated-builder" {
		t.Fatalf("default controls unexpected: approval=%v parallelism=%s", resp.HumanApprovalRequired, resp.Parallelism)
	}
	if resp.GeneratedAt.IsZero() || resp.RootAgent.Name != "Product Loop Orchestrator" || len(resp.Agents) != 6 || len(resp.ToolClasses) != 4 || len(resp.MemoryLayers) != 3 || len(resp.GroundingLayers) != 3 || len(resp.DeploymentPhases) != 4 {
		t.Fatalf("loop shape incomplete: %#v", resp)
	}
	if len(resp.VerifierGates) == 0 || len(resp.StopConditions) == 0 || len(resp.Boundaries) == 0 || !strings.Contains(resp.FirstBuild, "Daily Product Engineering Triage Loop") {
		t.Fatalf("governance details missing: %#v", resp)
	}
}

func TestBuildProductLoopOverridesAndBranches(t *testing.T) {
	noApproval := false
	now := time.Date(2026, 6, 10, 12, 30, 0, 0, time.UTC)
	resp := BuildProductLoop(ProductLoopRequest{
		Objective:             "  Ship ADK loop safely  ",
		WorkflowType:          "TripCode Monitor",
		EvidenceMode:          "Memory-First",
		RiskLevel:             "low",
		RequireHumanApproval:  &noApproval,
		AllowParallelBuilders: true,
	}, now)
	if resp.Objective != "Ship ADK loop safely" || resp.WorkflowType != "tripcode_monitor" || resp.EvidenceMode != "memory_first" || resp.RiskLevel != "low" {
		t.Fatalf("overrides not normalized: %#v", resp)
	}
	if resp.HumanApprovalRequired || resp.Parallelism != "parallel-builders-with-one-worktree-per-task" || !resp.GeneratedAt.Equal(now) {
		t.Fatalf("override controls unexpected: %#v", resp)
	}

	security := BuildProductLoop(ProductLoopRequest{RiskLevel: "Security", RequireHumanApproval: &noApproval}, now)
	if !security.HumanApprovalRequired || security.RiskLevel != "security" {
		t.Fatalf("security risk did not force approval: %#v", security)
	}
	high := BuildProductLoop(ProductLoopRequest{RiskLevel: "high", RequireHumanApproval: &noApproval}, now)
	if !high.HumanApprovalRequired {
		t.Fatalf("high risk did not force approval: %#v", high)
	}
}

func TestNormalizedProductLoopValue(t *testing.T) {
	if got := normalizedProductLoopValue("", "fallback"); got != "fallback" {
		t.Fatalf("empty normalized value = %q", got)
	}
	if got := normalizedProductLoopValue("  A-B C  ", "fallback"); got != "a_b_c" {
		t.Fatalf("normalized value = %q", got)
	}
}
