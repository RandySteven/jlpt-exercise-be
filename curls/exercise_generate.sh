#!/usr/bin/env bash
# POST /exercise/generate
set -euo pipefail
source "$(dirname "$0")/env.sh"

if [[ -z "$TOKEN" ]]; then
  echo "Set TOKEN to the access token from auth_login.sh" >&2
  exit 1
fi

curl -sS -X POST "$BASE_URL/exercise/generate" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "idempotency_key": "exercise-key-1",
    "jlpt_n_level": 3
  }'
echo
