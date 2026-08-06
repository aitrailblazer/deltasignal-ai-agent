package aapldemo

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"time"
)

//go:embed fixtures/aapl_tool_results.json
var fixtureFS embed.FS

type FixtureClient struct {
	Results map[string]json.RawMessage
}

var FixtureGeneratedAt = time.Date(2026, 7, 30, 0, 0, 0, 0, time.UTC)

func NewFixtureClient() (FixtureClient, error) {
	raw, err := fixtureFS.ReadFile("fixtures/aapl_tool_results.json")
	if err != nil {
		return FixtureClient{}, err
	}
	var results map[string]json.RawMessage
	if err := json.Unmarshal(raw, &results); err != nil {
		return FixtureClient{}, fmt.Errorf("decode AAPL fixture: %w", err)
	}
	return FixtureClient{Results: results}, nil
}

func (c FixtureClient) CallTool(_ context.Context, tool string, _ map[string]any) (json.RawMessage, int, error) {
	raw, ok := c.Results[tool]
	if !ok {
		return nil, 404, fmt.Errorf("fixture has no result for %s", tool)
	}
	return append(json.RawMessage(nil), raw...), 200, nil
}
