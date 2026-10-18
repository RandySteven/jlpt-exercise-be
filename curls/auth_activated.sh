#!/usr/bin/env bash
# POST /auth/activated
set -euo pipefail
source "$(dirname "$0")/env.sh"

curl -sS -X POST "$BASE_URL/auth/activated" \
  -H "Content-Type: application/json" \
  -d '{
    "token": "activation-token"
  }'
echo
