#!/usr/bin/env bash
set -euo pipefail

PROJECT="${PROJECT:-${GOOGLE_CLOUD_PROJECT:-${CLOUDSDK_CORE_PROJECT:-startup-ai-deltasignal}}}"
BASE_URL="${DELTASIGNAL_AGENT_URL:-http://localhost:8080}"
DEMO_KEY="${DELTASIGNAL_DEMO_API_KEY:-}"

echo "== DeltaSignal spend report =="
echo "project: ${PROJECT}"
echo "agent_url: ${BASE_URL}"

if command -v gcloud >/dev/null 2>&1; then
  echo
  echo "-- gcloud billing project link --"
  if ! gcloud billing projects describe "${PROJECT}" --format='value(billingEnabled,billingAccountName)' 2>/tmp/deltasignal_gcloud_billing_err.txt; then
    echo "billing project link unavailable"
    sed 's/^/  /' /tmp/deltasignal_gcloud_billing_err.txt || true
  fi
else
  echo
  echo "-- gcloud billing project link --"
  echo "gcloud unavailable"
fi

echo
echo "-- app usage ledger --"
headers=()
if [[ -n "${DEMO_KEY}" ]]; then
  headers=(-H "X-Demo-Key: ${DEMO_KEY}")
fi

if command -v curl >/dev/null 2>&1; then
  if ! curl -fsS "${headers[@]}" "${BASE_URL}/v1/usage"; then
    echo
    echo "usage endpoint unavailable; start the app or set DELTASIGNAL_AGENT_URL"
  fi
else
  echo "curl unavailable"
fi

echo
echo
echo "note: app usage is local-estimate. Exact Google spend requires Cloud Billing export or the Billing console."
