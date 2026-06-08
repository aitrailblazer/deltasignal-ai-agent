package agent

import (
	"fmt"
	"sort"
	"strings"
)

func enrichEvidenceFromMCP(e Evidence, result map[string]any, args map[string]any) Evidence {
	if e.PayloadMode == "" {
		if mode, ok := args["payload_mode"].(string); ok {
			e.PayloadMode = strings.TrimSpace(mode)
		}
	}
	if e.PayloadMode == "" {
		if mode, ok := args["output_mode"].(string); ok {
			e.PayloadMode = strings.TrimSpace(mode)
		}
	}
	if e.SourceDate == "" {
		e.SourceDate = firstStringForKeys(result, "source_date", "sourceDate", "as_of_date", "data_date", "period_end", "period_end_date")
	}
	if e.ComputedAt == "" {
		e.ComputedAt = firstStringForKeys(result, "computed_at", "computedAt", "generated_at", "updated_at", "last_updated")
	}
	if e.FilingDate == "" {
		e.FilingDate = firstStringForKeys(result, "filing_date", "filingDate", "filed_date", "filedDate")
	}
	if e.FiledAt == "" {
		e.FiledAt = firstStringForKeys(result, "filed_at", "filedAt", "accepted_at", "acceptance_datetime")
	}
	if e.Stale == nil {
		if stale, ok := firstBoolForKeys(result, "stale", "is_stale", "stale_flag", "isStale"); ok {
			e.Stale = &stale
		}
	}
	e.Caveats = uniqueStrings(append(e.Caveats, stringsForKeys(result, "caveats", "caveat", "warnings", "warning", "missing_or_unresolved")...))
	e.QualityFlags = uniqueStrings(append(e.QualityFlags, stringsForKeys(result, "quality_flags", "qualityFlags", "quality_flag", "quality", "flags")...))
	e.EvidenceHashes = uniqueStrings(append(e.EvidenceHashes, stringsForKeys(result, "evidence_hashes", "evidenceHashes", "evidence_hash", "sha256", "content_sha256", "response_hash", "source_hash")...))
	if e.RouteProvenance == "" {
		e.RouteProvenance = firstStringForKeys(result, "route_provenance", "routeProvenance", "route", "tool", "tool_name", "source_route")
	}
	e.ProvenanceLabels = uniqueStrings(append(e.ProvenanceLabels, stringsForKeys(result, "provenance_labels", "provenanceLabels", "provenance", "labels")...))
	return e
}

func BuildEvidenceFidelitySummary(findings []SpecialistResult) EvidenceFidelitySummary {
	var summary EvidenceFidelitySummary
	for _, finding := range findings {
		for _, evidence := range finding.Evidence {
			summary.PayloadModes = appendIfNonEmpty(summary.PayloadModes, evidence.PayloadMode)
			summary.SourceDates = appendIfNonEmpty(summary.SourceDates, evidence.SourceDate)
			summary.ComputedAt = appendIfNonEmpty(summary.ComputedAt, evidence.ComputedAt)
			summary.FilingDates = appendIfNonEmpty(summary.FilingDates, evidence.FilingDate)
			summary.FiledAt = appendIfNonEmpty(summary.FiledAt, evidence.FiledAt)
			if evidence.Stale != nil {
				summary.StaleMarkers = appendIfNonEmpty(summary.StaleMarkers, fmt.Sprintf("%s stale=%t", evidence.Title, *evidence.Stale))
			}
			summary.Caveats = append(summary.Caveats, evidence.Caveats...)
			summary.QualityFlags = append(summary.QualityFlags, evidence.QualityFlags...)
			summary.EvidenceHashes = append(summary.EvidenceHashes, evidence.EvidenceHashes...)
			summary.RouteProvenance = appendIfNonEmpty(summary.RouteProvenance, evidence.RouteProvenance)
			summary.ProvenanceLabels = append(summary.ProvenanceLabels, evidence.ProvenanceLabels...)
		}
	}
	summary.PayloadModes = uniqueStrings(summary.PayloadModes)
	summary.SourceDates = uniqueStrings(summary.SourceDates)
	summary.ComputedAt = uniqueStrings(summary.ComputedAt)
	summary.FilingDates = uniqueStrings(summary.FilingDates)
	summary.FiledAt = uniqueStrings(summary.FiledAt)
	summary.StaleMarkers = uniqueStrings(summary.StaleMarkers)
	summary.Caveats = uniqueStrings(summary.Caveats)
	summary.QualityFlags = uniqueStrings(summary.QualityFlags)
	summary.EvidenceHashes = uniqueStrings(summary.EvidenceHashes)
	summary.RouteProvenance = uniqueStrings(summary.RouteProvenance)
	summary.ProvenanceLabels = uniqueStrings(summary.ProvenanceLabels)
	if len(summary.PayloadModes)+len(summary.SourceDates)+len(summary.ComputedAt)+len(summary.FilingDates)+len(summary.FiledAt)+len(summary.StaleMarkers)+len(summary.Caveats)+len(summary.QualityFlags)+len(summary.EvidenceHashes)+len(summary.RouteProvenance)+len(summary.ProvenanceLabels) == 0 {
		summary.Status = "minimal"
	} else {
		summary.Status = "provenance-carried"
	}
	return summary
}

func evidenceFidelityLine(summary EvidenceFidelitySummary) string {
	if summary.Status != "provenance-carried" {
		return "Evidence fidelity: upstream packets did not provide richer provenance fields in this response."
	}
	parts := []string{"Evidence fidelity: provenance-carried"}
	parts = appendLimited(parts, "payload_mode", summary.PayloadModes, 2)
	parts = appendLimited(parts, "source_date", summary.SourceDates, 2)
	parts = appendLimited(parts, "computed_at", summary.ComputedAt, 2)
	parts = appendLimited(parts, "filing_date", summary.FilingDates, 2)
	parts = appendLimited(parts, "stale", summary.StaleMarkers, 2)
	parts = appendLimited(parts, "quality_flags", summary.QualityFlags, 3)
	parts = appendLimited(parts, "caveats", summary.Caveats, 3)
	parts = appendLimited(parts, "evidence_hashes", summary.EvidenceHashes, 2)
	parts = appendLimited(parts, "route", summary.RouteProvenance, 2)
	return strings.Join(parts, "; ")
}

func appendLimited(parts []string, label string, values []string, limit int) []string {
	if len(values) == 0 {
		return parts
	}
	if limit > 0 && len(values) > limit {
		values = values[:limit]
	}
	return append(parts, label+"="+strings.Join(values, ", "))
}

func firstStringForKeys(root any, keys ...string) string {
	want := normalizedKeySet(keys...)
	var out string
	walkJSON(root, func(key string, value any) bool {
		if !want[normalizeEvidenceKey(key)] {
			return true
		}
		values := coerceStringList(value)
		if len(values) == 0 {
			return true
		}
		out = values[0]
		return false
	})
	return out
}

func stringsForKeys(root any, keys ...string) []string {
	want := normalizedKeySet(keys...)
	var out []string
	walkJSON(root, func(key string, value any) bool {
		if !want[normalizeEvidenceKey(key)] {
			return true
		}
		out = append(out, coerceStringList(value)...)
		return true
	})
	return uniqueStrings(out)
}

func firstBoolForKeys(root any, keys ...string) (bool, bool) {
	want := normalizedKeySet(keys...)
	var out bool
	var found bool
	walkJSON(root, func(key string, value any) bool {
		if !want[normalizeEvidenceKey(key)] {
			return true
		}
		switch v := value.(type) {
		case bool:
			out, found = v, true
			return false
		case string:
			normalized := strings.ToLower(strings.TrimSpace(v))
			if normalized == "true" || normalized == "false" {
				out, found = normalized == "true", true
				return false
			}
		}
		return true
	})
	return out, found
}

func walkJSON(root any, visit func(key string, value any) bool) bool {
	switch value := root.(type) {
	case map[string]any:
		keys := make([]string, 0, len(value))
		for key := range value {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			child := value[key]
			if !visit(key, child) {
				return false
			}
			if !walkJSON(child, visit) {
				return false
			}
		}
	case []any:
		for _, child := range value {
			if !walkJSON(child, visit) {
				return false
			}
		}
	case []map[string]any:
		for _, child := range value {
			if !walkJSON(child, visit) {
				return false
			}
		}
	}
	return true
}

func coerceStringList(value any) []string {
	switch v := value.(type) {
	case string:
		return splitEvidenceString(v)
	case []string:
		return uniqueStrings(v)
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			out = append(out, coerceStringList(item)...)
		}
		return uniqueStrings(out)
	case map[string]any:
		if text, ok := v["label"].(string); ok {
			return splitEvidenceString(text)
		}
		if text, ok := v["message"].(string); ok {
			return splitEvidenceString(text)
		}
		if text, ok := v["value"].(string); ok {
			return splitEvidenceString(text)
		}
		return []string{compactText(summarizeJSON(v), 240)}
	default:
		if value == nil {
			return nil
		}
		return []string{compactText(fmt.Sprint(value), 240)}
	}
}

func splitEvidenceString(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return []string{compactText(value, 240)}
}

func normalizedKeySet(keys ...string) map[string]bool {
	out := make(map[string]bool, len(keys))
	for _, key := range keys {
		out[normalizeEvidenceKey(key)] = true
	}
	return out
}

func normalizeEvidenceKey(key string) string {
	key = strings.ToLower(strings.TrimSpace(key))
	replacer := strings.NewReplacer("_", "", "-", "", ".", "", " ", "")
	return replacer.Replace(key)
}

func appendIfNonEmpty(values []string, value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return values
	}
	return append(values, compactText(value, 240))
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}
