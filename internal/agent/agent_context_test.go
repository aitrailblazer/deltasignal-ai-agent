package agent

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errors.New("read failed") }
func (failingReader) Close() error             { return nil }

type closeErrorBody struct {
	*strings.Reader
}

func (b closeErrorBody) Close() error { return errors.New("close failed") }

type nopBody struct {
	*strings.Reader
}

func (b nopBody) Close() error { return nil }

func TestLoadAgentContextBlocksNonAllowlistedURL(t *testing.T) {
	snapshot, err := LoadAgentContext(context.Background(), []string{"https://example.com/not-allowed.txt"})
	if err != nil {
		t.Fatalf("LoadAgentContext returned error: %v", err)
	}
	if !snapshot.Enabled || len(snapshot.Sources) != 1 {
		t.Fatalf("unexpected snapshot: %#v", snapshot)
	}
	if snapshot.Sources[0].Status != "blocked" || !strings.Contains(snapshot.Sources[0].Error, "allowlist") {
		t.Fatalf("unexpected blocked source: %#v", snapshot.Sources[0])
	}
}

func TestFetchDefaultAgentContextUsesDefaultURLs(t *testing.T) {
	oldURLs := DefaultAgentContextURLs
	oldTransport := http.DefaultTransport
	t.Cleanup(func() {
		DefaultAgentContextURLs = oldURLs
		http.DefaultTransport = oldTransport
	})
	DefaultAgentContextURLs = []string{"https://aitrailblazer.github.io/deltasignal-atlas-codex-plugin/CLAUDE.md"}
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.String() != DefaultAgentContextURLs[0] {
			t.Fatalf("unexpected URL: %s", req.URL.String())
		}
		return &http.Response{StatusCode: http.StatusOK, Body: nopBody{strings.NewReader("guide text")}}, nil
	})
	snapshot, err := FetchDefaultAgentContext(context.Background())
	if err != nil {
		t.Fatalf("FetchDefaultAgentContext returned error: %v", err)
	}
	if len(snapshot.Sources) != 1 || snapshot.Sources[0].Status != "fetched" || snapshot.Sources[0].Bytes != len("guide text") || snapshot.Sources[0].SHA256 == "" {
		t.Fatalf("unexpected snapshot: %#v", snapshot)
	}
}

func TestLoadAgentContextHandlesUnavailableSources(t *testing.T) {
	oldURLs := DefaultAgentContextURLs
	oldTransport := http.DefaultTransport
	t.Cleanup(func() {
		DefaultAgentContextURLs = oldURLs
		http.DefaultTransport = oldTransport
	})
	DefaultAgentContextURLs = []string{
		":bad-url",
		"https://aitrailblazer.github.io/deltasignal-atlas-codex-plugin/CLAUDE.md",
		"https://aitrailblazer.github.io/deltasignal-atlas-codex-plugin/llms-full.txt",
	}

	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("network failed")
	})
	snapshot, err := LoadAgentContext(context.Background(), []string{":bad-url", DefaultAgentContextURLs[1]})
	if err != nil {
		t.Fatalf("LoadAgentContext returned error: %v", err)
	}
	if snapshot.Sources[0].Status != "unavailable" || snapshot.Sources[1].Status != "unavailable" {
		t.Fatalf("unexpected unavailable snapshot: %#v", snapshot.Sources)
	}

	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch req.URL.String() {
		case DefaultAgentContextURLs[1]:
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(failingReader{})}, nil
		default:
			return &http.Response{StatusCode: http.StatusOK, Body: closeErrorBody{strings.NewReader("ok")}}, nil
		}
	})
	snapshot, err = LoadAgentContext(context.Background(), []string{DefaultAgentContextURLs[1], DefaultAgentContextURLs[2]})
	if err != nil {
		t.Fatalf("LoadAgentContext returned error: %v", err)
	}
	if snapshot.Sources[0].Status != "unavailable" || !strings.Contains(snapshot.Sources[0].Error, "read failed") || snapshot.Sources[1].Status != "unavailable" || !strings.Contains(snapshot.Sources[1].Error, "close failed") {
		t.Fatalf("unexpected read/close snapshot: %#v", snapshot.Sources)
	}

	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusNotFound, Body: nopBody{strings.NewReader("missing")}}, nil
	})
	snapshot, err = LoadAgentContext(context.Background(), []string{DefaultAgentContextURLs[1]})
	if err != nil {
		t.Fatalf("LoadAgentContext returned error: %v", err)
	}
	if snapshot.Sources[0].Status != "unavailable" || !strings.Contains(snapshot.Sources[0].Error, "HTTP status 404") {
		t.Fatalf("unexpected status snapshot: %#v", snapshot.Sources[0])
	}
}

func TestLoadAgentContextTruncatesOversizedSource(t *testing.T) {
	oldTransport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = oldTransport })
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: nopBody{strings.NewReader(strings.Repeat("a", maxAgentContextBytes+10))}}, nil
	})
	snapshot, err := LoadAgentContext(context.Background(), []string{DefaultAgentContextURLs[0]})
	if err != nil {
		t.Fatalf("LoadAgentContext returned error: %v", err)
	}
	source := snapshot.Sources[0]
	if source.Status != "fetched" || source.Bytes != maxAgentContextBytes || !strings.Contains(source.Error, "truncated") {
		t.Fatalf("unexpected truncated source: %#v", source)
	}
}

func TestCompactAgentContextExcerpt(t *testing.T) {
	got := compactAgentContextExcerpt("alpha\n\n beta\tgamma delta", 16)
	if got != "alpha beta gamma..." {
		t.Fatalf("excerpt = %q", got)
	}
	if got := compactAgentContextExcerpt("short", 0); got != "short" {
		t.Fatalf("zero-limit excerpt = %q", got)
	}
	if got := compactAgentContextExcerpt("éclair", 1); got != "..." {
		t.Fatalf("utf8-trim excerpt = %q", got)
	}
}
