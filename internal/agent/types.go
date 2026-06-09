package agent

import "time"

type BriefRequest struct {
	Issuer   string `json:"issuer"`
	Question string `json:"question"`
	Live     bool   `json:"live,omitempty"`
}

type TripCodeResearchRequest struct {
	TripCode              string `json:"tripcode"`
	Issuer                string `json:"issuer,omitempty"`
	SessionID             string `json:"session_id,omitempty"`
	Question              string `json:"question,omitempty"`
	PayloadMode           string `json:"payload_mode,omitempty"`
	IncludeArticleBody    bool   `json:"include_article_body,omitempty"`
	IncludeFilingEvidence bool   `json:"include_filing_evidence,omitempty"`
	IncludePriorArticles  bool   `json:"include_prior_articles,omitempty"`
	IncludeThesisMap      bool   `json:"include_thesis_map,omitempty"`
	IncludeAgentContext   bool   `json:"include_agent_context,omitempty"`
}

type TripCodeResearchResponse struct {
	TripCode       string                  `json:"tripcode"`
	Issuer         string                  `json:"issuer,omitempty"`
	GeneratedAt    time.Time               `json:"generated_at"`
	Mode           string                  `json:"mode"`
	Packet         map[string]any          `json:"packet"`
	GeminiSummary  string                  `json:"gemini_summary,omitempty"`
	Disclosures    []string                `json:"disclosures"`
	Memory         *TripCodeMemorySnapshot `json:"memory,omitempty"`
	AgentContext   *AgentContextSnapshot   `json:"agent_context,omitempty"`
	ExecutionTrace []ExecutionTraceStep    `json:"execution_trace,omitempty"`
	Cost           *CostSnapshot           `json:"cost,omitempty"`
	Runtime        *RuntimeTelemetry       `json:"runtime,omitempty"`
}

type AgentContextSnapshot struct {
	Enabled bool                 `json:"enabled"`
	Purpose string               `json:"purpose,omitempty"`
	Sources []AgentContextSource `json:"sources"`
}

type AgentContextSource struct {
	URL        string `json:"url"`
	Status     string `json:"status"`
	HTTPStatus int    `json:"http_status,omitempty"`
	Bytes      int    `json:"bytes,omitempty"`
	SHA256     string `json:"sha256,omitempty"`
	Excerpt    string `json:"excerpt,omitempty"`
	Error      string `json:"error,omitempty"`
}

type ExecutionTraceStep struct {
	Order    int    `json:"order"`
	Actor    string `json:"actor"`
	Action   string `json:"action"`
	Evidence string `json:"evidence,omitempty"`
}

type TripCodeMemoryEntry struct {
	TripCode            string    `json:"tripcode"`
	Issuer              string    `json:"issuer,omitempty"`
	Mode                string    `json:"mode"`
	GeneratedAt         time.Time `json:"generated_at"`
	ArticleTitle        string    `json:"article_title,omitempty"`
	RiverNodeCount      int       `json:"river_node_count,omitempty"`
	PacketKeys          []string  `json:"packet_keys,omitempty"`
	MonitorItems        []string  `json:"monitor_items,omitempty"`
	WeakenedAssumptions []string  `json:"weakened_assumptions,omitempty"`
}

type TripCodeMemorySnapshot struct {
	SessionID     string                `json:"session_id"`
	Available     bool                  `json:"available"`
	Turns         int                   `json:"turns"`
	LastTripCode  string                `json:"last_tripcode,omitempty"`
	LastIssuer    string                `json:"last_issuer,omitempty"`
	LastUpdatedAt time.Time             `json:"last_updated_at,omitempty"`
	Entries       []TripCodeMemoryEntry `json:"entries,omitempty"`
}

type Evidence struct {
	Source           string   `json:"source"`
	Title            string   `json:"title"`
	Observation      string   `json:"observation"`
	URL              string   `json:"url,omitempty"`
	SourceDate       string   `json:"source_date,omitempty"`
	ComputedAt       string   `json:"computed_at,omitempty"`
	FilingDate       string   `json:"filing_date,omitempty"`
	FiledAt          string   `json:"filed_at,omitempty"`
	Stale            *bool    `json:"stale,omitempty"`
	Caveats          []string `json:"caveats,omitempty"`
	QualityFlags     []string `json:"quality_flags,omitempty"`
	EvidenceHashes   []string `json:"evidence_hashes,omitempty"`
	PayloadMode      string   `json:"payload_mode,omitempty"`
	RouteProvenance  string   `json:"route_provenance,omitempty"`
	ProvenanceLabels []string `json:"provenance_labels,omitempty"`
}

type SpecialistResult struct {
	Agent      string     `json:"agent"`
	Summary    string     `json:"summary"`
	Confidence string     `json:"confidence"`
	Evidence   []Evidence `json:"evidence"`
}

type BriefResponse struct {
	Issuer           string                  `json:"issuer"`
	Question         string                  `json:"question"`
	GeneratedAt      time.Time               `json:"generated_at"`
	Mode             string                  `json:"mode"`
	Plan             []string                `json:"plan"`
	Findings         []SpecialistResult      `json:"findings"`
	EvidenceFidelity EvidenceFidelitySummary `json:"evidence_fidelity"`
	Brief            string                  `json:"brief"`
	NextAction       string                  `json:"next_action"`
	Disclosures      []string                `json:"disclosures"`
	Cost             *CostSnapshot           `json:"cost,omitempty"`
	Runtime          *RuntimeTelemetry       `json:"runtime,omitempty"`
}

type EvidenceFidelitySummary struct {
	Status           string   `json:"status"`
	PayloadModes     []string `json:"payload_modes,omitempty"`
	SourceDates      []string `json:"source_dates,omitempty"`
	ComputedAt       []string `json:"computed_at,omitempty"`
	FilingDates      []string `json:"filing_dates,omitempty"`
	FiledAt          []string `json:"filed_at,omitempty"`
	StaleMarkers     []string `json:"stale_markers,omitempty"`
	Caveats          []string `json:"caveats,omitempty"`
	QualityFlags     []string `json:"quality_flags,omitempty"`
	EvidenceHashes   []string `json:"evidence_hashes,omitempty"`
	RouteProvenance  []string `json:"route_provenance,omitempty"`
	ProvenanceLabels []string `json:"provenance_labels,omitempty"`
}

type CostSnapshot struct {
	Enabled         bool             `json:"enabled"`
	Source          string           `json:"source"`
	Currency        string           `json:"currency"`
	RequestKind     string           `json:"request_kind,omitempty"`
	RequestCostUSD  float64          `json:"request_cost_usd,omitempty"`
	TrackedSpentUSD float64          `json:"tracked_spent_usd,omitempty"`
	BudgetUSD       float64          `json:"budget_usd,omitempty"`
	RemainingUSD    float64          `json:"remaining_usd,omitempty"`
	OfficialBilling *BillingSnapshot `json:"official_billing,omitempty"`
	Note            string           `json:"note,omitempty"`
}

type BillingSnapshot struct {
	Available          bool    `json:"available"`
	Source             string  `json:"source"`
	ProjectID          string  `json:"project_id,omitempty"`
	BillingAccountName string  `json:"billing_account_name,omitempty"`
	BillingEnabled     bool    `json:"billing_enabled,omitempty"`
	SpentUSD           float64 `json:"spent_usd,omitempty"`
	CreditBudgetUSD    float64 `json:"credit_budget_usd,omitempty"`
	RemainingUSD       float64 `json:"remaining_usd,omitempty"`
	UpdatedAt          string  `json:"updated_at,omitempty"`
	Note               string  `json:"note,omitempty"`
}

type RuntimeTelemetry struct {
	RequestID     string             `json:"request_id"`
	TraceID       string             `json:"trace_id,omitempty"`
	Route         string             `json:"route"`
	Method        string             `json:"method"`
	StartedAt     time.Time          `json:"started_at"`
	DurationMS    int64              `json:"duration_ms"`
	Runtime       string             `json:"runtime"`
	Observability []string           `json:"observability,omitempty"`
	RateLimit     *RateLimitSnapshot `json:"rate_limit,omitempty"`
	Memory        *MemoryStoreStatus `json:"memory,omitempty"`
}

type RateLimitSnapshot struct {
	Enabled           bool   `json:"enabled"`
	Allowed           bool   `json:"allowed"`
	Limit             int    `json:"limit,omitempty"`
	Remaining         int    `json:"remaining,omitempty"`
	WindowSeconds     int    `json:"window_seconds,omitempty"`
	RetryAfterSeconds int    `json:"retry_after_seconds,omitempty"`
	KeyScope          string `json:"key_scope,omitempty"`
}

type MemoryStoreStatus struct {
	Backend    string `json:"backend"`
	Durable    bool   `json:"durable"`
	SessionID  string `json:"session_id,omitempty"`
	EntryLimit int    `json:"entry_limit,omitempty"`
	Loaded     bool   `json:"loaded,omitempty"`
	Persisted  bool   `json:"persisted,omitempty"`
	LastError  string `json:"last_error,omitempty"`
}
