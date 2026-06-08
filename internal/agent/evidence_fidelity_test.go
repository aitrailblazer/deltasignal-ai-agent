package agent

import "testing"

func TestEnrichEvidenceFromMCPExtractsProvenance(t *testing.T) {
	stale := false
	evidence := enrichEvidenceFromMCP(Evidence{Title: "MCP tool: demo"}, map[string]any{
		"source_date":      "2026-06-07",
		"computed_at":      "2026-06-08T17:37:01Z",
		"filing_date":      "2026-05-10",
		"stale":            stale,
		"caveats":          []any{"unaudited"},
		"quality_flags":    []any{"fresh"},
		"evidence_hashes":  []any{"abc123"},
		"route_provenance": "deltasignal_company_report",
		"provenance_labels": []any{
			map[string]any{"label": "SEC/XBRL"},
		},
	}, map[string]any{"payload_mode": "compact"})

	if evidence.SourceDate != "2026-06-07" || evidence.ComputedAt != "2026-06-08T17:37:01Z" || evidence.FilingDate != "2026-05-10" {
		t.Fatalf("dates not extracted: %#v", evidence)
	}
	if evidence.Stale == nil || *evidence.Stale {
		t.Fatalf("stale marker not extracted: %#v", evidence.Stale)
	}
	if evidence.PayloadMode != "compact" || evidence.RouteProvenance != "deltasignal_company_report" {
		t.Fatalf("mode/provenance not extracted: %#v", evidence)
	}
	if len(evidence.Caveats) != 1 || evidence.Caveats[0] != "unaudited" || len(evidence.QualityFlags) != 1 || evidence.QualityFlags[0] != "fresh" || len(evidence.EvidenceHashes) != 1 || evidence.EvidenceHashes[0] != "abc123" {
		t.Fatalf("arrays not extracted: %#v", evidence)
	}
	if len(evidence.ProvenanceLabels) != 1 || evidence.ProvenanceLabels[0] != "SEC/XBRL" {
		t.Fatalf("labels not extracted: %#v", evidence.ProvenanceLabels)
	}
}

func TestEnrichEvidenceFromMCPFillsOnlyMissingFields(t *testing.T) {
	stale := true
	evidence := enrichEvidenceFromMCP(Evidence{
		SourceDate:      "kept-source",
		ComputedAt:      "kept-computed",
		FilingDate:      "kept-filing",
		FiledAt:         "kept-filed",
		Stale:           &stale,
		PayloadMode:     "kept-mode",
		RouteProvenance: "kept-route",
		Caveats:         []string{"kept-caveat"},
		QualityFlags:    []string{"kept-quality"},
		EvidenceHashes:  []string{"kept-hash"},
	}, map[string]any{
		"source_date":      "new-source",
		"computed_at":      "new-computed",
		"filing_date":      "new-filing",
		"filed_at":         "new-filed",
		"stale":            false,
		"caveats":          []string{"new-caveat"},
		"quality_flags":    []string{"new-quality"},
		"evidence_hashes":  []string{"new-hash"},
		"route_provenance": "new-route",
	}, map[string]any{"output_mode": "ignored"})
	if evidence.SourceDate != "kept-source" || evidence.ComputedAt != "kept-computed" || evidence.FilingDate != "kept-filing" || evidence.FiledAt != "kept-filed" {
		t.Fatalf("existing date fields not preserved: %#v", evidence)
	}
	if evidence.Stale == nil || !*evidence.Stale || evidence.PayloadMode != "kept-mode" || evidence.RouteProvenance != "kept-route" {
		t.Fatalf("existing stale/mode/route not preserved: %#v", evidence)
	}
	if len(evidence.Caveats) != 2 || len(evidence.QualityFlags) != 2 || len(evidence.EvidenceHashes) != 2 {
		t.Fatalf("metadata not appended and deduped: %#v", evidence)
	}
}

func TestBuildEvidenceFidelitySummary(t *testing.T) {
	stale := true
	summary := BuildEvidenceFidelitySummary([]SpecialistResult{{
		Evidence: []Evidence{{
			Title:          "Company report",
			SourceDate:     "2026-06-07",
			ComputedAt:     "2026-06-08T17:37:01Z",
			Stale:          &stale,
			Caveats:        []string{"missing peer rows"},
			QualityFlags:   []string{"partial"},
			EvidenceHashes: []string{"hash1"},
			PayloadMode:    "compact",
		}},
	}})
	if summary.Status != "provenance-carried" {
		t.Fatalf("status = %q", summary.Status)
	}
	if len(summary.StaleMarkers) != 1 || summary.StaleMarkers[0] != "Company report stale=true" {
		t.Fatalf("stale markers = %#v", summary.StaleMarkers)
	}
	if line := evidenceFidelityLine(summary); line == "" || line == "Evidence fidelity: upstream packets did not provide richer provenance fields in this response." {
		t.Fatalf("unexpected fidelity line: %q", line)
	}

	minimal := BuildEvidenceFidelitySummary(nil)
	if minimal.Status != "minimal" {
		t.Fatalf("minimal status = %q", minimal.Status)
	}
}

func TestEvidenceFidelityHelpersCoverBranches(t *testing.T) {
	root := map[string]any{
		"empty":       "   ",
		"bool_string": "false",
		"bad_bool":    "maybe",
		"items": []map[string]any{
			{"message": "message caveat"},
			{"value": "value caveat"},
			{"nested": map[string]any{"x": "y"}},
		},
		"list":      []string{"a", "a", "b"},
		"nil_value": nil,
	}
	if got := firstStringForKeys(root, "empty", "missing"); got != "" {
		t.Fatalf("firstStringForKeys empty = %q", got)
	}
	if got := stringsForKeys(root, "list"); len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("stringsForKeys list = %#v", got)
	}
	var walked []string
	walkJSON(root["items"], func(key string, value any) bool {
		walked = append(walked, key)
		return true
	})
	if len(walked) == 0 {
		t.Fatalf("walkJSON did not traverse []map")
	}
	if walkJSON([]any{map[string]any{"stop": "now"}}, func(key string, value any) bool {
		return key != "stop"
	}) {
		t.Fatal("walkJSON []any stop should return false")
	}
	if walkJSON([]map[string]any{{"stop": "now"}}, func(key string, value any) bool {
		return key != "stop"
	}) {
		t.Fatal("walkJSON []map stop should return false")
	}
	if got, ok := firstBoolForKeys(root, "bool_string"); !ok || got {
		t.Fatalf("firstBoolForKeys false string = %t %t", got, ok)
	}
	if _, ok := firstBoolForKeys(root, "bad_bool"); ok {
		t.Fatal("bad bool string should not parse")
	}
	if got := coerceStringList(nil); got != nil {
		t.Fatalf("nil should not coerce: %#v", got)
	}
	if got := coerceStringList(7); len(got) != 1 || got[0] != "7" {
		t.Fatalf("number coerce = %#v", got)
	}
	if got := coerceStringList(map[string]any{"message": "message caveat"}); len(got) != 1 || got[0] != "message caveat" {
		t.Fatalf("message map coerce = %#v", got)
	}
	if got := coerceStringList(map[string]any{"value": "value caveat"}); len(got) != 1 || got[0] != "value caveat" {
		t.Fatalf("value map coerce = %#v", got)
	}
	if got := coerceStringList(map[string]any{"nested": map[string]any{"x": "y"}}); len(got) != 1 || got[0] == "" {
		t.Fatalf("generic map coerce = %#v", got)
	}
	if got := splitEvidenceString("   "); got != nil {
		t.Fatalf("empty split = %#v", got)
	}
	parts := appendLimited([]string{"base"}, "many", []string{"1", "2", "3"}, 2)
	if len(parts) != 2 || parts[1] != "many=1, 2" {
		t.Fatalf("appendLimited truncated = %#v", parts)
	}
	parts = appendLimited(parts, "all", []string{"x"}, 0)
	if parts[len(parts)-1] != "all=x" {
		t.Fatalf("appendLimited no limit = %#v", parts)
	}
}
