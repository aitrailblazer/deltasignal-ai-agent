# DeltaSignal Gemini AI Agent

DeltaSignal Gemini AI Agent is a competition build for the Google for Startups AI Agents Challenge.

The project is a new Google Cloud-native autonomous issuer intelligence agent. It turns fragmented public-company stress signals, SEC-grounded evidence, and peer context into a concise B2B action brief for analysts, funds, and startup operators.

## Competition Positioning

- **Track:** Build (Net-new Agents)
- **Cloud project:** `startup-ai-deltasignal`
- **Runtime target:** Cloud Run first; Agent Runtime if schedule allows
- **Model path:** Gemini through Vertex AI / Gemini Enterprise Agent Platform
- **Agent pattern:** Coordinator plus specialist agents for stress, evidence, peer context, and review
- **Tooling:** MCP-compatible bounded tools and deterministic demo fixtures

Existing DeltaSignal systems are treated as authorized integrations and data sources. This repository is the new competition project.

The primary disclosed integration is **DeltaSignal ATLAS-7**, a public SEC/XBRL-grounded issuer-intelligence surface for crypto-exposed public companies. It exposes MCP, OpenAPI/REST, Arazzo workflow metadata, x402-compatible public access, and bounded workflows such as readiness checks, Morning Brief, Company Report, Pressure Board, Alpha Sweep, peer ranking, company fundamentals, risk distribution, alpha signals, and daily-change evidence.

## Repository Contents

- `index.html`: Professional landing/spec page for GitHub Pages or local preview.
- `assets/atlas7-evidence-pipeline.png`: Hero image asset sourced from the public ATLAS-7 site.
- `cmd/server`: HTTP service for the competition demo.
- `internal/agent`: Coordinator, specialist tool interface, deterministic demo tools, and Gemini synthesis hook.
- `Dockerfile`: Cloud Run-ready container image.
- `cloudbuild.yaml`: Cloud Build and Cloud Run deployment template.
- `docs/DeltaSignal_AI_Agent_Submission_Packet_2026_06_04.html`: The single official judge-facing submission packet with architecture, data sources, findings, and demo plan.

## Local Setup

```bash
gcloud config set account constantine@aitrailblazer.com
gcloud config set project startup-ai-deltasignal
gcloud auth application-default set-quota-project startup-ai-deltasignal

source .envrc
GOWORK=off go test ./...
GOWORK=off go run ./cmd/server
```

Health check:

```bash
curl http://localhost:8080/healthz
```

Demo request:

```bash
curl -s http://localhost:8080/v1/brief \
  -H 'Content-Type: application/json' \
  -d '{"issuer":"MSTR","question":"What issuer stress or opportunity should an analyst investigate next?"}' | jq
```

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
gcloud builds submit \
  --project=startup-ai-deltasignal \
  --config=cloudbuild.yaml \
  --substitutions=_REGION=us-central1
```

## Public Repo Safety

This repository intentionally excludes local challenge PDFs, temporary text extraction, credentials, API keys, ADC files, and private data. Any live DeltaSignal integrations must use Secret Manager, IAM, and bounded service accounts.
