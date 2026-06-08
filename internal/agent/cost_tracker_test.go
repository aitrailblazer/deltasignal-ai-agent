package agent

import "testing"

func costEqual(got float64, want float64) bool {
	const epsilon = 0.0000001
	diff := got - want
	if diff < 0 {
		diff = -diff
	}
	return diff < epsilon
}

func TestCostTrackerRecordAndSnapshot(t *testing.T) {
	tracker := NewCostTracker(CostTrackerConfig{
		Enabled:              true,
		BudgetUSD:            1,
		BriefCostUSD:         0.10,
		TripCodeCostUSD:      0.20,
		SessionMemoryCostUSD: 0.01,
		Billing: BillingTrackerConfig{
			Available:       true,
			Source:          "cloud-billing-export",
			ProjectID:       "startup-ai-deltasignal",
			BillingEnabled:  true,
			SpentUSD:        12.25,
			CreditBudgetUSD: 500,
			UpdatedAt:       "2026-06-08T10:00:00Z",
		},
	})

	if got := tracker.Record("brief"); got == nil || got.RequestKind != "brief" || got.RequestCostUSD != 0.10 || !costEqual(got.TrackedSpentUSD, 0.10) || !costEqual(got.RemainingUSD, 0.90) {
		t.Fatalf("brief snapshot = %#v", got)
	} else if got.OfficialBilling == nil || !got.OfficialBilling.Available || got.OfficialBilling.Source != "cloud-billing-export" || got.OfficialBilling.ProjectID != "startup-ai-deltasignal" || !costEqual(got.OfficialBilling.RemainingUSD, 487.75) {
		t.Fatalf("official billing snapshot = %#v", got.OfficialBilling)
	}
	if got := tracker.Record("tripcode"); got == nil || got.RequestCostUSD != 0.20 || !costEqual(got.TrackedSpentUSD, 0.30) || !costEqual(got.RemainingUSD, 0.70) {
		t.Fatalf("tripcode snapshot = %#v", got)
	}
	if got := tracker.Record("session-memory"); got == nil || got.RequestCostUSD != 0.01 || !costEqual(got.TrackedSpentUSD, 0.31) || !costEqual(got.RemainingUSD, 0.69) {
		t.Fatalf("session memory snapshot = %#v", got)
	}
	snapshot := tracker.Snapshot()
	if !snapshot.Enabled || snapshot.Source != "local-estimate" || snapshot.Currency != "USD" || !costEqual(snapshot.TrackedSpentUSD, 0.31) || !costEqual(snapshot.RemainingUSD, 0.69) {
		t.Fatalf("snapshot = %#v", snapshot)
	}
}

func TestCostTrackerDisabledAndNil(t *testing.T) {
	tracker := NewCostTracker(CostTrackerConfig{})
	if got := tracker.Record("brief"); got != nil {
		t.Fatalf("disabled Record = %#v, want nil", got)
	}
	if got := tracker.Snapshot(); got.Enabled || got.Source != "local-estimate" || got.Currency != "USD" {
		t.Fatalf("disabled Snapshot = %#v", got)
	} else if got.OfficialBilling == nil || got.OfficialBilling.Available || got.OfficialBilling.Note == "" {
		t.Fatalf("disabled official billing = %#v", got.OfficialBilling)
	}
	var nilTracker *CostTracker
	if got := nilTracker.Record("brief"); got != nil {
		t.Fatalf("nil Record = %#v, want nil", got)
	}
	if got := nilTracker.Snapshot(); got.Enabled || got.Source != "not-configured" || got.Currency != "USD" {
		t.Fatalf("nil Snapshot = %#v", got)
	} else if got.OfficialBilling == nil || got.OfficialBilling.Available || got.OfficialBilling.Note == "" {
		t.Fatalf("nil official billing = %#v", got.OfficialBilling)
	}
}

func TestCostTrackerBoundsAndUnknownKind(t *testing.T) {
	tracker := NewCostTracker(CostTrackerConfig{
		Enabled:         true,
		BudgetUSD:       -1,
		BriefCostUSD:    -1,
		TripCodeCostUSD: 2,
		Source:          " custom ",
		Billing: BillingTrackerConfig{
			Available:       true,
			CreditBudgetUSD: 1,
			SpentUSD:        2,
			RemainingUSD:    -3,
		},
	})
	if got := tracker.Record("unknown"); got == nil || got.RequestCostUSD != 0 || got.BudgetUSD != 0 || got.Source != "custom" {
		t.Fatalf("unknown kind snapshot = %#v", got)
	} else if got.OfficialBilling == nil || got.OfficialBilling.RemainingUSD != 0 {
		t.Fatalf("bounded official billing = %#v", got.OfficialBilling)
	}
	if got := tracker.Record("tripcode"); got == nil || got.RequestCostUSD != 2 || got.RemainingUSD != 0 {
		t.Fatalf("over-budget snapshot = %#v", got)
	}
	overBudget := NewCostTracker(CostTrackerConfig{Enabled: true, BudgetUSD: 1, TripCodeCostUSD: 2})
	if got := overBudget.Record("tripcode"); got == nil || got.RequestCostUSD != 2 || got.BudgetUSD != 1 || got.RemainingUSD != 0 {
		t.Fatalf("positive over-budget snapshot = %#v", got)
	}
	if got := nonNegative(-0.01); got != 0 {
		t.Fatalf("nonNegative(-0.01) = %f", got)
	}
	if got := nonNegative(0.25); got != 0.25 {
		t.Fatalf("nonNegative(0.25) = %f", got)
	}
}
