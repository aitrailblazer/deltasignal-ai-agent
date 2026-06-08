#!/usr/bin/env sh
set -eu

BASE_URL="${DELTASIGNAL_AGENT_URL:-http://localhost:8080}"
TRIPCODE="${DELTASIGNAL_DEMO_TRIPCODE:-TF-SUB-9DA70A7F98}"
SESSION_ID="${DELTASIGNAL_DEMO_SESSION_ID:-judge-demo-hut}"
DEMO_KEY="${DELTASIGNAL_DEMO_API_KEY:-}"

curl_auth() {
  if [ -n "$DEMO_KEY" ]; then
    curl -sS -H "X-Demo-Key: $DEMO_KEY" "$@"
  else
    curl -sS "$@"
  fi
}

echo "DeltaSignal Gemini AI Agent judge demo"
echo "Base URL: $BASE_URL"
echo "TripCode: $TRIPCODE"
echo "Session: $SESSION_ID"
echo

echo "1. Resolve TripCode into research packet and Memory Bank"
curl_auth "$BASE_URL/resolve?tripcode=$TRIPCODE&session_id=$SESSION_ID&issuer=HUT&payload_mode=compact&include_filing_evidence=true&include_prior_articles=true&include_thesis_map=true"
echo
echo

echo "2. Ask a follow-up that uses prior session memory"
curl_auth -X POST "$BASE_URL/resolve?session_id=$SESSION_ID" \
  -H "Content-Type: text/plain" \
  --data "Using the previous HUT River, what are the top 3 things to monitor next and which assumptions weakened?"
echo
echo

echo "3. Show cost tracking after the demo requests"
curl_auth "$BASE_URL/v1/usage"
echo
