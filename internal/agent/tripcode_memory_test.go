package agent

import (
	"os"
	"path/filepath"
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

func TestTripCodeMemoryStoreFilePersistence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "memory", "sessions.json")
	store := NewTripCodeMemoryStore(2)
	if err := store.EnableFilePersistence(path); err != nil {
		t.Fatalf("EnableFilePersistence returned error: %v", err)
	}
	if status := store.Status("hut"); status.Backend != "file" || !status.Durable || !status.Loaded || status.Persisted {
		t.Fatalf("initial status = %#v", status)
	}
	store.Remember("hut", TripCodeResearchResponse{
		TripCode:    "TF-SUB-9DA70A7F98",
		Issuer:      "HUT",
		GeneratedAt: time.Date(2026, 6, 8, 12, 0, 0, 0, time.UTC),
		Mode:        "test",
		Packet:      map[string]any{"article": map[string]any{"title": "Hut 8"}},
	})
	if status := store.Status("hut"); !status.Persisted || status.LastError != "" || status.EntryLimit != 2 || status.SessionID != "hut" {
		t.Fatalf("persisted status = %#v", status)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("memory file not written: %v", err)
	}
	if len(raw) == 0 {
		t.Fatal("memory file was empty")
	}

	reloaded := NewTripCodeMemoryStore(2)
	if err := reloaded.EnableFilePersistence(path); err != nil {
		t.Fatalf("reload persistence returned error: %v", err)
	}
	snapshot := reloaded.Snapshot("hut")
	if !snapshot.Available || snapshot.LastTripCode != "TF-SUB-9DA70A7F98" {
		t.Fatalf("reloaded snapshot = %#v", snapshot)
	}
}

func TestTripCodeMemoryStorePersistenceErrors(t *testing.T) {
	var nilStore *TripCodeMemoryStore
	if err := nilStore.EnableFilePersistence("x"); err != nil {
		t.Fatalf("nil EnableFilePersistence returned error: %v", err)
	}
	if status := nilStore.Status("x"); status.Backend != "not-configured" {
		t.Fatalf("nil status = %#v", status)
	}
	store := NewTripCodeMemoryStore(1)
	if err := store.EnableFilePersistence(" "); err != nil {
		t.Fatalf("blank EnableFilePersistence returned error: %v", err)
	}
	if status := store.Status("memory"); status.Backend != "memory" || status.Durable {
		t.Fatalf("memory status = %#v", status)
	}
	store.backend = ""
	if status := store.Status("memory"); status.Backend != "memory" {
		t.Fatalf("empty backend status = %#v", status)
	}
	badPath := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(badPath, []byte("{bad"), 0o600); err != nil {
		t.Fatalf("write bad file: %v", err)
	}
	if err := store.EnableFilePersistence(badPath); err == nil {
		t.Fatal("expected bad JSON persistence load error")
	}
	if status := store.Status("bad"); status.LastError == "" || status.Backend != "file" {
		t.Fatalf("bad status = %#v", status)
	}

	dirPath := t.TempDir()
	if err := store.EnableFilePersistence(dirPath); err == nil {
		t.Fatal("expected directory read persistence error")
	}

	blocker := filepath.Join(t.TempDir(), "blocker")
	if err := os.WriteFile(blocker, []byte("not a directory"), 0o600); err != nil {
		t.Fatalf("write blocker: %v", err)
	}
	store = NewTripCodeMemoryStore(1)
	store.backend = "file"
	store.path = filepath.Join(blocker, "sessions.json")
	store.Remember("hut", TripCodeResearchResponse{TripCode: "TF-SUB-X", GeneratedAt: time.Now().UTC(), Packet: map[string]any{"title": "x"}})
	if status := store.Status("hut"); status.LastError == "" || status.Persisted {
		t.Fatalf("mkdir error status = %#v", status)
	}

	store = NewTripCodeMemoryStore(1)
	store.backend = "file"
	store.path = t.TempDir()
	store.Remember("hut", TripCodeResearchResponse{TripCode: "TF-SUB-X", GeneratedAt: time.Now().UTC(), Packet: map[string]any{"title": "x"}})
	if status := store.Status("hut"); status.LastError == "" || status.Persisted {
		t.Fatalf("write error status = %#v", status)
	}
}

func TestTripCodeMemoryStoreLoadCapsPersistedEntries(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sessions.json")
	raw := `{"hut":[
	  {"tripcode":"TF-SUB-1","generated_at":"2026-06-08T12:00:00Z","mode":"test"},
	  {"tripcode":"TF-SUB-2","generated_at":"2026-06-08T12:01:00Z","mode":"test"},
	  {"tripcode":"TF-SUB-3","generated_at":"2026-06-08T12:02:00Z","mode":"test"}
	]}`
	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		t.Fatalf("write persisted sessions: %v", err)
	}
	store := NewTripCodeMemoryStore(2)
	if err := store.EnableFilePersistence(path); err != nil {
		t.Fatalf("EnableFilePersistence returned error: %v", err)
	}
	snapshot := store.Snapshot("hut")
	if snapshot.Turns != 2 || snapshot.Entries[0].TripCode != "TF-SUB-2" || snapshot.LastTripCode != "TF-SUB-3" {
		t.Fatalf("capped persisted snapshot = %#v", snapshot)
	}

	nullPath := filepath.Join(t.TempDir(), "null-sessions.json")
	if err := os.WriteFile(nullPath, []byte("null"), 0o600); err != nil {
		t.Fatalf("write null sessions: %v", err)
	}
	nullStore := NewTripCodeMemoryStore(2)
	if err := nullStore.EnableFilePersistence(nullPath); err != nil {
		t.Fatalf("EnableFilePersistence null returned error: %v", err)
	}
	if snapshot := nullStore.Snapshot("hut"); snapshot.Available {
		t.Fatalf("null snapshot should not be available: %#v", snapshot)
	}
}
