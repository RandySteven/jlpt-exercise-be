#!/usr/bin/env bash
# POST /auth/oauth2/url
set -euo pipefail
source "$(dirname "$0")/env.sh"

curl -sS -X POST "$BASE_URL/auth/oauth2/url" \
  -H "Content-Type: application/json" \
  -d '{
    "idempotency_key": "oauth-key-1"
  }'
echo
