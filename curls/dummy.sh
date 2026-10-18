#!/usr/bin/env bash
# GET /dummy
set -euo pipefail
source "$(dirname "$0")/env.sh"

curl -sS -X GET "$BASE_URL/dummy"
echo
