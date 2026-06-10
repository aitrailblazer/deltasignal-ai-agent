package agent

import (
	"strings"
	"time"
)

type ProductLoopRequest struct {
	Objective             string `json:"objective"`
	WorkflowType          string `json:"workflow_type,omitempty"`
	EvidenceMode          string `json:"evidence_mode,omitempty"`
	RiskLevel             string `json:"risk_level,omitempty"`
	RequireHumanApproval  *bool  `json:"require_human_approval,omitempty"`
	AllowParallelBuilders bool   `json:"allow_parallel_builders,omitempty"`
}

type ProductLoopResponse struct {
	Objective             string                 `json:"objective"`
	GeneratedAt           time.Time              `json:"generated_at"`
	Status                string                 `json:"status"`
	Track                 string                 `json:"track"`
	ADKRole               string                 `json:"adk_role"`
	WorkflowType          string                 `json:"workflow_type"`
	EvidenceMode          string                 `json:"evidence_mode"`
	RiskLevel             string                 `json:"risk_level"`
	HumanApprovalRequired bool                   `json:"human_approval_required"`
	Parallelism           string                 `json:"parallelism"`
	RootAgent             ProductLoopAgent       `json:"root_agent"`
	Agents                []ProductLoopAgent     `json:"agents"`
	ToolClasses           []ProductLoopToolClass `json:"tool_classes"`
	MemoryLayers          []ProductLoopLayer     `json:"memory_layers"`
	GroundingLayers       []ProductLoopLayer     `json:"grounding_layers"`
	DeploymentPhases      []ProductLoopPhase     `json:"deployment_phases"`
	VerifierGates         []string               `json:"verifier_gates"`
	StopConditions        []string               `json:"stop_conditions"`
	FirstBuild            string                 `json:"first_build"`
	Boundaries            []string               `json:"boundaries"`
	Cost                  *CostSnapshot          `json:"cost,omitempty"`
	Runtime               *RuntimeTelemetry      `json:"runtime,omitempty"`
}

type ProductLoopAgent struct {
	Name             string   `json:"name"`
	Role             string   `json:"role"`
	Responsibilities []string `json:"responsibilities"`
}

type ProductLoopToolClass struct {
	Name  string   `json:"name"`
	Tools []string `json:"tools"`
}

type ProductLoopLayer struct {
	Name   string   `json:"name"`
	Use    string   `json:"use"`
	Stores []string `json:"stores,omitempty"`
}

type ProductLoopPhase struct {
	Name            string   `json:"name"`
	Goal            string   `json:"goal"`
	Controls        []string `json:"controls"`
	SuccessCriteria []string `json:"success_criteria,omitempty"`
}

func BuildProductLoop(req ProductLoopRequest, now time.Time) ProductLoopResponse {
	objective := strings.TrimSpace(req.Objective)
	if objective == "" {
		objective = "Run the DeltaSignal ADK product loop as a governed issuer-diligence workflow."
	}
	workflowType := normalizedProductLoopValue(req.WorkflowType, "daily_product_engineering_triage")
	evidenceMode := normalizedProductLoopValue(req.EvidenceMode, "source_of_truth_first")
	riskLevel := normalizedProductLoopValue(req.RiskLevel, "medium")
	humanApproval := true
	if req.RequireHumanApproval != nil {
		humanApproval = *req.RequireHumanApproval
	}
	if strings.EqualFold(riskLevel, "high") || strings.EqualFold(riskLevel, "security") {
		humanApproval = true
	}
	parallelism := "single-isolated-builder"
	if req.AllowParallelBuilders {
		parallelism = "parallel-builders-with-one-worktree-per-task"
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}

	return ProductLoopResponse{
		Objective:             objective,
		GeneratedAt:           now.UTC(),
		Status:                "planned",
		Track:                 "Track 2 Optimize -> Track 3 Path",
		ADKRole:               "ADK orchestrates the loop; Go Cloud Run and ATLAS-7 MCP remain bounded evidence/tool runtimes.",
		WorkflowType:          workflowType,
		EvidenceMode:          evidenceMode,
		RiskLevel:             riskLevel,
		HumanApprovalRequired: humanApproval,
		Parallelism:           parallelism,
		RootAgent: ProductLoopAgent{
			Name: "Product Loop Orchestrator",
			Role: "Own loop state, route specialists, enforce stop conditions, and escalate high-risk work.",
			Responsibilities: []string{
				"read durable loop memory",
				"trigger discovery",
				"select workflow agents",
				"enforce verifier-before-publisher rule",
				"request human approval when risk or uncertainty is high",
			},
		},
		Agents: []ProductLoopAgent{
			{Name: "Discovery Agent", Role: "Find work before a human prompt.", Responsibilities: []string{"scan CI, issues, logs, feedback, TripCode monitor queues, issuer Rivers, and ATLAS-7 evidence changes"}},
			{Name: "Triage Agent", Role: "Rank and deduplicate findings.", Responsibilities: []string{"score customer impact, revenue impact, security risk, regression risk, evidence strength, and effort"}},
			{Name: "Planner Agent", Role: "Turn findings into constrained execution specs.", Responsibilities: []string{"write goal, evidence, affected systems, acceptance criteria, required tools, risk level, verifier checklist, and approval requirement"}},
			{Name: "Builder Agent", Role: "Implement one scoped task in isolation.", Responsibilities: []string{"use one branch or worktree per task", "avoid merge, deploy, and publish authority"}},
			{Name: "Verifier Agent", Role: "Keep maker separate from checker.", Responsibilities: []string{"run tests, lint, typecheck, diff review, policy checks, evidence binding checks, secret scan, and output evaluation"}},
			{Name: "Publisher Agent", Role: "Publish only after verifier pass.", Responsibilities: []string{"open PRs, update tickets, write changelog, create artifacts, post summaries, and mark memory complete or blocked"}},
		},
		ToolClasses: []ProductLoopToolClass{
			{Name: "Repo And Build Tools", Tools: []string{"read_repo", "write_patch", "create_worktree", "run_tests", "run_lint", "run_typecheck", "inspect_diff", "open_pr"}},
			{Name: "Product Ops Tools", Tools: []string{"read_issues", "update_ticket", "read_ci", "post_summary", "create_release_note", "query_customer_feedback"}},
			{Name: "Grounding Tools", Tools: []string{"search_docs", "search_memory", "retrieve_artifact", "query_database_readonly", "search_web_with_grounding"}},
			{Name: "Governance Tools", Tools: []string{"policy_check", "secret_scan", "license_check", "eval_trace", "human_approval_request"}},
		},
		MemoryLayers: []ProductLoopLayer{
			{Name: "Session State", Use: "current run state", Stores: []string{"current task", "active agent", "tool calls", "pending approvals", "temporary artifacts", "acceptance criteria"}},
			{Name: "Durable Loop Memory", Use: "cross-run continuity", Stores: []string{"loop_runs", "findings", "tasks", "agent_actions", "verification_records", "human_decisions", "artifact_refs", "blocked_items", "known_project_conventions"}},
			{Name: "Product Knowledge Memory", Use: "versioned policy and skill retrieval", Stores: []string{"architecture", "code style", "release policy", "security policy", "renderer contract", "DeltaSignal report rules", "MCP tool policy"}},
		},
		GroundingLayers: []ProductLoopLayer{
			{Name: "Source Of Truth Grounding", Use: "default grounding", Stores: []string{"repo", "tests", "DB schema", "API contracts", "design docs", "product specs", "tickets", "prior verification records"}},
			{Name: "Memory Grounding", Use: "continuity grounding", Stores: []string{"prior loop attempts", "known failed fixes", "recurring regressions", "reviewer decisions", "customer-impact history"}},
			{Name: "External Grounding", Use: "current facts outside controlled corpus only", Stores: []string{"dependency docs", "vendor APIs", "security advisories", "platform documentation", "current web facts"}},
		},
		DeploymentPhases: []ProductLoopPhase{
			{Name: "Phase 1 Local Dev Harness", Goal: "prove loop behavior without autonomy risk", Controls: []string{"manual trigger only", "isolated temp branches", "no direct PR creation", "no deploy authority"}, SuccessCriteria: []string{"useful triage", "constrained diffs", "verifier catches bad work", "accurate memory records"}},
			{Name: "Phase 2 Cloud Run Background Worker", Goal: "scheduled loop execution", Controls: []string{"scheduled trigger", "durable DB memory", "artifact storage", "task-scoped connectors", "human approval before merge/deploy/publication"}},
			{Name: "Phase 3 Parallel Agent Execution", Goal: "scale safely", Controls: []string{"task queue", "one worker per task", "branch/worktree isolation", "concurrency limits", "cost budgets", "review queue"}},
			{Name: "Phase 4 Production Agent Platform", Goal: "durable observable governed loops", Controls: []string{"full tracing", "eval dashboards", "approval workflows", "rollback hooks", "audit logs", "cost monitoring", "model/tool version pinning", "golden task regression suite"}},
		},
		VerifierGates: []string{
			"tests pass",
			"lint and typecheck pass",
			"diff matches acceptance criteria",
			"no unrelated churn",
			"no secret leakage",
			"evidence and provenance preserved",
			"non-advice boundary preserved",
			"human approval present when required",
		},
		StopConditions: []string{
			"missing required evidence",
			"tool unavailable",
			"budget exhausted",
			"verifier failed twice",
			"human approval required",
			"risk exceeds task credential scope",
		},
		FirstBuild: "Daily Product Engineering Triage Loop: discover CI/issues/logs/tickets, rank candidates, plan one constrained low-risk task, implement in isolation, verify, open PR, record memory, and require human merge.",
		Boundaries: []string{
			"ADK is the orchestration layer, not a substitute for ATLAS-7 evidence.",
			"Builder cannot merge, deploy, publish, or self-verify.",
			"Memory is context, not source evidence.",
			"External grounding is used only when facts are outside the controlled product corpus.",
		},
	}
}

func normalizedProductLoopValue(value, fallback string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return fallback
	}
	value = strings.NewReplacer(" ", "_", "-", "_").Replace(value)
	return value
}
