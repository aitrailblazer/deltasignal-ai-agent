# AGENTS.md

This file governs agent work in the `aitrailblazer/deltasignal-ai-agent` repository.

## Project Boundary

DeltaSignal Gemini AI Agent is the Google for Startups AI Agents Challenge project for a Google Cloud-native autonomous issuer intelligence agent.

The agent turns SEC/XBRL-grounded Delta Signal ATLAS-7 evidence, SPECTRA field maps, peer context, and bounded research memory into concise B2B action briefs. It must preserve evidence boundaries and must not become an investment-advice bot.

Canonical GCP project: `startup-ai-deltasignal`

Expected runtime:

- Go backend on Cloud Run
- Gemini through Vertex AI / Google Gen AI SDK
- MCP-compatible tool boundary for Delta Signal ATLAS-7 evidence
- Coordinator plus specialist lanes for stress, evidence, peer context, and review
- JSON action brief output for demos and judge review

## Local Commands

Use these from the repository root:

```bash
GOWORK=off go test ./...
GOWORK=off go run ./cmd/server
curl http://localhost:8080/health
curl -s http://localhost:8080/v1/brief \
  -H 'Content-Type: application/json' \
  -d '{"issuer":"HUT","question":"What issuer stress or opportunity should an analyst investigate next?"}' | jq
```

The public demo endpoints are `POST /v1/brief`, `POST /v1/tripcode`, and the judge-friendly `GET/POST /resolve`. The health endpoint is `GET /health`.

## Evidence Rules

All outputs must stay evidence-bound.

Required preservation fields when returned by upstream evidence:

- `source_date`
- `computed_at`
- filing dates or filed dates
- stale flags
- caveats
- quality flags
- evidence hashes
- payload mode
- route provenance
- billing/cost metadata
- non-advice boundary

Never invent missing evidence. Missing facts, unavailable SPECTRA rows, unsupported tickers, stale public MCP tools, or missing TripCode blobs must be surfaced as missing or unavailable.

Do not produce buy, sell, short, long, hold, entry, exit, stop-loss, price-target, guaranteed-move, or personalized recommendation language.

Use this boundary when needed:

```text
Delta Signal outputs are evidence-routing alerts and diligence triage only. They are not investment advice, recommendations, price targets, or order instructions.
```

## ATLAS-7 And SPECTRA

ATLAS-7 is the evidence layer. It is generated from SEC CompanyFacts / XBRL, daily-change processing, issuer matching, ticker/CIK normalization, dated rows, history rows, caveats, manifests, and MCP-ready evidence envelopes.

SPECTRA is the field-map layer. It reads ATLAS-history-shaped rows, resolves source date and issuer identity, normalizes numeric/date/boolean fields, derives pressure/range labels, and returns field-map provenance, caveats, and structured soft-miss states.

SPECTRA is not a separate source of truth. It explains and maps generated ATLAS-7 history. If SPECTRA is unavailable for an issuer, preserve that as a soft miss.

## Default Workflow Order

Broad workflow:

1. Readiness
2. Morning Brief
3. Risk Distribution
4. Top Stressed
5. Alpha Opportunities
6. Issuer drilldowns only as needed

Issuer workflow:

1. Company Report
2. Company Fundamentals
3. Covenant Stress
4. Peer Ranking
5. Alpha Signals
6. SPECTRA Field Map
7. ATLAS history or point-in-time history

TripCode workflow:

1. Resolve the article `TF-SUB` TripCode.
2. List linked article/River nodes.
3. Build the article-centered thesis map.
4. Compare claims to filing-backed evidence.
5. Surface missing evidence, weakened assumptions, invalidation logic, and monitor-next items.

## MCP And Public Availability

Treat OpenAPI and MCP discovery as authoritative for live route/tool availability. Public discovery can be free; protected execution may require x402 payment, grant context, or an authorized internal key.

Do not claim a route or MCP tool is publicly production-ready unless current discovery confirms it. If local code supports a capability but the public MCP endpoint does not expose it, describe it as local/source-ready or planned-public, not live-public.

TripCode/River research is especially discovery-gated. The HUT proof path is valid as a deployed Cloud Run Build demo, but public MCP claims must remain conditional until the deployed ATLAS-7 MCP surface exposes the relevant composite tool and the required River/index blobs are present. The current Cloud Run service must preserve `live_mcp_error` and deterministic fallback disclosure when public MCP rejects `deltasignal_resolve_tripcode_research_packet`.

Known TripCode/HUT public-readiness caveat:

- `trendforge/v1/resolver/objects/TF-RIVER/TF-RIVER-HUT.json` may be required for River root resolution.
- `trendforge/v1/resolver/indexes/by_issuer/HUT.json` may be required for issuer-level River lookup.
- If either blob or the matching MCP tool is missing, return missing/unavailable status instead of fabricating continuity.

## Agent Behavior

Codex should make minimal scoped edits, preserve existing user work, run repo-local tests when touching code, and avoid introducing secrets.

Claude Code should read this file first, inspect current OpenAPI/MCP discovery before making live claims, and treat all Delta Signal outputs as evidence envelopes rather than free-form financial analysis.

Gemini/ADK wrappers should use the same evidence contract: tool calls first, synthesis second, no unsupported market chatter, no personalized advice.

## Competition Boundary

Delta Signal ATLAS-7 is an authorized integration/data source. The competition project is the agentic Google Cloud orchestration layer around that evidence: routing, synthesis, tool selection, action brief generation, and judge-facing demo flow.

The strongest demo path is:

1. User asks an issuer question.
2. Coordinator selects bounded ATLAS-7/MCP tools.
3. Specialist lanes retrieve stress, evidence, and peer context.
4. Gemini produces an evidence-preserving action brief.
5. The response includes caveats, source dates, metadata, and non-advice language.

For HUT or TripCode demos, clearly distinguish narrative thesis text from underlying ATLAS-7/MCP evidence. A thesis can be analytically useful without being fully evidence-backed until contract, financing, insider activity, valuation, and Delta Signal issuer-level outputs have been verified. If the deployed Go service uses deterministic HUT fallback because the public MCP composite tool is unavailable, say that directly.
