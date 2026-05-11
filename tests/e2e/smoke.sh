#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8085}"
unique_suffix="$(date +%s)"

echo "Check health"
curl -fsS "$BASE_URL/health" >/dev/null

echo "Register user"
TOKEN=$(
  curl -fsS -X POST "$BASE_URL/register" \
    -H "Content-Type: application/json" \
    -d "{
      \"email\": \"e2e-$unique_suffix@example.com\",
      \"username\": \"e2e_user_$unique_suffix\",
      \"password\": \"strong-password\",
      \"firstName\": \"E2E\",
      \"lastName\": \"User\"
    }" | sed -n 's/.*"accessToken":"\([^"]*\)".*/\1/p'
)

if [ -z "$TOKEN" ]; then
  echo "FAIL: access token is empty"
  exit 1
fi

echo "Create account"
ACCOUNT_ID=$(
  curl -fsS -X POST "$BASE_URL/accounts" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d '{"accountType":"current"}' | sed -n 's/.*"id":"\([^"]*\)".*/\1/p'
)

if [ -z "$ACCOUNT_ID" ]; then
  echo "FAIL: account id is empty"
  exit 1
fi

echo "Deposit"
curl -fsS -X POST "$BASE_URL/accounts/$ACCOUNT_ID/deposit" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"amount\": \"1000.00\",
    \"description\": \"E2E deposit\",
    \"idempotencyKey\": \"e2e-deposit-$unique_suffix\"
  }" >/dev/null

echo "Issue card"
CARD_RESPONSE=$(
  curl -fsS -X POST "$BASE_URL/cards" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d "{
      \"accountId\": \"$ACCOUNT_ID\",
      \"alias\": \"E2E card\"
    }"
)

CARD_ID=$(echo "$CARD_RESPONSE" | sed -n 's/.*"id":"\([^"]*\)".*/\1/p')

if [ -z "$CARD_ID" ]; then
  echo "FAIL: card id is empty"
  echo "$CARD_RESPONSE"
  exit 1
fi

echo "Create MFA challenge for credit"
MFA_RESPONSE=$(
  curl -fsS -X POST "$BASE_URL/mfa/challenges" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d '{
      "purpose": "credit_issue",
      "context": {
        "source": "e2e"
      }
    }'
)

MFA_CHALLENGE_ID=$(echo "$MFA_RESPONSE" | sed -n 's/.*"id":"\([^"]*\)".*/\1/p')

if [ -z "$MFA_CHALLENGE_ID" ]; then
  echo "FAIL: mfa challenge id is empty"
  echo "$MFA_RESPONSE"
  exit 1
fi

echo "Read MFA code from database"
MFA_CODE=$(
  docker exec banking-postgres psql -U banking_user -d banking_service -Atc "
    SELECT payload->>'code'
    FROM email_outbox
    WHERE template_code = 'mfa_code'
    ORDER BY created_at DESC
    LIMIT 1;
  "
)

if [ -z "$MFA_CODE" ]; then
  echo "FAIL: MFA code is empty"
  exit 1
fi

echo "Verify MFA"
curl -fsS -X POST "$BASE_URL/mfa/challenges/$MFA_CHALLENGE_ID/verify" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"code\": \"$MFA_CODE\"
  }" >/dev/null

echo "Issue credit"
CREDIT_RESPONSE=$(
  curl -fsS -X POST "$BASE_URL/credits" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -H "X-MFA-Challenge-ID: $MFA_CHALLENGE_ID" \
    -d "{
      \"disbursementAccountId\": \"$ACCOUNT_ID\",
      \"repaymentAccountId\": \"$ACCOUNT_ID\",
      \"principalAmount\": \"10000.00\",
      \"termMonths\": 3,
      \"bankMargin\": \"4.0000\"
    }"
)

CREDIT_ID=$(echo "$CREDIT_RESPONSE" | sed -n 's/.*"credit":{"id":"\([^"]*\)".*/\1/p')

if [ -z "$CREDIT_ID" ]; then
  echo "FAIL: credit id is empty"
  echo "$CREDIT_RESPONSE"
  exit 1
fi

echo "Get credit schedule"
curl -fsS "$BASE_URL/credits/$CREDIT_ID/schedule" \
  -H "Authorization: Bearer $TOKEN" >/dev/null

echo "Get analytics"
curl -fsS "$BASE_URL/analytics" \
  -H "Authorization: Bearer $TOKEN" >/dev/null

echo "Predict balance"
curl -fsS "$BASE_URL/accounts/$ACCOUNT_ID/predict?days=30" \
  -H "Authorization: Bearer $TOKEN" >/dev/null

echo "PASS: E2E smoke test completed"
