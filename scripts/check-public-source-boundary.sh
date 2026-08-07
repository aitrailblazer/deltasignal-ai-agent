#!/usr/bin/env bash
set -euo pipefail

readonly_paths=(
  "Dockerfile"
  "cloudbuild.yaml"
  "client-demo/google-cloud-aapl/.env.example"
  "client-demo/google-cloud-aapl/Dockerfile"
  "client-demo/google-cloud-aapl/cloudbuild.yaml"
  "cmd/aapl-client-demo-server/main.go"
  "internal/aapldemo/gemini.go"
  "internal/agent/gemini.go"
  "internal/agent/gemini_test.go"
)

failed=0
for path in "${readonly_paths[@]}"; do
  if git ls-files --error-unmatch "$path" >/dev/null 2>&1; then
    printf 'public-source-boundary: tracked private runtime path: %s\n' "$path" >&2
    failed=1
  fi
done

if ((failed)); then
  exit 1
fi

printf 'public-source-boundary: PASS\n'
