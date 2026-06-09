package main

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRequestRuntimeTelemetryAndLog(t *testing.T) {
	t.Setenv("K_SERVICE", "cloud-run-service")
	req := httptest.NewRequest(http.MethodPost, "/v1/brief", nil)
	req.Header.Set("X-Request-ID", "request-1")
	req.Header.Set("X-Cloud-Trace-Context", "trace-123/span;o=1")
	rt := beginRequest(req, "/v1/brief")
	if rt.requestID != "request-1" || rt.traceID != "trace-123" || rt.runtime != "cloud-run-service" {
		t.Fatalf("runtime = %#v", rt)
	}
	telemetry := rt.Telemetry(nil, nil)
	if telemetry.RequestID != "request-1" || telemetry.TraceID != "trace-123" || telemetry.Route != "/v1/brief" || telemetry.Runtime != "cloud-run-service" || len(telemetry.Observability) == 0 {
		t.Fatalf("telemetry = %#v", telemetry)
	}
	var buf bytes.Buffer
	rt.Log(slog.New(slog.NewJSONHandler(&buf, nil)), http.StatusOK, "mode")
	if !strings.Contains(buf.String(), "request-1") || !strings.Contains(buf.String(), "mode") {
		t.Fatalf("log output = %s", buf.String())
	}
	rt.Log(nil, http.StatusOK, "ignored")
}

func TestRateLimiterAllowAndReset(t *testing.T) {
	limiter := NewRateLimiter(true, 1, time.Second)
	now := time.Date(2026, 6, 8, 12, 0, 0, 0, time.UTC)
	limiter.now = func() time.Time { return now }
	req := httptest.NewRequest(http.MethodGet, "/resolve", nil)
	req.Header.Set("X-Demo-Key", "judge")
	first := limiter.Allow(req)
	if !first.Enabled || !first.Allowed || first.Remaining != 0 || first.KeyScope != "demo-key" {
		t.Fatalf("first allow = %#v", first)
	}
	second := limiter.Allow(req)
	if second.Allowed || second.RetryAfterSeconds == 0 {
		t.Fatalf("second allow = %#v", second)
	}
	now = now.Add(time.Second)
	third := limiter.Allow(req)
	if !third.Allowed {
		t.Fatalf("third allow after reset = %#v", third)
	}
}

func TestRateLimiterDisabledAndKeyScopes(t *testing.T) {
	if got := (*RateLimiter)(nil).Allow(httptest.NewRequest(http.MethodGet, "/", nil)); got.Enabled || !got.Allowed {
		t.Fatalf("nil limiter = %#v", got)
	}
	disabled := NewRateLimiter(false, 0, 0)
	if got := disabled.Allow(httptest.NewRequest(http.MethodGet, "/", nil)); got.Enabled || !got.Allowed {
		t.Fatalf("disabled limiter = %#v", got)
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer token")
	if key, scope := rateLimitKey(req); key != "bearer:token" || scope != "bearer" {
		t.Fatalf("bearer key/scope = %q %q", key, scope)
	}
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	if key, scope := rateLimitKey(req); key != "ip:127.0.0.1" || scope != "ip" {
		t.Fatalf("ip hostport key/scope = %q %q", key, scope)
	}
	req.RemoteAddr = "not-host-port"
	if key, scope := rateLimitKey(req); key != "ip:not-host-port" || scope != "ip" {
		t.Fatalf("ip raw key/scope = %q %q", key, scope)
	}
	req.RemoteAddr = ""
	if key, scope := rateLimitKey(req); key != "anonymous" || scope != "anonymous" {
		t.Fatalf("anonymous key/scope = %q %q", key, scope)
	}
}

func TestRateLimitHeadersAndEnvHelpers(t *testing.T) {
	rr := httptest.NewRecorder()
	writeRateLimitHeaders(rr, NewRateLimiter(false, 0, 0).Allow(httptest.NewRequest(http.MethodGet, "/", nil)))
	if rr.Header().Get("X-RateLimit-Limit") != "" {
		t.Fatal("disabled rate limit should not write headers")
	}
	rr = httptest.NewRecorder()
	limiter := NewRateLimiter(true, 1, time.Second)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	writeRateLimitHeaders(rr, limiter.Allow(req))
	if rr.Header().Get("X-RateLimit-Limit") != "1" || rr.Header().Get("X-RateLimit-Window-Seconds") != "1" {
		t.Fatalf("headers = %#v", rr.Header())
	}
	rr = httptest.NewRecorder()
	writeRateLimitHeaders(rr, limiter.Allow(req))
	if rr.Header().Get("Retry-After") == "" {
		t.Fatalf("denied headers = %#v", rr.Header())
	}
	t.Setenv("GOOD_INT", "7")
	t.Setenv("BAD_INT", "bad")
	if envInt("GOOD_INT", 1) != 7 || envInt("BAD_INT", 1) != 1 || envInt("MISSING_INT", 2) != 2 {
		t.Fatal("envInt fallback failed")
	}
	t.Setenv("DELTASIGNAL_RATE_LIMIT_ENABLED", "true")
	t.Setenv("DELTASIGNAL_RATE_LIMIT_PER_WINDOW", "3")
	t.Setenv("DELTASIGNAL_RATE_LIMIT_WINDOW_SECONDS", "2")
	limiter = RateLimiterFromEnv()
	if limiter == nil || !limiter.enabled || limiter.limit != 3 || limiter.window != 2*time.Second {
		t.Fatalf("RateLimiterFromEnv = %#v", limiter)
	}
	if traceIDFromHeader("") != "" || traceIDFromHeader("abc/def") != "abc" || traceIDFromHeader("abc") != "abc" {
		t.Fatal("traceIDFromHeader failed")
	}
	if got := randomHex(2); len(got) != 4 {
		t.Fatalf("randomHex len = %q", got)
	}
	if got := randomHex(0); len(got) != 24 {
		t.Fatalf("randomHex default len = %q", got)
	}
	if maxInt64(-1, 0) != 0 || maxInt64(2, 0) != 2 {
		t.Fatal("maxInt64 failed")
	}
}
