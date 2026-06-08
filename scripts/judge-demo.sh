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

mkdir -p "$ARTIFACT_DIR"

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
