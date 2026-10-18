#!/usr/bin/env bash
# POST /auth/login
set -euo pipefail
source "$(dirname "$0")/env.sh"

curl -sS -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "randy@example.com",
    "password": "password123"
  }'
echo
