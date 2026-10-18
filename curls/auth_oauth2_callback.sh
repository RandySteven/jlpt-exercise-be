#!/usr/bin/env bash
# POST /auth/oauth2/callback
set -euo pipefail
source "$(dirname "$0")/env.sh"

curl -sS -X POST "$BASE_URL/auth/oauth2/callback" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "oauth-code"
  }'
echo
