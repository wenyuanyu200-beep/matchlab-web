#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://127.0.0.1:8080}"
STAMP="$(date +%s)"
PASSWORD="MatchLab123!"

require_tool() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "missing required tool: $1" >&2
    exit 1
  }
}

api_json() {
  local method="$1"
  local path="$2"
  local token="${3:-}"
  local body="${4:-}"
  if [[ -n "$token" ]]; then
    curl -fsS -X "$method" "$BASE_URL$path" \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $token" \
      --data "$body"
  else
    curl -fsS -X "$method" "$BASE_URL$path" \
      -H "Content-Type: application/json" \
      --data "$body"
  fi
}

require_tool curl
require_tool jq

email_a="prod-a-$STAMP@example.com"
email_b="prod-b-$STAMP@example.com"

echo "health"
curl -fsS "$BASE_URL/api/health" | jq .

echo "register user A"
api_json POST /api/auth/register "" \
  "{\"email\":\"$email_a\",\"password\":\"$PASSWORD\",\"nickname\":\"Prod A\",\"school\":\"MatchLab\"}" | jq .

echo "login user A"
token_a="$(api_json POST /api/auth/login "" "{\"email\":\"$email_a\",\"password\":\"$PASSWORD\"}" | jq -r '.data.token')"

echo "register user B"
api_json POST /api/auth/register "" \
  "{\"email\":\"$email_b\",\"password\":\"$PASSWORD\",\"nickname\":\"Prod B\",\"school\":\"MatchLab\"}" | jq .

echo "login user B"
token_b="$(api_json POST /api/auth/login "" "{\"email\":\"$email_b\",\"password\":\"$PASSWORD\"}" | jq -r '.data.token')"

echo "submit questionnaire for user B"
api_json POST /api/questionnaires "$token_b" '{
  "mode": "activity",
  "answers": {
    "interests": ["AI", "backend"],
    "skills": ["Go", "PostgreSQL"],
    "available_time": "weekends",
    "activity_types": ["project"],
    "goal": "build production systems",
    "communication_style": "async"
  }
}' | jq .

echo "create activity as user A"
activity_id="$(api_json POST /api/activities "$token_a" '{
  "title": "Production Recovery Match Test",
  "type": "project",
  "description": "End-to-end deployment verification activity",
  "required_count": 2,
  "tags": ["AI", "backend"],
  "preferred_tags": ["Go", "PostgreSQL"],
  "time_text": "weekends",
  "location_text": "online"
}' | jq -r '.data.activity.id')"
echo "activity_id=$activity_id"

echo "apply activity as user B"
application_id="$(api_json POST "/api/activities/$activity_id/apply" "$token_b" \
  '{"reason":"I can help verify production recovery."}' | jq -r '.data.application.id')"
echo "application_id=$application_id"

echo "approve application as user A"
api_json POST "/api/applications/$application_id/approve" "$token_a" '{}' | jq .

echo "recommend as user B"
api_json POST /api/match/recommend "$token_b" '{"target_type":"activity","limit":10}' | jq .

echo "current matches as user B"
api_json GET /api/me/matches "$token_b" '{}' | jq .

echo "production E2E recommendation flow passed"
