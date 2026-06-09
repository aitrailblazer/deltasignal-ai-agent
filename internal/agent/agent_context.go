package agent

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"
)

const maxAgentContextBytes = 256 * 1024

var DefaultAgentContextURLs = []string{
	"https://aitrailblazer.github.io/deltasignal-atlas-codex-plugin/CLAUDE.md",
	"https://aitrailblazer.github.io/deltasignal-atlas-codex-plugin/llms-full.txt",
}

var FetchDefaultAgentContext = func(ctx context.Context) (AgentContextSnapshot, error) {
	return LoadAgentContext(ctx, DefaultAgentContextURLs)
}

func LoadAgentContext(ctx context.Context, urls []string) (AgentContextSnapshot, error) {
	snapshot := AgentContextSnapshot{
		Enabled: true,
		Purpose: "Public ATLAS-7 operating guides used by the cloud agent to select the TripCode/River/MCP evidence workflow inside the hosted runtime.",
		Sources: make([]AgentContextSource, 0, len(urls)),
	}
	client := &http.Client{Timeout: 6 * time.Second}
	for _, rawURL := range urls {
		source := AgentContextSource{URL: rawURL}
		if !allowedAgentContextURL(rawURL) {
			source.Status = "blocked"
			source.Error = "URL is not in the public ATLAS-7 agent context allowlist"
			snapshot.Sources = append(snapshot.Sources, source)
			continue
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
		if err != nil {
			source.Status = "unavailable"
			source.Error = err.Error()
			snapshot.Sources = append(snapshot.Sources, source)
			continue
		}
		resp, err := client.Do(req)
		if err != nil {
			source.Status = "unavailable"
			source.Error = err.Error()
			snapshot.Sources = append(snapshot.Sources, source)
			continue
		}
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, maxAgentContextBytes+1))
		closeErr := resp.Body.Close()
		source.HTTPStatus = resp.StatusCode
		if readErr != nil {
			source.Status = "unavailable"
			source.Error = readErr.Error()
			snapshot.Sources = append(snapshot.Sources, source)
			continue
		}
		if closeErr != nil {
			source.Status = "unavailable"
			source.Error = closeErr.Error()
			snapshot.Sources = append(snapshot.Sources, source)
			continue
		}
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			source.Status = "unavailable"
			source.Error = fmt.Sprintf("HTTP status %d", resp.StatusCode)
			snapshot.Sources = append(snapshot.Sources, source)
			continue
		}
		if len(body) > maxAgentContextBytes {
			body = body[:maxAgentContextBytes]
			source.Error = "truncated at max_agent_context_bytes"
		}
		sum := sha256.Sum256(body)
		source.Status = "fetched"
		source.Bytes = len(body)
		source.SHA256 = hex.EncodeToString(sum[:])
		source.Excerpt = compactAgentContextExcerpt(string(body), 1400)
		snapshot.Sources = append(snapshot.Sources, source)
	}
	return snapshot, nil
}

func allowedAgentContextURL(rawURL string) bool {
	for _, allowed := range DefaultAgentContextURLs {
		if rawURL == allowed {
			return true
		}
	}
	return false
}

func compactAgentContextExcerpt(text string, limit int) string {
	text = strings.Join(strings.Fields(text), " ")
	if limit <= 0 || len(text) <= limit {
		return text
	}
	for !utf8.ValidString(text[:limit]) && limit > 0 {
		limit--
	}
	return strings.TrimSpace(text[:limit]) + "..."
}
