package main

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aitrailblazer/deltasignal-ai-agent/internal/agent"
)

type requestRuntime struct {
	requestID string
	traceID   string
	route     string
	method    string
	startedAt time.Time
	runtime   string
}

func beginRequest(r *http.Request, route string) requestRuntime {
	requestID := strings.TrimSpace(r.Header.Get("X-Request-ID"))
	if requestID == "" {
		requestID = randomHex(12)
	}
	traceID := traceIDFromHeader(r.Header.Get("X-Cloud-Trace-Context"))
	if traceID == "" {
		traceID = requestID
	}
	return requestRuntime{
		requestID: requestID,
		traceID:   traceID,
		route:     route,
		method:    r.Method,
		startedAt: time.Now().UTC(),
		runtime:   envOr("K_SERVICE", "local"),
	}
}

func (rt requestRuntime) Telemetry(rate *agent.RateLimitSnapshot, memory *agent.MemoryStoreStatus) *agent.RuntimeTelemetry {
	return &agent.RuntimeTelemetry{
		RequestID:  rt.requestID,
		TraceID:    rt.traceID,
		Route:      rt.route,
		Method:     rt.method,
		StartedAt:  rt.startedAt,
		DurationMS: maxInt64(time.Since(rt.startedAt).Milliseconds(), 0),
		Runtime:    rt.runtime,
		Observability: []string{
			"structured-json-logs",
			"request-id",
			"trace-id",
			"duration-ms",
			"rate-limit-snapshot",
			"memory-backend-status",
		},
		RateLimit: rate,
		Memory:    memory,
	}
}

func (rt requestRuntime) Log(logger *slog.Logger, status int, mode string) {
	if logger == nil {
		return
	}
	logger.Info(
		"deltasignal request completed",
		"request_id", rt.requestID,
		"trace_id", rt.traceID,
		"route", rt.route,
		"method", rt.method,
		"status", status,
		"mode", mode,
		"duration_ms", maxInt64(time.Since(rt.startedAt).Milliseconds(), 0),
	)
}

type RateLimiter struct {
	mu      sync.Mutex
	enabled bool
	limit   int
	window  time.Duration
	now     func() time.Time
	buckets map[string]rateBucket
}

type rateBucket struct {
	start time.Time
	used  int
}

func NewRateLimiter(enabled bool, limit int, window time.Duration) *RateLimiter {
	if limit <= 0 {
		limit = 60
	}
	if window <= 0 {
		window = time.Minute
	}
	return &RateLimiter{
		enabled: enabled,
		limit:   limit,
		window:  window,
		now:     time.Now,
		buckets: make(map[string]rateBucket),
	}
}

func RateLimiterFromEnv() *RateLimiter {
	return NewRateLimiter(
		envBool("DELTASIGNAL_RATE_LIMIT_ENABLED", false),
		envInt("DELTASIGNAL_RATE_LIMIT_PER_WINDOW", envInt("DELTASIGNAL_RATE_LIMIT_PER_MINUTE", 60)),
		time.Duration(envInt("DELTASIGNAL_RATE_LIMIT_WINDOW_SECONDS", 60))*time.Second,
	)
}

func (l *RateLimiter) Allow(r *http.Request) agent.RateLimitSnapshot {
	if l == nil || !l.enabled {
		return agent.RateLimitSnapshot{Enabled: false, Allowed: true}
	}
	key, scope := rateLimitKey(r)
	now := l.now().UTC()

	l.mu.Lock()
	defer l.mu.Unlock()

	bucket := l.buckets[key]
	if bucket.start.IsZero() || now.Sub(bucket.start) >= l.window {
		bucket = rateBucket{start: now}
	}
	bucket.used++
	l.buckets[key] = bucket

	remaining := l.limit - bucket.used
	if remaining < 0 {
		remaining = 0
	}
	allowed := bucket.used <= l.limit
	retryAfter := 0
	if !allowed {
		retryAfter = int(bucket.start.Add(l.window).Sub(now).Seconds())
		if retryAfter < 1 {
			retryAfter = 1
		}
	}
	return agent.RateLimitSnapshot{
		Enabled:           true,
		Allowed:           allowed,
		Limit:             l.limit,
		Remaining:         remaining,
		WindowSeconds:     int(l.window.Seconds()),
		RetryAfterSeconds: retryAfter,
		KeyScope:          scope,
	}
}

func rateLimitKey(r *http.Request) (string, string) {
	if got := strings.TrimSpace(r.Header.Get("X-Demo-Key")); got != "" {
		return "demo-key:" + got, "demo-key"
	}
	if got := strings.TrimPrefix(strings.TrimSpace(r.Header.Get("Authorization")), "Bearer "); got != "" {
		return "bearer:" + got, "bearer"
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && strings.TrimSpace(host) != "" {
		return "ip:" + host, "ip"
	}
	if strings.TrimSpace(r.RemoteAddr) != "" {
		return "ip:" + r.RemoteAddr, "ip"
	}
	return "anonymous", "anonymous"
}

func writeRateLimitHeaders(w http.ResponseWriter, snapshot agent.RateLimitSnapshot) {
	if !snapshot.Enabled {
		return
	}
	w.Header().Set("X-RateLimit-Limit", strconv.Itoa(snapshot.Limit))
	w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(snapshot.Remaining))
	w.Header().Set("X-RateLimit-Window-Seconds", strconv.Itoa(snapshot.WindowSeconds))
	if snapshot.RetryAfterSeconds > 0 {
		w.Header().Set("Retry-After", strconv.Itoa(snapshot.RetryAfterSeconds))
	}
}

func envInt(key string, fallback int) int {
	value := strings.TrimSpace(getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

var getenv = func(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}

func randomHex(bytesLen int) string {
	if bytesLen <= 0 {
		bytesLen = 12
	}
	buf := make([]byte, bytesLen)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}

func traceIDFromHeader(header string) string {
	header = strings.TrimSpace(header)
	if header == "" {
		return ""
	}
	if idx := strings.Index(header, "/"); idx > 0 {
		return header[:idx]
	}
	return header
}

func maxInt64(value, minimum int64) int64 {
	if value < minimum {
		return minimum
	}
	return value
}
