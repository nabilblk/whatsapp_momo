#!/bin/bash

# WhatsApp Promo Code API Test Script
# Usage: ./test_api.sh [API_URL]

API_URL="${1:-http://localhost:8080}"

echo "Testing API at: $API_URL"
echo "========================================"
echo ""

# Test health endpoint
echo "1. Testing health endpoint..."
curl -s "$API_URL/api/v1/health" | jq .
echo ""

# Test valid code
echo "2. Testing VALID100 (should succeed with 1GB data)..."
curl -s -X POST "$API_URL/api/v1/redeem" \
  -H "Content-Type: application/json" \
  -d '{
    "phone_number": "+237600000001",
    "code": "VALID100",
    "language": "fr"
  }' | jq .
echo ""

# Test valid code (English)
echo "3. Testing VALID200 in English (should succeed with 500 FCFA)..."
curl -s -X POST "$API_URL/api/v1/redeem" \
  -H "Content-Type: application/json" \
  -d '{
    "phone_number": "+237600000002",
    "code": "VALID200",
    "language": "en"
  }' | jq .
echo ""

# Test expired code
echo "4. Testing EXPIRED01 (should fail - expired)..."
curl -s -X POST "$API_URL/api/v1/redeem" \
  -H "Content-Type: application/json" \
  -d '{
    "phone_number": "+237600000003",
    "code": "EXPIRED01",
    "language": "fr"
  }' | jq .
echo ""

# Test used code
echo "5. Testing USED001 (should fail - already used)..."
curl -s -X POST "$API_URL/api/v1/redeem" \
  -H "Content-Type: application/json" \
  -d '{
    "phone_number": "+237600000004",
    "code": "USED001",
    "language": "fr"
  }' | jq .
echo ""

# Test invalid code
echo "6. Testing INVALID (should fail - invalid)..."
curl -s -X POST "$API_URL/api/v1/redeem" \
  -H "Content-Type: application/json" \
  -d '{
    "phone_number": "+237600000005",
    "code": "INVALID",
    "language": "fr"
  }' | jq .
echo ""

# Test unknown code
echo "7. Testing UNKNOWN123 (should fail - not found)..."
curl -s -X POST "$API_URL/api/v1/redeem" \
  -H "Content-Type: application/json" \
  -d '{
    "phone_number": "+237600000006",
    "code": "UNKNOWN123",
    "language": "fr"
  }' | jq .
echo ""

# Test rate limiting (send 6 requests rapidly)
echo "8. Testing rate limiting (6 rapid requests)..."
for i in {1..6}; do
  echo "  Request $i:"
  curl -s -X POST "$API_URL/api/v1/redeem" \
    -H "Content-Type: application/json" \
    -d '{
      "phone_number": "+237600000007",
      "code": "VALID100",
      "language": "fr"
    }' | jq -c '{status, error_code}'
done
echo ""

echo "========================================"
echo "All tests completed!"
