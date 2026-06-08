package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	demoStdout     io.Writer = os.Stdout
	demoStderr     io.Writer = os.Stderr
	demoExit                 = os.Exit
	demoHTTPClient           = http.DefaultClient
)

type tripcodeRequest struct {
	SessionID   string `json:"session_id,omitempty"`
	TripCode    string `json:"tripcode,omitempty"`
	Issuer      string `json:"issuer,omitempty"`
	PayloadMode string `json:"payload_mode,omitempty"`
	Question    string `json:"question,omitempty"`
}

type tripcodeResponse struct {
	TripCode      string         `json:"tripcode"`
	Issuer        string         `json:"issuer"`
	Mode          string         `json:"mode"`
	Packet        map[string]any `json:"packet"`
	GeminiSummary string         `json:"gemini_summary"`
	Cost          *costSnapshot  `json:"cost"`
	Memory        *struct {
		SessionID    string `json:"session_id"`
		Available    bool   `json:"available"`
		Turns        int    `json:"turns"`
		LastTripCode string `json:"last_tripcode"`
		LastIssuer   string `json:"last_issuer"`
		Entries      []struct {
			ArticleTitle        string   `json:"article_title"`
			RiverNodeCount      int      `json:"river_node_count"`
			MonitorItems        []string `json:"monitor_items"`
			WeakenedAssumptions []string `json:"weakened_assumptions"`
		} `json:"entries"`
	} `json:"memory"`
}

type costSnapshot struct {
	Enabled         bool    `json:"enabled"`
	Source          string  `json:"source"`
	RequestKind     string  `json:"request_kind"`
	RequestCostUSD  float64 `json:"request_cost_usd"`
	TrackedSpentUSD float64 `json:"tracked_spent_usd"`
	BudgetUSD       float64 `json:"budget_usd"`
	RemainingUSD    float64 `json:"remaining_usd"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	baseURL := envOr("DELTASIGNAL_AGENT_URL", "http://localhost:8080")
	demoKey := strings.TrimSpace(os.Getenv("DELTASIGNAL_DEMO_API_KEY"))
	tripcode := envOr("DELTASIGNAL_DEMO_TRIPCODE", "TF-SUB-9DA70A7F98")
	issuer := envOr("DELTASIGNAL_DEMO_ISSUER", "HUT")
	sessionID := envOr("DELTASIGNAL_DEMO_SESSION_ID", "hut-demo")

	fmt.Fprintln(demoStdout, "=== DeltaSignal TripCode Demo: Turn 1, resolve article ===")
	first, err := callTripCode(ctx, baseURL, demoKey, tripcodeRequest{
		SessionID:   sessionID,
		TripCode:    tripcode,
		Issuer:      issuer,
		PayloadMode: "compact",
	})
	if err != nil {
		exitf("turn 1 failed: %v", err)
	}
	printResolved(first)

	fmt.Fprintln(demoStdout)
	fmt.Fprintln(demoStdout, "=== DeltaSignal TripCode Demo: Turn 2, use session memory ===")
	second, err := callTripCode(ctx, baseURL, demoKey, tripcodeRequest{
		SessionID: sessionID,
		Question:  "Using the previous River, what should stay in context?",
	})
	if err != nil {
		exitf("turn 2 failed: %v", err)
	}
	printMemory(second)
	printCost(second)
}

func callTripCode(ctx context.Context, baseURL, demoKey string, payload tripcodeRequest) (tripcodeResponse, error) {
	raw, _ := json.Marshal(payload)
	url := strings.TrimRight(baseURL, "/") + "/v1/tripcode"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return tripcodeResponse{}, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	if demoKey != "" {
		req.Header.Set("X-Demo-Key", demoKey)
	}

	resp, err := demoHTTPClient.Do(req)
	if err != nil {
		return tripcodeResponse{}, fmt.Errorf("call %s: %w", url, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return tripcodeResponse{}, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return tripcodeResponse{}, fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var out tripcodeResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return tripcodeResponse{}, fmt.Errorf("decode response: %w", err)
	}
	return out, nil
}

func printResolved(resp tripcodeResponse) {
	fmt.Fprintf(demoStdout, "Mode: %s\n", resp.Mode)
	fmt.Fprintf(demoStdout, "TripCode: %s\n", resp.TripCode)
	fmt.Fprintf(demoStdout, "Issuer: %s\n", fallback(resp.Issuer, "unknown"))
	fmt.Fprintf(demoStdout, "Title: %s\n", fallback(extractTitle(resp.Packet), "not returned in compact packet"))
	fmt.Fprintf(demoStdout, "River nodes: %s\n", fallback(fmt.Sprint(extractRiverNodeCount(resp.Packet)), "unknown"))
	if resp.GeminiSummary != "" {
		fmt.Fprintln(demoStdout)
		fmt.Fprintln(demoStdout, "Gemini summary:")
		fmt.Fprintln(demoStdout, strings.TrimSpace(resp.GeminiSummary))
	}
	printCost(resp)
	printMemory(resp)
}

func printMemory(resp tripcodeResponse) {
	if resp.Memory == nil || !resp.Memory.Available {
		fmt.Fprintln(demoStdout, "Session memory: not available")
		return
	}
	fmt.Fprintf(demoStdout, "Session memory: %s, turns=%d, last=%s, issuer=%s\n", resp.Memory.SessionID, resp.Memory.Turns, resp.Memory.LastTripCode, resp.Memory.LastIssuer)
	if len(resp.Memory.Entries) == 0 {
		return
	}
	last := resp.Memory.Entries[len(resp.Memory.Entries)-1]
	if last.ArticleTitle != "" {
		fmt.Fprintf(demoStdout, "Remembered article: %s\n", last.ArticleTitle)
	}
	if last.RiverNodeCount > 0 {
		fmt.Fprintf(demoStdout, "Remembered River nodes: %d\n", last.RiverNodeCount)
	}
	if len(last.MonitorItems) > 0 {
		fmt.Fprintf(demoStdout, "Monitor next: %s\n", strings.Join(last.MonitorItems, "; "))
	}
	if len(last.WeakenedAssumptions) > 0 {
		fmt.Fprintf(demoStdout, "Weakened assumptions: %s\n", strings.Join(last.WeakenedAssumptions, "; "))
	}
}

func printCost(resp tripcodeResponse) {
	if resp.Cost == nil {
		return
	}
	if !resp.Cost.Enabled {
		fmt.Fprintln(demoStdout, "Cost: disabled")
		return
	}
	fmt.Fprintf(
		demoStdout,
		"Cost: request=$%.4f, tracked=$%.4f, remaining=$%.4f of $%.2f, source=%s\n",
		resp.Cost.RequestCostUSD,
		resp.Cost.TrackedSpentUSD,
		resp.Cost.RemainingUSD,
		resp.Cost.BudgetUSD,
		fallback(resp.Cost.Source, "local-estimate"),
	)
}

func extractTitle(packet map[string]any) string {
	for _, key := range []string{"title", "article_title", "headline"} {
		if text, ok := packet[key].(string); ok && strings.TrimSpace(text) != "" {
			return strings.TrimSpace(text)
		}
	}
	if article, ok := packet["article"].(map[string]any); ok {
		return extractTitle(article)
	}
	if rp, ok := packet["research_packet"].(map[string]any); ok {
		return extractTitle(rp)
	}
	return ""
}

func extractRiverNodeCount(packet map[string]any) int {
	for _, key := range []string{"river_nodes", "prior_article_nodes", "article_nodes"} {
		if count := arrayLen(packet[key]); count > 0 {
			return count
		}
	}
	for _, key := range []string{"river", "thesis_map", "research_packet"} {
		if nested, ok := packet[key].(map[string]any); ok {
			if count := extractRiverNodeCount(nested); count > 0 {
				return count
			}
			if number := numericInt(nested["node_count"]); number > 0 {
				return number
			}
		}
	}
	return 0
}

func arrayLen(value any) int {
	switch typed := value.(type) {
	case []any:
		return len(typed)
	default:
		return 0
	}
}

func numericInt(value any) int {
	switch typed := value.(type) {
	case int:
		return typed
	case float64:
		return int(typed)
	default:
		return 0
	}
}

func envOr(key, fallbackValue string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallbackValue
	}
	return value
}

func fallback(value, fallbackValue string) string {
	if strings.TrimSpace(value) == "" {
		return fallbackValue
	}
	return value
}

func exitf(format string, args ...any) {
	fmt.Fprintf(demoStderr, format+"\n", args...)
	demoExit(1)
}
