# DeltaSignal Gemini AI Agent

DeltaSignal Gemini AI Agent is a competition build for the Google for Startups AI Agents Challenge.

The project is a new Google Cloud-native autonomous issuer intelligence agent. It turns fragmented public-company stress signals, SEC-grounded evidence, peer context, and TripCode research memory into a concise B2B action brief for analysts, funds, and startup operators.

The competition idea is simple: TripCodes make financial research portable, inspectable, and executable by agents. A subscriber can give the DeltaSignal Gemini AI Agent, or another authorized MCP-aware workflow, one TripCode and ask it to resolve the article, reconstruct the thesis River, compare it with filing-backed evidence, and show what changed.

## Competition Positioning

- **Track:** Build (Net-new Agents)
- **Cloud project:** `startup-ai-deltasignal`
- **Runtime target:** Validated Cloud Run service at `https://deltasignal-ai-agent-mhaufviwaq-uc.a.run.app`; Agent Runtime remains a later path
- **Model path:** Gemini through Vertex AI / Gemini Enterprise Agent Platform
- **Agent pattern:** Coordinator plus specialist agents for stress, evidence, peer context, and review
- **Tooling:** MCP-compatible bounded tools and deterministic demo fixtures

Existing DeltaSignal systems are treated as authorized integrations and data sources. This repository is the new competition project.

The primary disclosed integration is **DeltaSignal ATLAS-7**, a public SEC/XBRL-grounded issuer-intelligence surface for crypto-exposed public companies. It exposes MCP, OpenAPI/REST, Arazzo workflow metadata, x402-compatible public access, and bounded workflows such as readiness checks, Morning Brief, Company Report, Pressure Board, Alpha Sweep, peer ranking, company fundamentals, risk distribution, alpha signals, and daily-change evidence.

The strongest product extension is **TripCode Research Memory**. DeltaSignal articles can expose stable resolver keys such as `TF-SUB-9DA70A7F98`; the agent resolves the article object, prior River nodes, filing evidence refs, caveats, scenario map, invalidation checklist, and monitor-next queue. In the HUT proof case, one TripCode recovered the current article plus 9 prior HUT research nodes and returned a structured thesis map.

The core resolver contract is `deltasignal_resolve_tripcode_research_packet`: one TripCode returns the article object, River continuity, filing-backed evidence refs, thesis evolution map, missing evidence, and the four required boundaries: article memory, evidence identity, resolver identity, and non-advice. The public ATLAS-7 MCP endpoint now exposes this composite tool, and the deployed Build demo passes its structured `research_packet` through the Go Cloud Run agent while preserving deterministic fallback as a safety path.

## Why This Is Different

Most filing agents can find documents or summarize disclosure changes. TripCode Research Memory goes further: it lets an agent reconstruct what DeltaSignal previously argued, which claims are now supported by filing evidence, which assumptions weakened, and what needs to be monitored next.

The HUT proof path is the current deployed proof for the core resolver flow:

- `TF-SUB-9DA70A7F98` resolves to the current HUT article.
- The resolver discovers the HUT River with 10 article nodes.
- The deployed Cloud Run service returns thesis changes, confirmed signal, weakened assumptions, bridge risks, proof milestones, invalidation logic, and monitor-next items.
- Gemini synthesis through Vertex AI returns a diligence summary from the resolved HUT packet.
- A second `/resolve` turn with the same `session_id` returns session-memory mode and preserves HUT River context.
- Public execution now resolves the live ATLAS-7 composite TripCode packet; fallback remains enabled only for resilience if the MCP endpoint is unavailable.

The composite packet contract is exposed through this Go service at `/v1/tripcode` and `/resolve`. The next competition hardening work is to promote the composite TripCode resolver into the public ATLAS-7 MCP endpoint, add observability traces, add synthetic HUT edge-case simulation, and record the short judge demo video.

## Repository Contents

- `index.html`: Professional landing/spec page for GitHub Pages or local preview.
- `CHANGELOG.html`: Public changelog with implementation changes, cost controls, coverage gates, and deployment posture.
- `assets/atlas7-evidence-pipeline.png`: Hero image asset sourced from the public ATLAS-7 site.
- `cmd/server`: HTTP service for the competition demo.
- `cmd/tripcode-demo`: Two-turn TripCode demo client for the 3-minute video path.
- `internal/agent`: Coordinator, specialist tool interface, deterministic demo tools, and Gemini synthesis hook.
- `scripts/judge-demo.sh`: Judge-friendly curl flow for TripCode resolution, Memory Bank follow-up, and usage ledger check.
- `Dockerfile`: Cloud Run-ready container image.
- `Makefile`: Local test/build/demo/deploy shortcuts.
- `cloudbuild.yaml`: Cloud Build and Cloud Run deployment template.
- `docs/DeltaSignal_AI_Agent_Submission_Packet_2026_06_04.html`: The single official judge-facing submission packet with architecture, data sources, findings, and demo plan.

## Local Setup

```bash
gcloud config set account constantine@aitrailblazer.com
gcloud config set project startup-ai-deltasignal
gcloud auth application-default set-quota-project startup-ai-deltasignal

source .envrc
make check
make run
```

Health check:

```bash
curl http://localhost:8080/health
```

Demo request:

```bash
curl -s http://localhost:8080/v1/brief \
  -H 'Content-Type: application/json' \
  -d '{"issuer":"MSTR","question":"What issuer stress or opportunity should an analyst investigate next?"}' | jq
```

## Judge Test Mode

For the strongest judging path, use Gemini synthesis with live ATLAS-7 MCP tools behind a temporary demo key:

```bash
export DELTASIGNAL_USE_GEMINI=true
export DELTASIGNAL_USE_LIVE_MCP=true
export DELTASIGNAL_DEMO_API_KEY="<temporary judge key>"
export MCP_API_KEY="<internal MCP key>"
```

This tests the real Gemini agent synthesis path and attempts live ATLAS-7 MCP evidence retrieval. The `/health` route stays public; protected routes require either `X-Demo-Key: <temporary judge key>` or `Authorization: Bearer <temporary judge key>` when `DELTASIGNAL_DEMO_API_KEY` is set.

Judge request:

```bash
curl -s https://deltasignal-ai-agent-mhaufviwaq-uc.a.run.app/v1/brief \
  -H 'Content-Type: application/json' \
  -H "X-Demo-Key: <temporary judge key>" \
  -d '{"issuer":"HUT","question":"What issuer stress or opportunity should an analyst investigate next?"}' | jq
```

Expected mode: `vertex-ai-gemini`.

Expected evidence source: `deltasignal-atlas-7-mcp`.

## Test Spend Visibility

For judge rehearsals and Google Cloud credit control, enable the local test-spend tracker. Every successful `/v1/brief` and `/v1/tripcode` response will include a `cost` object with estimated request cost, tracked session spend, configured budget, and estimated remaining budget.

Public implementation notes and cost-control changes are recorded in `CHANGELOG.html` and linked from the landing page.

```bash
export DELTASIGNAL_COST_TRACKING=true
export DELTASIGNAL_GOOGLE_CREDIT_BUDGET_USD=500
export DELTASIGNAL_ESTIMATED_BRIEF_COST_USD=0.02
export DELTASIGNAL_ESTIMATED_TRIPCODE_COST_USD=0.03
export DELTASIGNAL_ESTIMATED_SESSION_MEMORY_COST_USD=0.00
```

Check the current ledger:

```bash
curl -s http://localhost:8080/v1/usage \
  -H "X-Demo-Key: <temporary judge key>" | jq
```

This is intentionally labeled `local-estimate`. It is immediate and useful after every test request, but it is not the official Google account balance. Exact spend and remaining credits must come from Google Cloud Billing export or the Billing console, which can lag actual requests.

The response also includes an `official_billing` block. Leave it unavailable unless the values are copied from an official source:

```bash
export DELTASIGNAL_OFFICIAL_BILLING_AVAILABLE=true
export DELTASIGNAL_OFFICIAL_BILLING_SOURCE=cloud-billing-export
export DELTASIGNAL_BILLING_PROJECT_ID=startup-ai-deltasignal
export DELTASIGNAL_BILLING_ACCOUNT_NAME=billingAccounts/018530-615AB8-696753
export DELTASIGNAL_BILLING_ENABLED=true
export DELTASIGNAL_OFFICIAL_BILLING_SPENT_USD=12.34
export DELTASIGNAL_OFFICIAL_BILLING_CREDIT_BUDGET_USD=500
export DELTASIGNAL_OFFICIAL_BILLING_UPDATED_AT=2026-06-08T11:00:00Z
```

If `DELTASIGNAL_OFFICIAL_BILLING_AVAILABLE=false`, `/v1/usage` reports the linked project/account metadata when configured and marks exact official spend as unavailable. This is deliberate: the app must not fabricate a current bill.

## Live DeltaSignal ATLAS-7 MCP Mode

The repo can call the DeltaSignal ATLAS-7 MCP with an API key for internal testing. This is opt-in so the public demo remains deterministic and no secret is required.

```bash
export DELTASIGNAL_USE_LIVE_MCP=true
export DELTASIGNAL_ATLAS_BASE_URL=https://api.aitrailblazer.net
export MCP_API_KEY="<your local MCP key>"

GOWORK=off go run ./cmd/server
```

The live client attempts these MCP tools:

- `deltasignal_pressure_board` for the stress lane
- `deltasignal_company_report` for issuer evidence
- `deltasignal_peer_ranking` for peer context
- `deltasignal_resolve_tripcode_research_packet` for the TripCode / River proof path

Tool names can be overridden with `DELTASIGNAL_MCP_STRESS_TOOL`, `DELTASIGNAL_MCP_COMPANY_TOOL`, `DELTASIGNAL_MCP_PEER_TOOL`, and `DELTASIGNAL_MCP_TRIPCODE_TOOL`. Keep `MCP_API_KEY` in your shell, Keychain, Secret Manager, or deployment secret store. Do not commit it.

If `DELTASIGNAL_USE_LIVE_MCP=true` is set without an API key, the service falls back to deterministic demo tools. If live MCP is configured but the endpoint is unavailable or returns an error, the deployed service returns `live_mcp_error` and falls back to the deterministic HUT River fixture when `DELTASIGNAL_ENABLE_TRIPCODE_FALLBACK=true`.

Judge-friendly TripCode proof request:

```bash
curl -s "http://localhost:8080/resolve?tripcode=TF-SUB-9DA70A7F98&session_id=hut-demo&issuer=HUT&payload_mode=compact&include_filing_evidence=true&include_prior_articles=true&include_thesis_map=true" \
  -H "X-Demo-Key: <temporary judge key>" | jq
```

Judge-friendly Memory Bank follow-up:

```bash
curl -s -X POST "http://localhost:8080/resolve?session_id=hut-demo" \
  -H "X-Demo-Key: <temporary judge key>" \
  -H "Content-Type: text/plain" \
  --data "Using the previous HUT River, what are the top 3 things to monitor next and which assumptions weakened?" | jq
```

The `/resolve` shortcut is the demo-oriented wrapper for judges. It calls the same Go TripCode resolver and session memory path as `/v1/tripcode`.

TripCode proof request:

```bash
curl -s http://localhost:8080/v1/tripcode \
  -H 'Content-Type: application/json' \
  -H "X-Demo-Key: <temporary judge key>" \
  -d '{"session_id":"hut-demo","tripcode":"TF-SUB-9DA70A7F98","issuer":"HUT","payload_mode":"compact"}' | jq
```

The route returns the resolved packet plus fixed evidence boundaries: TripCodes are resolver keys, article memory is not official filing evidence, missing evidence remains missing, and the output is diligence triage rather than investment advice. In the current public deployment, the packet includes the live MCP `research_packet` with article, River, filing evidence, thesis map, provenance, and evidence boundaries; deterministic HUT fallback remains available only if live MCP fails.

When `DELTASIGNAL_USE_GEMINI=true`, the same route also returns `gemini_summary`. The summary is generated from the resolved MCP packet plus any `session_id` memory and must preserve the evidence boundaries. If Gemini fails, the route still returns the raw resolver packet.

Session memory check:

```bash
curl -s http://localhost:8080/v1/tripcode \
  -H 'Content-Type: application/json' \
  -H "X-Demo-Key: <temporary judge key>" \
  -d '{"session_id":"hut-demo","question":"Using the previous HUT River, what should stay in context?"}' | jq
```

This is the Go equivalent of the ADK Session + Memory Bank idea. A `session_id` stores a compact River memory snapshot for follow-up turns: last TripCode, issuer, article title, packet keys, River node count, monitor hints, and weakened-assumption hints. The heavy research object remains resolver-backed in the DeltaSignal MCP/Azure layer.

### Competition Demo Flow

Use this for the 3-minute video. It shows the strongest flow: one TripCode resolves the research packet, then the next turn proves session memory.

Terminal 1:

```bash
export GOOGLE_CLOUD_PROJECT=startup-ai-deltasignal
export GOOGLE_CLOUD_LOCATION=global
export GOOGLE_GENAI_USE_VERTEXAI=true
export DELTASIGNAL_USE_GEMINI=true
export DELTASIGNAL_USE_LIVE_MCP=true
export DELTASIGNAL_DEMO_API_KEY="<temporary judge key>"
export MCP_API_KEY="<internal MCP key>"
GOWORK=off go run ./cmd/server
```

Terminal 2:

```bash
export DELTASIGNAL_AGENT_URL=http://localhost:8080
export DELTASIGNAL_DEMO_API_KEY="<temporary judge key>"
make demo
```

For the exact judge curl script:

```bash
export DELTASIGNAL_AGENT_URL=http://localhost:8080
export DELTASIGNAL_DEMO_API_KEY="<temporary judge key>"
scripts/judge-demo.sh
```

For Cloud Run, set `DELTASIGNAL_AGENT_URL` to `https://deltasignal-ai-agent-mhaufviwaq-uc.a.run.app`. The Go demo prints mode, TripCode, issuer, title when available, River node count when available, optional Gemini summary, and the second-turn session memory snapshot. The curl demo prints the raw JSON packet, follow-up memory result, and current local-estimate usage ledger.

## Go Agent Path, Not Python Wrapper

The minimal TripCode wrapper for this submission is implemented in Go, not as a separate Python ADK app.

The Python-style `MCPTool` snippet is useful as architecture shorthand, but the shipped challenge code uses:

- Go Cloud Run service: `cmd/server`
- Go competition demo client: `cmd/tripcode-demo`
- Gemini synthesis through the Google Gen AI SDK for Go
- Live MCP JSON-RPC client: `internal/agent/atlas_mcp.go`
- TripCode resolver route: `POST /v1/tripcode`
- Session/River memory snapshot: `internal/agent/tripcode_memory.go`
- TripCode Gemini summary path: `GeminiSynthesizer.SynthesizeTripCode`
- Demo protection: `DELTASIGNAL_DEMO_API_KEY`
- MCP protection: `MCP_API_KEY` / `DELTASIGNAL_API_KEY`

If we later add full Go ADK launcher support, it should follow the official Go ADK shape: a `main.go` entry point using the Go ADK launcher and `agent.NewSingleLoader(...)`. Until then, the production-ready competition path is the current Go Cloud Run service plus live MCP tool call.

## Vertex AI / Gemini Mode

The default service path is deterministic demo mode so judges can run the workflow without private data access.

To enable Gemini synthesis through Vertex AI:

```bash
export GOOGLE_CLOUD_PROJECT=startup-ai-deltasignal
export GOOGLE_CLOUD_LOCATION=global
export GOOGLE_GENAI_USE_VERTEXAI=true
export DELTASIGNAL_USE_GEMINI=true
export GEMINI_MODEL=gemini-2.5-flash
GOWORK=off go run ./cmd/server
```

Do not use Google AI Studio API keys as the primary competition path. The challenge guide states that AI Studio Gemini usage is not available for Google Cloud credit usage.

## Deploy

Create an Artifact Registry Docker repository first if needed:

```bash
gcloud artifacts repositories create deltasignal-ai-agent \
  --repository-format=docker \
  --location=us-central1 \
  --project=startup-ai-deltasignal
```

Deploy with Cloud Build:

```bash
make deploy
```

For a direct source deploy during fast iteration:

```bash
make deploy-source
```

Both deployment paths use Cloud Run settings intended for the challenge demo window: `min-instances=0`, `max-instances=3`, `512Mi`, `1 CPU`, and concurrency `80`. Set secrets such as `MCP_API_KEY` and `DELTASIGNAL_DEMO_API_KEY` through Secret Manager or Cloud Run service configuration, not in the repository.

Current validated service:

```bash
SERVICE_URL=https://deltasignal-ai-agent-mhaufviwaq-uc.a.run.app
curl "$SERVICE_URL/health"
DELTASIGNAL_AGENT_URL="$SERVICE_URL" DELTASIGNAL_DEMO_API_KEY="<temporary judge key>" scripts/judge-demo.sh
```

Useful Make targets:

```bash
make help
make check
make coverage
make build
make docker-build
make demo
```

## Public Repo Safety

This repository intentionally excludes local challenge PDFs, temporary text extraction, credentials, API keys, ADC files, and private data. Any live DeltaSignal integrations must use Secret Manager, IAM, and bounded service accounts.
