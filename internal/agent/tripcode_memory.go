package agent

import (
	"sort"
	"strings"
	"sync"
	"time"
)

const defaultTripCodeMemoryLimit = 20

type TripCodeMemoryStore struct {
	mu         sync.Mutex
	maxEntries int
	sessions   map[string][]TripCodeMemoryEntry
}

func NewTripCodeMemoryStore(maxEntries int) *TripCodeMemoryStore {
	if maxEntries <= 0 {
		maxEntries = defaultTripCodeMemoryLimit
	}
	return &TripCodeMemoryStore{
		maxEntries: maxEntries,
		sessions:   make(map[string][]TripCodeMemoryEntry),
	}
}

func (s *TripCodeMemoryStore) Remember(sessionID string, response TripCodeResearchResponse) TripCodeMemorySnapshot {
	if s == nil {
		return TripCodeMemorySnapshot{}
	}
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return TripCodeMemorySnapshot{}
	}

	entry := newTripCodeMemoryEntry(response)
	s.mu.Lock()
	defer s.mu.Unlock()

	entries := append(s.sessions[sessionID], entry)
	if len(entries) > s.maxEntries {
		entries = entries[len(entries)-s.maxEntries:]
	}
	s.sessions[sessionID] = entries
	return snapshotForEntries(sessionID, entries)
}

func (s *TripCodeMemoryStore) Snapshot(sessionID string) TripCodeMemorySnapshot {
	if s == nil {
		return TripCodeMemorySnapshot{SessionID: strings.TrimSpace(sessionID)}
	}
	sessionID = strings.TrimSpace(sessionID)
	s.mu.Lock()
	defer s.mu.Unlock()

	entries := append([]TripCodeMemoryEntry(nil), s.sessions[sessionID]...)
	return snapshotForEntries(sessionID, entries)
}

func newTripCodeMemoryEntry(response TripCodeResearchResponse) TripCodeMemoryEntry {
	return TripCodeMemoryEntry{
		TripCode:            response.TripCode,
		Issuer:              response.Issuer,
		Mode:                response.Mode,
		GeneratedAt:         response.GeneratedAt,
		ArticleTitle:        firstString(response.Packet, "title", "article_title", "headline"),
		RiverNodeCount:      inferRiverNodeCount(response.Packet),
		PacketKeys:          sortedKeys(response.Packet),
		MonitorItems:        firstStringList(response.Packet, "monitor_next", "what_to_monitor_next", "monitor_items"),
		WeakenedAssumptions: firstStringList(response.Packet, "weakened_assumptions", "prior_assumptions_weakened", "assumptions_weakened"),
	}
}

func snapshotForEntries(sessionID string, entries []TripCodeMemoryEntry) TripCodeMemorySnapshot {
	snapshot := TripCodeMemorySnapshot{
		SessionID: sessionID,
		Turns:     len(entries),
		Entries:   entries,
	}
	if len(entries) == 0 {
		return snapshot
	}
	last := entries[len(entries)-1]
	snapshot.Available = true
	snapshot.LastTripCode = last.TripCode
	snapshot.LastIssuer = last.Issuer
	snapshot.LastUpdatedAt = last.GeneratedAt
	return snapshot
}

func sortedKeys(packet map[string]any) []string {
	if len(packet) == 0 {
		return nil
	}
	keys := make([]string, 0, len(packet))
	for key := range packet {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func firstString(packet map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := packet[key].(string); ok && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
		if nested, ok := packet[key].(map[string]any); ok {
			for _, nestedKey := range []string{"title", "headline", "article_title"} {
				if value, ok := nested[nestedKey].(string); ok && strings.TrimSpace(value) != "" {
					return strings.TrimSpace(value)
				}
			}
		}
	}
	if article, ok := packet["article"].(map[string]any); ok {
		return firstString(article, "title", "headline", "article_title")
	}
	return ""
}

func firstStringList(packet map[string]any, keys ...string) []string {
	for _, key := range keys {
		if values := stringList(packet[key]); len(values) > 0 {
			return values
		}
	}
	if thesisMap, ok := packet["thesis_map"].(map[string]any); ok {
		for _, key := range keys {
			if values := stringList(thesisMap[key]); len(values) > 0 {
				return values
			}
		}
	}
	return nil
}

func stringList(value any) []string {
	switch typed := value.(type) {
	case []string:
		return compactStringSlice(typed, 8)
	case []any:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			switch v := item.(type) {
			case string:
				if strings.TrimSpace(v) != "" {
					out = append(out, strings.TrimSpace(v))
				}
			case map[string]any:
				if text := firstString(v, "summary", "title", "claim", "item", "monitor"); text != "" {
					out = append(out, text)
				}
			}
		}
		return compactStringSlice(out, 8)
	default:
		return nil
	}
}

func compactStringSlice(values []string, limit int) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		out = append(out, value)
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out
}

func inferRiverNodeCount(packet map[string]any) int {
	for _, key := range []string{"river_nodes", "prior_article_nodes", "article_nodes"} {
		if count := arrayLikeLen(packet[key]); count > 0 {
			return count
		}
	}
	if river, ok := packet["river"].(map[string]any); ok {
		for _, key := range []string{"nodes", "article_nodes", "prior_article_nodes"} {
			if count := arrayLikeLen(river[key]); count > 0 {
				return count
			}
		}
	}
	if thesisMap, ok := packet["thesis_map"].(map[string]any); ok {
		for _, key := range []string{"river_nodes", "prior_article_nodes", "article_nodes"} {
			if count := arrayLikeLen(thesisMap[key]); count > 0 {
				return count
			}
		}
	}
	return 0
}

func arrayLikeLen(value any) int {
	switch typed := value.(type) {
	case []any:
		return len(typed)
	case []map[string]any:
		return len(typed)
	case []string:
		return len(typed)
	default:
		return 0
	}
}

func DefaultTripCodeDisclosures() []string {
	return []string{
		"TripCodes are proprietary DeltaSignal resolver keys, not SEC accession numbers, CIKs, filing periods, or XBRL concept identifiers.",
		"Article memory is DeltaSignal research memory; filing claims require filing-backed TF-XBRL or official SEC evidence.",
		"Missing River nodes, filing evidence, and DeltaSignal signal objects must remain marked missing or unresolved.",
		"DeltaSignal outputs are evidence-routing alerts and diligence triage only, not investment advice, recommendations, price targets, or order instructions.",
	}
}

func NewSessionMemoryResponse(req TripCodeResearchRequest, snapshot TripCodeMemorySnapshot) TripCodeResearchResponse {
	return TripCodeResearchResponse{
		TripCode:    snapshot.LastTripCode,
		Issuer:      snapshot.LastIssuer,
		GeneratedAt: time.Now().UTC(),
		Mode:        "session-memory",
		Packet: map[string]any{
			"question":       strings.TrimSpace(req.Question),
			"session_memory": snapshot,
			"next_step":      "Use this session memory as River continuity context, then resolve a new TripCode or call the Gemini synthesis layer for a fresh answer.",
		},
		Disclosures: DefaultTripCodeDisclosures(),
		Memory:      &snapshot,
	}
}
