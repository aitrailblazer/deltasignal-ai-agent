package agent

import (
	"testing"
	"time"
)

func TestTripCodeMemoryStoreRememberAndSnapshot(t *testing.T) {
	store := NewTripCodeMemoryStore(2)
	resp := TripCodeResearchResponse{
		TripCode:    "TF-SUB-9DA70A7F98",
		Issuer:      "HUT",
		GeneratedAt: time.Date(2026, 6, 7, 12, 0, 0, 0, time.UTC),
		Mode:        "live-mcp-tripcode",
		Packet: map[string]any{
			"article": map[string]any{
				"title": "Hut 8: The Re-Rating Has A Deadline",
			},
			"river": map[string]any{
				"nodes": []any{
					map[string]any{"tripcode": "TF-SUB-1"},
					map[string]any{"tripcode": "TF-SUB-2"},
				},
			},
			"thesis_map": map[string]any{
				"weakened_assumptions": []any{"Timing certainty weakened"},
				"monitor_next":         []any{"Power delivery milestones"},
			},
		},
	}

	snapshot := store.Remember("hut-demo", resp)
	if !snapshot.Available {
		t.Fatal("snapshot should be available after Remember")
	}
	if snapshot.LastTripCode != "TF-SUB-9DA70A7F98" || snapshot.LastIssuer != "HUT" {
		t.Fatalf("snapshot identity mismatch: %#v", snapshot)
	}
	if snapshot.Entries[0].ArticleTitle != "Hut 8: The Re-Rating Has A Deadline" {
		t.Fatalf("ArticleTitle = %q", snapshot.Entries[0].ArticleTitle)
	}
	if snapshot.Entries[0].RiverNodeCount != 2 {
		t.Fatalf("RiverNodeCount = %d, want 2", snapshot.Entries[0].RiverNodeCount)
	}
	if len(snapshot.Entries[0].MonitorItems) != 1 || snapshot.Entries[0].MonitorItems[0] != "Power delivery milestones" {
		t.Fatalf("MonitorItems = %#v", snapshot.Entries[0].MonitorItems)
	}

	snapshot = store.Snapshot("hut-demo")
	if snapshot.Turns != 1 {
		t.Fatalf("Turns = %d, want 1", snapshot.Turns)
	}
}

func TestTripCodeMemoryStoreCapsSessionEntries(t *testing.T) {
	store := NewTripCodeMemoryStore(2)
	for _, tripcode := range []string{"TF-SUB-1", "TF-SUB-2", "TF-SUB-3"} {
		store.Remember("session", TripCodeResearchResponse{
			TripCode:    tripcode,
			GeneratedAt: time.Now().UTC(),
			Mode:        "test",
			Packet:      map[string]any{"title": tripcode},
		})
	}

	snapshot := store.Snapshot("session")
	if snapshot.Turns != 2 {
		t.Fatalf("Turns = %d, want capped 2", snapshot.Turns)
	}
	if snapshot.Entries[0].TripCode != "TF-SUB-2" || snapshot.LastTripCode != "TF-SUB-3" {
		t.Fatalf("unexpected capped entries: %#v", snapshot.Entries)
	}
}

func TestNewSessionMemoryResponse(t *testing.T) {
	snapshot := TripCodeMemorySnapshot{
		SessionID:    "hut-demo",
		Available:    true,
		LastTripCode: "TF-SUB-9DA70A7F98",
		LastIssuer:   "HUT",
		Turns:        1,
	}
	resp := NewSessionMemoryResponse(TripCodeResearchRequest{
		SessionID: "hut-demo",
		Question:  "What weakened?",
	}, snapshot)
	if resp.Mode != "session-memory" {
		t.Fatalf("Mode = %q", resp.Mode)
	}
	if resp.Memory == nil || resp.Memory.LastTripCode != "TF-SUB-9DA70A7F98" {
		t.Fatalf("memory missing from response: %#v", resp.Memory)
	}
	if resp.Packet["question"] != "What weakened?" {
		t.Fatalf("question not preserved: %#v", resp.Packet)
	}
}

func TestTripCodeMemoryStoreEdgeBranches(t *testing.T) {
	var nilStore *TripCodeMemoryStore
	if snapshot := nilStore.Remember("session", TripCodeResearchResponse{}); snapshot.Available {
		t.Fatal("nil store remember should not be available")
	}
	if snapshot := nilStore.Snapshot("session"); snapshot.SessionID != "session" || snapshot.Available {
		t.Fatalf("nil store snapshot = %#v", snapshot)
	}

	store := NewTripCodeMemoryStore(0)
	if store.maxEntries != defaultTripCodeMemoryLimit {
		t.Fatalf("default maxEntries = %d", store.maxEntries)
	}
	if snapshot := store.Remember(" ", TripCodeResearchResponse{}); snapshot.SessionID != "" || snapshot.Available {
		t.Fatalf("empty session remember = %#v", snapshot)
	}
	if snapshot := store.Snapshot("missing"); snapshot.SessionID != "missing" || snapshot.Available || snapshot.Turns != 0 {
		t.Fatalf("missing snapshot = %#v", snapshot)
	}

	resp := TripCodeResearchResponse{
		TripCode:    "TF-SUB-NESTED",
		GeneratedAt: time.Date(2026, 6, 7, 12, 0, 0, 0, time.UTC),
		Mode:        "test",
		Packet: map[string]any{
			"article_title": "",
			"article":       map[string]any{"headline": "Nested Headline"},
			"article_nodes": []map[string]any{{"tripcode": "TF-SUB-1"}},
			"monitor_items": []string{"monitor one", "", "monitor two"},
			"prior_assumptions_weakened": []any{
				"cash timing",
				map[string]any{"claim": "power timing"},
				3,
			},
		},
	}
	snapshot := store.Remember("nested", resp)
	entry := snapshot.Entries[0]
	if entry.ArticleTitle != "Nested Headline" || entry.RiverNodeCount != 1 || len(entry.MonitorItems) != 2 || len(entry.WeakenedAssumptions) != 2 {
		t.Fatalf("nested entry = %#v", entry)
	}

	resp.Packet = map[string]any{
		"title":      "",
		"thesis_map": map[string]any{"river_nodes": []string{"a", "b"}, "monitor_items": []any{map[string]any{"summary": "watch filings"}}},
	}
	entry = newTripCodeMemoryEntry(resp)
	if entry.RiverNodeCount != 2 || len(entry.MonitorItems) != 1 {
		t.Fatalf("thesis map entry = %#v", entry)
	}

	resp.Packet = map[string]any{"river": map[string]any{"article_nodes": []any{1, 2, 3}}}
	if count := newTripCodeMemoryEntry(resp).RiverNodeCount; count != 3 {
		t.Fatalf("river article node count = %d", count)
	}

	resp.Packet = map[string]any{
		"status": "ready",
		"research_packet": map[string]any{
			"article": map[string]any{"title": "Live Nested Article"},
			"river":   map[string]any{"nodes": []any{1, 2, 3, 4}},
			"thesis_map": map[string]any{
				"monitor_next":         []any{"watch power"},
				"weakened_assumptions": []any{"watch timing"},
			},
		},
	}
	entry = newTripCodeMemoryEntry(resp)
	if entry.ArticleTitle != "Live Nested Article" || entry.RiverNodeCount != 4 || len(entry.MonitorItems) != 1 || len(entry.WeakenedAssumptions) != 1 {
		t.Fatalf("research_packet entry = %#v", entry)
	}

	if sortedKeys(nil) != nil {
		t.Fatal("nil sortedKeys should return nil")
	}
	if firstString(map[string]any{"title": map[string]any{"title": "Nested Title"}}, "title") != "Nested Title" {
		t.Fatal("nested firstString failed")
	}
	if firstString(map[string]any{}, "missing") != "" {
		t.Fatal("missing firstString should return empty")
	}
	if len(compactStringSlice([]string{"", "a", "b", "c"}, 2)) != 2 {
		t.Fatal("compactStringSlice limit failed")
	}
	if stringList("not-list") != nil || arrayLikeLen("not-list") != 0 {
		t.Fatal("non-list helpers should return empty")
	}
}
