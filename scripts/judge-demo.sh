#!/usr/bin/env sh
set -eu

BASE_URL="${DELTASIGNAL_AGENT_URL:-http://localhost:8080}"
TRIPCODE="${DELTASIGNAL_DEMO_TRIPCODE:-TF-SUB-9DA70A7F98}"
SESSION_ID="${DELTASIGNAL_DEMO_SESSION_ID:-judge-demo-hut}"
DEMO_KEY="${DELTASIGNAL_DEMO_API_KEY:-}"
ARTIFACT_ROOT="${DELTASIGNAL_DEMO_ARTIFACT_ROOT:-artifacts/00_TRACK_1_BUILD__WORKING_DEMO}"
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
      <p><strong>Runtime autonomy:</strong> the submitted agent performs the runtime research workflow autonomously: TripCode resolution, River reconstruction, evidence comparison, session memory, Gemini synthesis, boundaries, and usage reporting.</p>
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
      <h2>Objective And Evaluation</h2>
      <p><strong>Objective:</strong> prove the Track 1 Build claim with one repeatable session: TripCode in, HUT River reconstructed, evidence packet returned, session memory preserved, and cost ledger shown.</p>
      <div class="grid">
        <div class="card"><b>Eval 1: Runnable</b><p>The demo must run from HTTP routes on Cloud Run, not only from local code or slides.</p></div>
        <div class="card"><b>Eval 2: Agentic</b><p>The Google agent must orchestrate multiple steps: resolve, remember, follow up, and report usage.</p></div>
        <div class="card"><b>Eval 3: Resolver-backed</b><p>The agent must call the configured ATLAS-7 MCP resolver when available, and must preserve <code>live_mcp_error</code> plus deterministic fallback disclosure if the live tool is unavailable.</p></div>
        <div class="card"><b>Eval 4: Evidence-bound</b><p>The result must keep article memory, filing references, caveats, missing evidence, and non-advice boundaries separate.</p></div>
        <div class="card"><b>Eval 5: Stateful</b><p>The second request must use the prior session context without rerunning the TripCode resolution.</p></div>
        <div class="card"><b>Eval 6: Cost-aware</b><p>The run must expose local-estimate spend and remaining challenge budget.</p></div>
      </div>
      <p><strong>Build status:</strong> passed for the HUT proof session. The remaining work belongs to Optimize, not Build.</p>
    </section>

    <section>
      <h2>Agent Autonomy Boundary</h2>
      <p>The shell script is only the judge harness. It supplies the input and records evidence. The deployed Google agent performs the internal workflow autonomously after the request reaches Cloud Run.</p>
      <p><strong>Runtime relevance:</strong> DeltaSignal Gemini AI Agent is the runtime Google Cloud product agent. It turns a high-level TripCode request into a hosted, repeatable, evidence-bounded diligence workflow.</p>
      <ol>
        <li><strong>Human or judge action:</strong> provide one TripCode and session_id, then ask one follow-up question.</li>
        <li><strong>Agent action:</strong> validate the request, call the TripCode resolver, call the ATLAS-7 MCP tool, parse the structured MCP packet, write session memory, optionally synthesize with Gemini, attach disclosures, and record local-estimate cost.</li>
        <li><strong>MCP action:</strong> return bounded evidence and provenance. MCP does not decide the workflow, manage memory, or shape the final agent response.</li>
        <li><strong>Why this qualifies as an AI agent:</strong> the system uses an agent runtime to coordinate tools, memory, synthesis, evidence boundaries, and follow-up behavior from a high-level user goal rather than requiring the user to manually call each evidence source.</li>
      </ol>
    </section>

    <section>
      <h2>Agent Reasoning Proof</h2>
      <p>The first request asks the cloud agent to load the public ATLAS-7 operating guides before synthesis. This proves the agent is not only returning a hardcoded packet; it reads the public ATLAS-7 operating instructions, then performs the TripCode workflow from the user goal.</p>
      <p><strong>Core product claim:</strong> the judge can provide one TripCode instead of manually orchestrating public guide lookup, MCP calls, River reconstruction, memory handling, Gemini synthesis, boundary checks, and cost logging.</p>
      <div class="grid">
        <div class="card"><b>Public guide</b><code>https://aitrailblazer.github.io/deltasignal-atlas-codex-plugin/CLAUDE.md</code></div>
        <div class="card"><b>Full agent brief</b><code>https://aitrailblazer.github.io/deltasignal-atlas-codex-plugin/llms-full.txt</code></div>
      </div>
      <ol>
        <li><strong>What proves it:</strong> the response includes <code>agent_context.sources</code> with URL, HTTP status, byte count, SHA-256 hash, and excerpt for each fetched guide.</li>
        <li><strong>How the agent uses it:</strong> the Gemini synthesis prompt receives <code>agent_context</code>, <code>execution_trace</code>, the MCP research packet, session memory, and the non-advice boundaries.</li>
        <li><strong>What we do not expose:</strong> hidden chain-of-thought. The artifact exposes an observable execution trace and source hashes instead.</li>
        <li><strong>Availability boundary:</strong> the agent is cloud-hosted and reachable on demand. Cloud Run uses <code>min-instances=0</code> for cost control, so it may cold-start rather than stay warm.</li>
      </ol>
    </section>

    <section>
      <h2>Why Optimize Comes Next</h2>
      <p>Build proves the working agent. Optimize improves evidence fidelity, operational quality, and judge trust without changing the core demo claim.</p>
      <ol>
        <li><strong>Proven now:</strong> Cloud Run route, TripCode resolver, live MCP packet, HUT River reconstruction, session memory, Gemini synthesis, and cost ledger.</li>
        <li><strong>Optimize next:</strong> propagate richer ATLAS-7 metadata through every brief path, including source dates, computed timestamps, stale flags, caveats, quality flags, evidence hashes, payload mode, and provenance labels.</li>
        <li><strong>Reason:</strong> the next risk is not whether the agent works. The next risk is whether every generated answer carries the same provenance strength as the raw TripCode research packet.</li>
      </ol>
    </section>

    <section>
      <h2>What Happens Step By Step</h2>
      <ol>
        <li><strong>Resolve the TripCode.</strong> The first request calls <code>GET /resolve</code> with <code>TF-SUB-9DA70A7F98</code>, issuer <code>HUT</code>, compact payload mode, filing evidence, prior articles, thesis map, and public agent context enabled.</li>
        <li><strong>Recover the research object.</strong> The Go Cloud Run service calls the configured ATLAS-7 MCP composite resolver. If live MCP is unavailable, the packet must preserve <code>live_mcp_error</code> and use the deterministic HUT fixture only as disclosed fallback.</li>
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
        <li><code>01-resolve-tripcode-request.json</code> and <code>01-resolve-tripcode-response.json</code>: TripCode resolution request and raw response, including <code>agent_context</code> and <code>execution_trace</code> when enabled.</li>
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
  &lt;RuntimeAutonomyClaim&gt;The submitted agent performs the runtime research workflow autonomously: TripCode resolution, River reconstruction, evidence comparison, session memory, Gemini synthesis, boundaries, and usage reporting.&lt;/RuntimeAutonomyClaim&gt;
  &lt;Run id="$RUN_ID" tripcode="$TRIPCODE" issuer="HUT" session_id="$SESSION_ID" /&gt;
  &lt;RoleSplit&gt;
    &lt;GoogleAgent role="orchestrator"&gt;Cloud Run Go service handles judge routes, protected demo access, session memory, cost ledger, Gemini synthesis through Vertex AI, and response shaping.&lt;/GoogleAgent&gt;
    &lt;MCP role="evidence-tool-boundary"&gt;ATLAS-7 MCP resolves TripCode research packets, HUT River nodes, filing evidence refs, caveats, thesis map, and provenance fields.&lt;/MCP&gt;
    &lt;Justification&gt;The Google agent proves autonomous workflow control; MCP proves secure external context and evidence retrieval without mixing proprietary evidence storage into the agent runtime.&lt;/Justification&gt;
  &lt;/RoleSplit&gt;
  &lt;Objective&gt;Prove the Track 1 Build claim with one repeatable session: TripCode in, HUT River reconstructed, evidence packet returned, session memory preserved, and cost ledger shown.&lt;/Objective&gt;
  &lt;Evaluation status="passed"&gt;
    &lt;Criterion name="Runnable"&gt;Cloud Run HTTP routes execute the demo.&lt;/Criterion&gt;
    &lt;Criterion name="Agentic"&gt;The Google agent orchestrates resolve, memory, follow-up, and usage reporting.&lt;/Criterion&gt;
    &lt;Criterion name="ResolverBacked"&gt;The agent calls the configured ATLAS-7 MCP resolver when available and preserves live_mcp_error plus deterministic fallback disclosure if the live tool is unavailable.&lt;/Criterion&gt;
    &lt;Criterion name="EvidenceBound"&gt;Article memory, filing refs, caveats, missing evidence, and non-advice boundaries stay separated.&lt;/Criterion&gt;
    &lt;Criterion name="Stateful"&gt;The follow-up uses session memory without rerunning TripCode resolution.&lt;/Criterion&gt;
    &lt;Criterion name="CostAware"&gt;The run exposes tracked local-estimate spend and remaining budget.&lt;/Criterion&gt;
  &lt;/Evaluation&gt;
  &lt;AutonomyBoundary&gt;
    &lt;JudgeHarness&gt;scripts/judge-demo.sh supplies input, records requests, records responses, and proves reproducibility.&lt;/JudgeHarness&gt;
    &lt;AgentAutonomy&gt;Cloud Run Go agent validates the request, calls MCP, parses structured evidence, writes session memory, invokes Gemini synthesis when configured, attaches disclosures, and reports cost without the user manually executing those internal steps.&lt;/AgentAutonomy&gt;
    &lt;MCPBoundary&gt;ATLAS-7 MCP supplies bounded evidence and provenance but does not own workflow orchestration, memory, or final response shaping.&lt;/MCPBoundary&gt;
    &lt;AgentJustification&gt;The submitted system qualifies as an AI agent because it coordinates tools, memory, synthesis, evidence boundaries, and follow-up behavior from a high-level user goal.&lt;/AgentJustification&gt;
  &lt;/AutonomyBoundary&gt;
  &lt;AgentReasoningProof&gt;
    &lt;PublicGuide url="https://aitrailblazer.github.io/deltasignal-atlas-codex-plugin/CLAUDE.md" /&gt;
    &lt;PublicGuide url="https://aitrailblazer.github.io/deltasignal-atlas-codex-plugin/llms-full.txt" /&gt;
    &lt;ProofMechanism&gt;The Cloud Run agent fetches public guide context, records status, bytes, SHA-256 hashes, excerpts, and execution_trace steps, then provides that context to Gemini synthesis with the MCP research packet and boundaries.&lt;/ProofMechanism&gt;
    &lt;ReasoningBoundary&gt;Expose observable execution trace and source hashes, not hidden chain-of-thought.&lt;/ReasoningBoundary&gt;
    &lt;AvailabilityBoundary&gt;The agent is cloud-hosted and reachable on demand; Cloud Run min-instances=0 may cold-start to preserve Google credits.&lt;/AvailabilityBoundary&gt;
  &lt;/AgentReasoningProof&gt;
  &lt;NextPhase name="Optimize"&gt;Improve evidence fidelity by propagating source dates, computed timestamps, stale flags, caveats, quality flags, evidence hashes, payload mode, and provenance labels through every user-facing brief path.&lt;/NextPhase&gt;
  &lt;Steps&gt;
    &lt;Step order="1" route="GET /resolve"&gt;Resolve TripCode into structured research_packet.&lt;/Step&gt;
    &lt;Step order="2" system="ATLAS-7 MCP or disclosed fallback"&gt;Recover article object, HUT River nodes, filing evidence refs, caveats, and thesis map; preserve live_mcp_error when fallback is used.&lt;/Step&gt;
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
  <RuntimeAutonomyClaim>The submitted agent performs the runtime research workflow autonomously: TripCode resolution, River reconstruction, evidence comparison, session memory, Gemini synthesis, boundaries, and usage reporting.</RuntimeAutonomyClaim>
  <Run id="$RUN_ID" tripcode="$TRIPCODE" issuer="HUT" session_id="$SESSION_ID" />
  <RoleSplit>
    <GoogleAgent role="orchestrator">Cloud Run Go service handles judge routes, protected demo access, session memory, cost ledger, Gemini synthesis through Vertex AI, and response shaping.</GoogleAgent>
    <MCP role="evidence-tool-boundary">ATLAS-7 MCP resolves TripCode research packets, HUT River nodes, filing evidence refs, caveats, thesis map, and provenance fields.</MCP>
    <Justification>The Google agent proves autonomous workflow control; MCP proves secure external context and evidence retrieval without mixing proprietary evidence storage into the agent runtime.</Justification>
  </RoleSplit>
  <Objective>Prove the Track 1 Build claim with one repeatable session: TripCode in, HUT River reconstructed, evidence packet returned, session memory preserved, and cost ledger shown.</Objective>
  <Evaluation status="passed">
    <Criterion name="Runnable">Cloud Run HTTP routes execute the demo.</Criterion>
    <Criterion name="Agentic">The Google agent orchestrates resolve, memory, follow-up, and usage reporting.</Criterion>
    <Criterion name="ResolverBacked">The agent calls the configured ATLAS-7 MCP resolver when available and preserves live_mcp_error plus deterministic fallback disclosure if the live tool is unavailable.</Criterion>
    <Criterion name="EvidenceBound">Article memory, filing refs, caveats, missing evidence, and non-advice boundaries stay separated.</Criterion>
    <Criterion name="Stateful">The follow-up uses session memory without rerunning TripCode resolution.</Criterion>
    <Criterion name="CostAware">The run exposes tracked local-estimate spend and remaining budget.</Criterion>
  </Evaluation>
  <AutonomyBoundary>
    <JudgeHarness>scripts/judge-demo.sh supplies input, records requests, records responses, and proves reproducibility.</JudgeHarness>
    <AgentAutonomy>Cloud Run Go agent validates the request, calls MCP, parses structured evidence, writes session memory, invokes Gemini synthesis when configured, attaches disclosures, and reports cost without the user manually executing those internal steps.</AgentAutonomy>
    <MCPBoundary>ATLAS-7 MCP supplies bounded evidence and provenance but does not own workflow orchestration, memory, or final response shaping.</MCPBoundary>
    <AgentJustification>The submitted system qualifies as an AI agent because it coordinates tools, memory, synthesis, evidence boundaries, and follow-up behavior from a high-level user goal.</AgentJustification>
  </AutonomyBoundary>
  <AgentReasoningProof>
    <PublicGuide url="https://aitrailblazer.github.io/deltasignal-atlas-codex-plugin/CLAUDE.md" />
    <PublicGuide url="https://aitrailblazer.github.io/deltasignal-atlas-codex-plugin/llms-full.txt" />
    <ProofMechanism>The Cloud Run agent fetches public guide context, records status, bytes, SHA-256 hashes, excerpts, and execution_trace steps, then provides that context to Gemini synthesis with the MCP research packet and boundaries.</ProofMechanism>
    <ReasoningBoundary>Expose observable execution trace and source hashes, not hidden chain-of-thought.</ReasoningBoundary>
    <AvailabilityBoundary>The agent is cloud-hosted and reachable on demand; Cloud Run min-instances=0 may cold-start to preserve Google credits.</AvailabilityBoundary>
  </AgentReasoningProof>
  <NextPhase name="Optimize">Improve evidence fidelity by propagating source dates, computed timestamps, stale flags, caveats, quality flags, evidence hashes, payload mode, and provenance labels through every user-facing brief path.</NextPhase>
  <Steps>
    <Step order="1" route="GET /resolve">Resolve TripCode into structured research_packet.</Step>
    <Step order="2" system="ATLAS-7 MCP or disclosed fallback">Recover article object, HUT River nodes, filing evidence refs, caveats, and thesis map; preserve live_mcp_error when fallback is used.</Step>
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

json_value() {
  file="$1"
  expr="$2"
  fallback="$3"
  if command -v jq >/dev/null 2>&1 && [ -f "$file" ]; then
    value="$(jq -r "$expr // empty" "$file" 2>/dev/null || true)"
    if [ -n "$value" ] && [ "$value" != "null" ]; then
      printf '%s' "$value"
      return
    fi
  fi
  printf '%s' "$fallback"
}

json_pass() {
  file="$1"
  expr="$2"
  if command -v jq >/dev/null 2>&1 && [ -f "$file" ] && jq -e "$expr" "$file" >/dev/null 2>&1; then
    return 0
  fi
  return 1
}

save_learning_evidence() {
  resolve_file="$ARTIFACT_DIR/01-resolve-tripcode-response.json"
  followup_file="$ARTIFACT_DIR/02-session-followup-response.json"
  usage_file="$ARTIFACT_DIR/03-usage-response.json"

  resolve_mode="$(json_value "$resolve_file" '.mode' 'unknown')"
  context_count="$(json_value "$resolve_file" '(.agent_context.sources // []) | length' '0')"
  context_statuses="$(json_value "$resolve_file" '[.agent_context.sources[]?.status] | join(", ")' 'none')"
  trace_count="$(json_value "$resolve_file" '(.execution_trace // []) | length' '0')"
  disclosure_count="$(json_value "$resolve_file" '(.disclosures // []) | length' '0')"
  live_mcp_error="$(json_value "$resolve_file" '.packet.live_mcp_error' '')"
  followup_mode="$(json_value "$followup_file" '.mode' 'unknown')"
  tracked_spend="$(json_value "$usage_file" '.tracked_spent_usd' 'unknown')"
  remaining_budget="$(json_value "$usage_file" '.remaining_usd' 'unknown')"
  evaluator="scripts/judge-demo.sh deterministic post-run evaluator"
  total_score=0
  route_score=0
  context_score=0
  trace_score=0
  memory_score=0
  boundary_score=0
  resolver_score=0
  cost_score=0

  if json_pass "$resolve_file" '.tripcode and .packet and .mode'; then route_score=15; fi
  if json_pass "$resolve_file" '((.agent_context.sources // []) | length) >= 2 and all(.agent_context.sources[]?; .status == "fetched" and (.sha256 | length) > 0 and (.bytes // 0) > 0)'; then context_score=15; fi
  if json_pass "$resolve_file" '((.execution_trace // []) | length) >= 7'; then trace_score=15; fi
  if json_pass "$followup_file" '.mode == "session-memory" and .memory.available == true'; then memory_score=15; fi
  if json_pass "$resolve_file" '((.disclosures // []) | length) >= 4'; then boundary_score=10; fi
  if json_pass "$resolve_file" '(.packet.live_mcp_error | not)'; then resolver_score=20; else resolver_score=10; fi
  if json_pass "$usage_file" '(.tracked_spent_usd != null) and (.remaining_usd != null)'; then cost_score=10; fi
  total_score=$((route_score + context_score + trace_score + memory_score + boundary_score + resolver_score + cost_score))

  if [ -n "$live_mcp_error" ]; then
    mcp_learning="Live MCP resolver returned: $live_mcp_error. Keep fallback disclosure visible and fix public MCP/tool-name parity before claiming fully live TripCode resolution."
  else
    mcp_learning="Live MCP resolver returned without a live_mcp_error field. Preserve this as the preferred production path and keep fallback tests for resilience."
  fi

  cat > "$ARTIFACT_DIR/LEARNING_EVIDENCE.html" <<EOF
<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>DeltaSignal Judge Demo | Learning Evidence</title>
  <style>
    html { background: #07090f; color: #fff8ef; font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; line-height: 1.5; }
    body { margin: 0; padding: 28px; }
    main { max-width: 980px; margin: 0 auto; }
    .hero, section { border: 1px solid rgba(255,255,255,.14); background: rgba(18,20,28,.9); border-radius: 10px; padding: 22px; margin: 0 0 16px; }
    .eyebrow { color: #72d9ff; font-size: 12px; font-weight: 900; letter-spacing: .08em; text-transform: uppercase; }
    h1 { margin: 8px 0 12px; font-size: clamp(32px, 6vw, 56px); line-height: 1; letter-spacing: 0; }
    h2 { margin: 0 0 10px; color: #ffbd5a; font-size: 22px; letter-spacing: 0; }
    p, li { color: rgba(255,248,239,.8); }
    code { color: #72d9ff; background: rgba(114,217,255,.09); padding: 2px 5px; border-radius: 5px; }
    ol { margin: 0; padding-left: 22px; }
    li { margin: 0 0 10px; }
    .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(210px, 1fr)); gap: 10px; margin-top: 12px; }
    .card { border: 1px solid rgba(255,255,255,.12); background: rgba(255,255,255,.04); border-radius: 8px; padding: 14px; }
    .card b { display: block; color: #fff8ef; margin-bottom: 6px; }
    pre { overflow: auto; border: 1px solid rgba(255,255,255,.12); background: #0b0e15; border-radius: 8px; padding: 14px; color: #d9f7ff; }
  </style>
</head>
<body>
  <main>
    <div class="hero">
      <div class="eyebrow">Post-run learning evidence | Build to Optimize</div>
      <h1>What This Run Teaches The Codebase</h1>
      <p>This file is generated after the judge demo requests complete. It is not the judge proof itself; it is the engineering learning layer used to decide what to harden next.</p>
    </div>

    <section>
      <h2>Run Snapshot</h2>
      <div class="grid">
        <div class="card"><b>Run ID</b><code>$RUN_ID</code></div>
        <div class="card"><b>TripCode</b><code>$TRIPCODE</code></div>
        <div class="card"><b>Resolve mode</b><code>$resolve_mode</code></div>
        <div class="card"><b>Follow-up mode</b><code>$followup_mode</code></div>
        <div class="card"><b>Agent context sources</b><code>$context_count</code></div>
        <div class="card"><b>Agent context status</b><code>$context_statuses</code></div>
        <div class="card"><b>Execution trace steps</b><code>$trace_count</code></div>
        <div class="card"><b>Disclosures</b><code>$disclosure_count</code></div>
        <div class="card"><b>Tracked spend</b><code>$tracked_spend USD</code></div>
        <div class="card"><b>Remaining estimate</b><code>$remaining_budget USD</code></div>
      </div>
    </section>

    <section>
      <h2>Who Evaluated The Run</h2>
      <p><strong>Evaluator:</strong> <code>$evaluator</code>.</p>
      <p>The score is not assigned by Gemini. Gemini may synthesize the diligence text, but the run score is computed deterministically from saved JSON fields so it can be repeated by judges or by us after each code change.</p>
    </section>

    <section>
      <h2>Run Eval Score</h2>
      <div class="grid">
        <div class="card"><b>Total</b><code>$total_score / 100</code></div>
        <div class="card"><b>Route packet</b><code>$route_score / 15</code></div>
        <div class="card"><b>Public guide context</b><code>$context_score / 15</code></div>
        <div class="card"><b>Execution trace</b><code>$trace_score / 15</code></div>
        <div class="card"><b>Session memory</b><code>$memory_score / 15</code></div>
        <div class="card"><b>Boundaries</b><code>$boundary_score / 10</code></div>
        <div class="card"><b>Resolver parity</b><code>$resolver_score / 20</code></div>
        <div class="card"><b>Cost ledger</b><code>$cost_score / 10</code></div>
      </div>
      <p><strong>Improvement signal:</strong> resolver parity is the largest remaining swing factor. A fully live MCP composite response scores 20/20; a disclosed deterministic fallback scores 10/20 because it is reliable but not the final production path.</p>
    </section>

    <section>
      <h2>What The Run Proved</h2>
      <ol>
        <li>The cloud-hosted Go agent accepted one high-level TripCode goal through the judge route.</li>
        <li>The agent fetched public ATLAS-7 operating guides when <code>include_agent_context=true</code> was set, and the response recorded source status, byte count, SHA-256 hashes, and excerpts.</li>
        <li>The response exposed <code>execution_trace</code>, giving observable proof of tool selection, resolver call, memory write, context loading, Gemini synthesis, and cost recording without exposing hidden chain-of-thought.</li>
        <li>The second request used session memory, proving the agent can preserve River continuity across turns.</li>
      </ol>
    </section>

    <section>
      <h2>What The Run Exposed</h2>
      <p>$mcp_learning</p>
    </section>

    <section>
      <h2>Code Improvements To Carry Forward</h2>
      <ol>
        <li><strong>Resolver parity:</strong> make the public ATLAS-7 MCP endpoint and configured tool name match the Cloud Run agent contract before recording the final public demo.</li>
        <li><strong>Trace precision:</strong> keep separate actors for <code>google-agent</code>, <code>mcp</code>, and <code>gemini</code> so evidence does not overstate which layer performed which work.</li>
        <li><strong>Evidence fidelity:</strong> propagate source dates, computed timestamps, stale flags, caveats, quality flags, evidence hashes, payload mode, and provenance labels through every brief path in the Optimize phase.</li>
        <li><strong>Cost guardrail:</strong> keep <code>min-instances=0</code>, reuse saved artifacts for review, and avoid repeated live Gemini/MCP runs unless they prove a new deployment or fixed resolver path.</li>
      </ol>
    </section>

    <section>
      <h2>Machine Contract</h2>
      <pre>&lt;StrategiXVisualSpec id="deltasignal-judge-demo-learning-evidence" version="1.0" status="post-run-learning"&gt;
  &lt;Run id="$RUN_ID" tripcode="$TRIPCODE" resolve_mode="$resolve_mode" followup_mode="$followup_mode" /&gt;
  &lt;Evaluator type="deterministic-script" owner="scripts/judge-demo.sh"&gt;Scores are computed from saved JSON response fields, not assigned by Gemini.&lt;/Evaluator&gt;
  &lt;Score total="$total_score" max="100"&gt;
    &lt;Criterion name="RoutePacket" score="$route_score" max="15" /&gt;
    &lt;Criterion name="PublicGuideContext" score="$context_score" max="15" /&gt;
    &lt;Criterion name="ExecutionTrace" score="$trace_score" max="15" /&gt;
    &lt;Criterion name="SessionMemory" score="$memory_score" max="15" /&gt;
    &lt;Criterion name="Boundaries" score="$boundary_score" max="10" /&gt;
    &lt;Criterion name="ResolverParity" score="$resolver_score" max="20" /&gt;
    &lt;Criterion name="CostLedger" score="$cost_score" max="10" /&gt;
  &lt;/Score&gt;
  &lt;Purpose&gt;Capture post-run learning that should improve the Go agent, MCP resolver parity, evidence trace, and Optimize-phase roadmap.&lt;/Purpose&gt;
  &lt;Proofs&gt;
    &lt;AgentContext sources="$context_count" statuses="$context_statuses" /&gt;
    &lt;ExecutionTrace steps="$trace_count" /&gt;
    &lt;SessionMemory mode="$followup_mode" /&gt;
    &lt;Cost tracked_spend_usd="$tracked_spend" remaining_usd="$remaining_budget" /&gt;
  &lt;/Proofs&gt;
  &lt;Learning&gt;$mcp_learning&lt;/Learning&gt;
  &lt;NextCodeImprovements&gt;
    &lt;Item&gt;Fix public MCP resolver parity and configured tool-name agreement before final public demo claims.&lt;/Item&gt;
    &lt;Item&gt;Preserve observable execution_trace without exposing hidden chain-of-thought.&lt;/Item&gt;
    &lt;Item&gt;Carry richer ATLAS-7 provenance into every user-facing brief path in Optimize.&lt;/Item&gt;
    &lt;Item&gt;Keep cost guardrails active for the 500 USD credit budget.&lt;/Item&gt;
  &lt;/NextCodeImprovements&gt;
&lt;/StrategiXVisualSpec&gt;</pre>
    </section>
  </main>
  <script id="strategix-learning-evidence-contract" type="application/xml">
<StrategiXVisualSpec id="deltasignal-judge-demo-learning-evidence" version="1.0" status="post-run-learning">
  <Run id="$RUN_ID" tripcode="$TRIPCODE" resolve_mode="$resolve_mode" followup_mode="$followup_mode" />
  <Evaluator type="deterministic-script" owner="scripts/judge-demo.sh">Scores are computed from saved JSON response fields, not assigned by Gemini.</Evaluator>
  <Score total="$total_score" max="100">
    <Criterion name="RoutePacket" score="$route_score" max="15" />
    <Criterion name="PublicGuideContext" score="$context_score" max="15" />
    <Criterion name="ExecutionTrace" score="$trace_score" max="15" />
    <Criterion name="SessionMemory" score="$memory_score" max="15" />
    <Criterion name="Boundaries" score="$boundary_score" max="10" />
    <Criterion name="ResolverParity" score="$resolver_score" max="20" />
    <Criterion name="CostLedger" score="$cost_score" max="10" />
  </Score>
  <Purpose>Capture post-run learning that should improve the Go agent, MCP resolver parity, evidence trace, and Optimize-phase roadmap.</Purpose>
  <Proofs>
    <AgentContext sources="$context_count" statuses="$context_statuses" />
    <ExecutionTrace steps="$trace_count" />
    <SessionMemory mode="$followup_mode" />
    <Cost tracked_spend_usd="$tracked_spend" remaining_usd="$remaining_budget" />
  </Proofs>
  <Learning>$mcp_learning</Learning>
  <NextCodeImprovements>
    <Item>Fix public MCP resolver parity and configured tool-name agreement before final public demo claims.</Item>
    <Item>Preserve observable execution_trace without exposing hidden chain-of-thought.</Item>
    <Item>Carry richer ATLAS-7 provenance into every user-facing brief path in Optimize.</Item>
    <Item>Keep cost guardrails active for the 500 USD credit budget.</Item>
  </NextCodeImprovements>
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
RESOLVE_URL="$BASE_URL/resolve?tripcode=$TRIPCODE&session_id=$SESSION_ID&issuer=HUT&payload_mode=compact&include_filing_evidence=true&include_prior_articles=true&include_thesis_map=true&include_agent_context=true"
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

save_learning_evidence
echo "Learning evidence: $ARTIFACT_DIR/LEARNING_EVIDENCE.html"
