#!/usr/bin/env sh
set -eu

BASE_URL="${DELTASIGNAL_AGENT_URL:-http://localhost:8080}"
TRIPCODE="${DELTASIGNAL_DEMO_TRIPCODE:-TF-SUB-9DA70A7F98}"
SESSION_ID="${DELTASIGNAL_DEMO_SESSION_ID:-judge-demo-hut}"
DEMO_KEY="${DELTASIGNAL_DEMO_API_KEY:-}"
ARTIFACT_ROOT="${DELTASIGNAL_DEMO_ARTIFACT_ROOT:-artifacts/judge-demo}"
RUN_ID="${DELTASIGNAL_DEMO_RUN_ID:-$(date -u +%Y%m%dT%H%M%SZ)}"
ARTIFACT_DIR="$ARTIFACT_ROOT/$RUN_ID"

curl_auth() {
  if [ -n "$DEMO_KEY" ]; then
    curl -sS -H "X-Demo-Key: $DEMO_KEY" "$@"
  else
    curl -sS "$@"
  fi
}

save_request() {
  name="$1"
  method="$2"
  url="$3"
  body="${4:-}"
  cat > "$ARTIFACT_DIR/$name-request.json" <<EOF
{
  "method": "$method",
  "url": "$url",
  "headers": {
    "X-Demo-Key": "[REDACTED]"
  },
  "body": $(if [ -n "$body" ]; then printf '%s' "$body" | awk 'BEGIN { printf "\"" } { gsub(/\\/,"\\\\"); gsub(/"/,"\\\""); printf "%s%s", sep, $0; sep="\\n" } END { printf "\"" }'; else printf 'null'; fi)
}
EOF
}

save_build_phase_explanation() {
  cat > "$ARTIFACT_DIR/BUILD_PHASE_EXPLANATION.html" <<EOF
<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>DeltaSignal Judge Demo Evidence | Build Phase</title>
  <style>
    html { background: #06080d; color: #fff8ef; font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; line-height: 1.5; }
    body { margin: 0; padding: 28px; }
    main { max-width: 980px; margin: 0 auto; }
    .hero, section { border: 1px solid rgba(255,255,255,.14); background: rgba(18,20,28,.88); border-radius: 10px; padding: 22px; margin: 0 0 16px; }
    .eyebrow { color: #5df58b; font-size: 12px; font-weight: 900; letter-spacing: .08em; text-transform: uppercase; }
    h1 { margin: 8px 0 12px; font-size: clamp(32px, 6vw, 58px); line-height: 1; letter-spacing: 0; }
    h2 { margin: 0 0 10px; color: #ffbd5a; font-size: 22px; letter-spacing: 0; }
    p { margin: 0 0 10px; color: rgba(255,248,239,.78); }
    ol { margin: 0; padding-left: 22px; color: rgba(255,248,239,.82); }
    li { margin: 0 0 10px; }
    code { color: #72d9ff; background: rgba(114,217,255,.09); padding: 2px 5px; border-radius: 5px; }
    .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(210px, 1fr)); gap: 10px; margin-top: 12px; }
    .card { border: 1px solid rgba(255,255,255,.12); background: rgba(255,255,255,.04); border-radius: 8px; padding: 14px; }
    .card b { display: block; color: #fff8ef; margin-bottom: 6px; }
    pre { overflow: auto; border: 1px solid rgba(255,255,255,.12); background: #0b0e15; border-radius: 8px; padding: 14px; color: #d9f7ff; }
  </style>
</head>
<body>
  <main>
    <div class="hero">
      <div class="eyebrow">Google for Startups AI Agents Challenge 2026 | Track 1 Build</div>
      <h1>Working Demo Evidence Session</h1>
      <p><strong>Phase covered:</strong> Build.</p>
      <p>Working demo: a Go-native Cloud Run agent resolves a TripCode, reconstructs the HUT River, compares article memory against evidence, preserves session context, and returns a judge-ready diligence packet.</p>
    </div>

    <section>
      <h2>Session Identity</h2>
      <div class="grid">
        <div class="card"><b>Run ID</b><code>$RUN_ID</code></div>
        <div class="card"><b>TripCode</b><code>$TRIPCODE</code></div>
        <div class="card"><b>Session</b><code>$SESSION_ID</code></div>
        <div class="card"><b>Base URL</b><code>$BASE_URL</code></div>
      </div>
    </section>

    <section>
      <h2>Role Split And Justification</h2>
      <div class="grid">
        <div class="card">
          <b>Google Agent Layer</b>
          <p>The Go-native Cloud Run service is the judge-facing agent runtime. It owns request handling, protected demo access, session memory, cost tracking, route orchestration, and optional Gemini synthesis through Vertex AI.</p>
        </div>
        <div class="card">
          <b>ATLAS-7 MCP Layer</b>
          <p>The MCP server is the external evidence and tool boundary. It resolves the TripCode into the article object, River context, filing evidence references, caveats, thesis map, and provenance fields.</p>
        </div>
      </div>
      <p><strong>Why this split matters:</strong> the Google agent demonstrates autonomous multi-step execution and production controls, while MCP keeps proprietary evidence retrieval modular, auditable, and replaceable. The agent reasons over returned evidence; the MCP does not become the agent.</p>
    </section>

    <section>
      <h2>What Happens Step By Step</h2>
      <ol>
        <li><strong>Resolve the TripCode.</strong> The first request calls <code>GET /resolve</code> with <code>TF-SUB-9DA70A7F98</code>, issuer <code>HUT</code>, compact payload mode, filing evidence, prior articles, and thesis map enabled.</li>
        <li><strong>Recover the research object.</strong> The Go Cloud Run service calls the live ATLAS-7 MCP composite resolver and receives the structured <code>research_packet</code> for the HUT article.</li>
        <li><strong>Reconstruct the HUT River.</strong> The response includes the current article plus prior HUT River context so the agent can reason over thesis evolution instead of a single isolated post.</li>
        <li><strong>Compare memory against evidence.</strong> The packet separates article memory, filing-backed references, caveats, missing evidence, and non-advice boundaries.</li>
        <li><strong>Preserve session context.</strong> The service writes a Memory Bank entry keyed by <code>session_id</code>, including article title, TripCode, issuer, packet keys, and River node count.</li>
        <li><strong>Answer a second-turn follow-up.</strong> The second request sends a plain-text follow-up to <code>POST /resolve?session_id=$SESSION_ID</code>, proving the session can recall the previous HUT River without resolving it again.</li>
        <li><strong>Show cost discipline.</strong> The final request calls <code>GET /v1/usage</code> to expose tracked local-estimate spend and remaining challenge budget.</li>
      </ol>
    </section>

    <section>
      <h2>Files In This Evidence Session</h2>
      <ol>
        <li><code>01-resolve-tripcode-request.json</code> and <code>01-resolve-tripcode-response.json</code>: TripCode resolution request and raw response.</li>
        <li><code>02-session-followup-request.json</code> and <code>02-session-followup-response.json</code>: session-memory follow-up request and raw response.</li>
        <li><code>03-usage-request.json</code> and <code>03-usage-response.json</code>: cost ledger request and raw response.</li>
        <li><code>BUILD_PHASE_EXPLANATION.html</code>: this Build-phase explanation and machine-readable contract.</li>
      </ol>
    </section>

    <section>
      <h2>Evidence Boundary</h2>
      <p>Demo keys are always redacted from saved request files. These artifacts prove the Build workflow and local-estimate cost tracking. Exact official Google Cloud spend still requires Cloud Billing export or Billing console data.</p>
    </section>

    <section>
      <h2>Machine Contract</h2>
      <pre>&lt;StrategiXVisualSpec id="deltasignal-judge-demo-build-session" version="1.0" status="evidence-session"&gt;
  &lt;Phase&gt;Build&lt;/Phase&gt;
  &lt;BuildClaim&gt;Working demo: a Go-native Cloud Run agent resolves a TripCode, reconstructs the HUT River, compares article memory against evidence, preserves session context, and returns a judge-ready diligence packet.&lt;/BuildClaim&gt;
  &lt;Run id="$RUN_ID" tripcode="$TRIPCODE" issuer="HUT" session_id="$SESSION_ID" /&gt;
  &lt;RoleSplit&gt;
    &lt;GoogleAgent role="orchestrator"&gt;Cloud Run Go service handles judge routes, protected demo access, session memory, cost ledger, Gemini synthesis through Vertex AI, and response shaping.&lt;/GoogleAgent&gt;
    &lt;MCP role="evidence-tool-boundary"&gt;ATLAS-7 MCP resolves TripCode research packets, HUT River nodes, filing evidence refs, caveats, thesis map, and provenance fields.&lt;/MCP&gt;
    &lt;Justification&gt;The Google agent proves autonomous workflow control; MCP proves secure external context and evidence retrieval without mixing proprietary evidence storage into the agent runtime.&lt;/Justification&gt;
  &lt;/RoleSplit&gt;
  &lt;Steps&gt;
    &lt;Step order="1" route="GET /resolve"&gt;Resolve TripCode into structured research_packet.&lt;/Step&gt;
    &lt;Step order="2" system="ATLAS-7 MCP"&gt;Recover article object, HUT River nodes, filing evidence refs, caveats, and thesis map.&lt;/Step&gt;
    &lt;Step order="3" system="Memory Bank"&gt;Preserve session context for follow-up turns.&lt;/Step&gt;
    &lt;Step order="4" route="POST /resolve"&gt;Answer follow-up using prior HUT River session memory.&lt;/Step&gt;
    &lt;Step order="5" route="GET /v1/usage"&gt;Report local-estimate spend and remaining challenge budget.&lt;/Step&gt;
  &lt;/Steps&gt;
  &lt;Boundary&gt;Redacted demo key; no investment advice; official billing requires Billing export or console data.&lt;/Boundary&gt;
&lt;/StrategiXVisualSpec&gt;</pre>
    </section>
  </main>
  <script id="strategix-build-phase-contract" type="application/xml">
<StrategiXVisualSpec id="deltasignal-judge-demo-build-session" version="1.0" status="evidence-session">
  <Phase>Build</Phase>
  <BuildClaim>Working demo: a Go-native Cloud Run agent resolves a TripCode, reconstructs the HUT River, compares article memory against evidence, preserves session context, and returns a judge-ready diligence packet.</BuildClaim>
  <Run id="$RUN_ID" tripcode="$TRIPCODE" issuer="HUT" session_id="$SESSION_ID" />
  <RoleSplit>
    <GoogleAgent role="orchestrator">Cloud Run Go service handles judge routes, protected demo access, session memory, cost ledger, Gemini synthesis through Vertex AI, and response shaping.</GoogleAgent>
    <MCP role="evidence-tool-boundary">ATLAS-7 MCP resolves TripCode research packets, HUT River nodes, filing evidence refs, caveats, thesis map, and provenance fields.</MCP>
    <Justification>The Google agent proves autonomous workflow control; MCP proves secure external context and evidence retrieval without mixing proprietary evidence storage into the agent runtime.</Justification>
  </RoleSplit>
  <Steps>
    <Step order="1" route="GET /resolve">Resolve TripCode into structured research_packet.</Step>
    <Step order="2" system="ATLAS-7 MCP">Recover article object, HUT River nodes, filing evidence refs, caveats, and thesis map.</Step>
    <Step order="3" system="Memory Bank">Preserve session context for follow-up turns.</Step>
    <Step order="4" route="POST /resolve">Answer follow-up using prior HUT River session memory.</Step>
    <Step order="5" route="GET /v1/usage">Report local-estimate spend and remaining challenge budget.</Step>
  </Steps>
  <Boundary>Redacted demo key; no investment advice; official billing requires Billing export or console data.</Boundary>
</StrategiXVisualSpec>
  </script>
</body>
</html>
EOF
}

mkdir -p "$ARTIFACT_DIR"
save_build_phase_explanation

echo "DeltaSignal Gemini AI Agent judge demo"
echo "Base URL: $BASE_URL"
echo "TripCode: $TRIPCODE"
echo "Session: $SESSION_ID"
echo "Artifacts: $ARTIFACT_DIR"
echo

echo "1. Resolve TripCode into research packet and Memory Bank"
RESOLVE_URL="$BASE_URL/resolve?tripcode=$TRIPCODE&session_id=$SESSION_ID&issuer=HUT&payload_mode=compact&include_filing_evidence=true&include_prior_articles=true&include_thesis_map=true"
save_request "01-resolve-tripcode" "GET" "$RESOLVE_URL"
curl_auth "$RESOLVE_URL" | tee "$ARTIFACT_DIR/01-resolve-tripcode-response.json"
echo
echo

echo "2. Ask a follow-up that uses prior session memory"
FOLLOWUP_URL="$BASE_URL/resolve?session_id=$SESSION_ID"
FOLLOWUP_BODY="Using the previous HUT River, what are the top 3 things to monitor next and which assumptions weakened?"
save_request "02-session-followup" "POST" "$FOLLOWUP_URL" "$FOLLOWUP_BODY"
curl_auth -X POST "$FOLLOWUP_URL" \
  -H "Content-Type: text/plain" \
  --data "$FOLLOWUP_BODY" | tee "$ARTIFACT_DIR/02-session-followup-response.json"
echo
echo

echo "3. Show cost tracking after the demo requests"
USAGE_URL="$BASE_URL/v1/usage"
save_request "03-usage" "GET" "$USAGE_URL"
curl_auth "$USAGE_URL" | tee "$ARTIFACT_DIR/03-usage-response.json"
echo
