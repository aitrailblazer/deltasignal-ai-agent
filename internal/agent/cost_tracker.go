package agent

import (
	"strings"
	"sync"
)

const costEstimateNote = "Local estimate for demo/test requests. Exact Google Cloud account spend and remaining credits require Cloud Billing export or the Billing console."
const officialBillingUnavailableNote = "Official Google Cloud Billing spend is unavailable unless supplied from Cloud Billing export, Billing console, or another official billing source."

type CostTrackerConfig struct {
	Enabled              bool
	BudgetUSD            float64
	BriefCostUSD         float64
	TripCodeCostUSD      float64
	SessionMemoryCostUSD float64
	Source               string
	Billing              BillingTrackerConfig
}

type BillingTrackerConfig struct {
	Available          bool
	Source             string
	ProjectID          string
	BillingAccountName string
	BillingEnabled     bool
	SpentUSD           float64
	CreditBudgetUSD    float64
	RemainingUSD       float64
	UpdatedAt          string
}

type CostTracker struct {
	mu sync.Mutex

	enabled              bool
	budgetUSD            float64
	briefCostUSD         float64
	tripCodeCostUSD      float64
	sessionMemoryCostUSD float64
	source               string
	spentUSD             float64
	billing              BillingSnapshot
}

func NewCostTracker(config CostTrackerConfig) *CostTracker {
	source := strings.TrimSpace(config.Source)
	if source == "" {
		source = "local-estimate"
	}
	return &CostTracker{
		enabled:              config.Enabled,
		budgetUSD:            nonNegative(config.BudgetUSD),
		briefCostUSD:         nonNegative(config.BriefCostUSD),
		tripCodeCostUSD:      nonNegative(config.TripCodeCostUSD),
		sessionMemoryCostUSD: nonNegative(config.SessionMemoryCostUSD),
		source:               source,
		billing:              newBillingSnapshot(config.Billing),
	}
}

func (t *CostTracker) Record(kind string) *CostSnapshot {
	if t == nil {
		return nil
	}
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.enabled {
		return nil
	}
	requestCost := t.requestCostLocked(kind)
	t.spentUSD += requestCost
	return t.snapshotLocked(kind, requestCost)
}

func (t *CostTracker) Snapshot() CostSnapshot {
	if t == nil {
		return CostSnapshot{
			Enabled:         false,
			Source:          "not-configured",
			Currency:        "USD",
			OfficialBilling: unavailableBillingSnapshot("", "", false),
			Note:            costEstimateNote,
		}
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	return *t.snapshotLocked("", 0)
}

func (t *CostTracker) requestCostLocked(kind string) float64 {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "brief":
		return t.briefCostUSD
	case "tripcode":
		return t.tripCodeCostUSD
	case "session-memory":
		return t.sessionMemoryCostUSD
	default:
		return 0
	}
}

func (t *CostTracker) snapshotLocked(kind string, requestCost float64) *CostSnapshot {
	snapshot := &CostSnapshot{
		Enabled:         t.enabled,
		Source:          t.source,
		Currency:        "USD",
		RequestKind:     strings.TrimSpace(kind),
		RequestCostUSD:  requestCost,
		TrackedSpentUSD: t.spentUSD,
		BudgetUSD:       t.budgetUSD,
		OfficialBilling: cloneBillingSnapshot(t.billing),
		Note:            costEstimateNote,
	}
	if t.budgetUSD > 0 {
		remaining := t.budgetUSD - t.spentUSD
		if remaining < 0 {
			remaining = 0
		}
		snapshot.RemainingUSD = remaining
	}
	return snapshot
}

func nonNegative(value float64) float64 {
	if value < 0 {
		return 0
	}
	return value
}

func newBillingSnapshot(config BillingTrackerConfig) BillingSnapshot {
	source := strings.TrimSpace(config.Source)
	if source == "" {
		source = "unavailable"
	}
	snapshot := BillingSnapshot{
		Available:          config.Available,
		Source:             source,
		ProjectID:          strings.TrimSpace(config.ProjectID),
		BillingAccountName: strings.TrimSpace(config.BillingAccountName),
		BillingEnabled:     config.BillingEnabled,
		SpentUSD:           nonNegative(config.SpentUSD),
		CreditBudgetUSD:    nonNegative(config.CreditBudgetUSD),
		RemainingUSD:       nonNegative(config.RemainingUSD),
		UpdatedAt:          strings.TrimSpace(config.UpdatedAt),
	}
	if snapshot.Available {
		if snapshot.CreditBudgetUSD > 0 && snapshot.RemainingUSD == 0 {
			remaining := snapshot.CreditBudgetUSD - snapshot.SpentUSD
			if remaining > 0 {
				snapshot.RemainingUSD = remaining
			}
		}
		return snapshot
	}
	snapshot.Note = officialBillingUnavailableNote
	return snapshot
}

func unavailableBillingSnapshot(projectID, billingAccountName string, billingEnabled bool) *BillingSnapshot {
	snapshot := newBillingSnapshot(BillingTrackerConfig{
		ProjectID:          projectID,
		BillingAccountName: billingAccountName,
		BillingEnabled:     billingEnabled,
	})
	return &snapshot
}

func cloneBillingSnapshot(snapshot BillingSnapshot) *BillingSnapshot {
	clone := snapshot
	return &clone
}
