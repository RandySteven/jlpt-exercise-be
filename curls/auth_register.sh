#!/usr/bin/env bash
# POST /auth/register
set -euo pipefail
source "$(dirname "$0")/env.sh"

curl -sS -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Randy",
    "last_name": "Steven",
    "email": "randy@example.com",
    "username": "randy",
    "password": "password123",
    "phone_number": "081234567890",
    "address": "Jakarta",
    "date_of_birth": "1995-01-01",
    "additional_info": {}
  }'
echo
