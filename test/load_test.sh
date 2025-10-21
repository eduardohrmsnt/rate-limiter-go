#!/bin/bash

echo "================================"
echo "Rate Limiter Load Test"
echo "================================"
echo ""

BASE_URL="http://localhost:8080/test"

echo "Test 1: IP Rate Limiting (should allow 5, then block)"
echo "------------------------------------------------------"
for i in {1..7}; do
  RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" $BASE_URL)
  echo "Request $i: HTTP $RESPONSE"
  
  if [ $i -le 5 ]; then
    if [ "$RESPONSE" != "200" ]; then
      echo "❌ FAIL: Request $i should return 200, got $RESPONSE"
    else
      echo "✓ PASS: Request $i allowed"
    fi
  else
    if [ "$RESPONSE" != "429" ]; then
      echo "❌ FAIL: Request $i should return 429, got $RESPONSE"
    else
      echo "✓ PASS: Request $i blocked"
    fi
  fi
  sleep 0.2
done

echo ""
echo "Test 2: Token Rate Limiting (should allow 10, then block)"
echo "--------------------------------------------------------"
TOKEN="test-token-abc123"
for i in {1..12}; do
  RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" -H "API_KEY: $TOKEN" $BASE_URL)
  echo "Request $i: HTTP $RESPONSE"
  
  if [ $i -le 10 ]; then
    if [ "$RESPONSE" != "200" ]; then
      echo "❌ FAIL: Request $i should return 200, got $RESPONSE"
    else
      echo "✓ PASS: Request $i allowed"
    fi
  else
    if [ "$RESPONSE" != "429" ]; then
      echo "❌ FAIL: Request $i should return 429, got $RESPONSE"
    else
      echo "✓ PASS: Request $i blocked"
    fi
  fi
  sleep 0.2
done

echo ""
echo "Test 3: Different IPs (should not interfere)"
echo "-------------------------------------------"
RESPONSE1=$(curl -s -o /dev/null -w "%{http_code}" $BASE_URL)
echo "IP 1 - Request 1: HTTP $RESPONSE1"

RESPONSE2=$(curl -s -o /dev/null -w "%{http_code}" -H "X-Forwarded-For: 10.0.0.2" $BASE_URL)
echo "IP 2 - Request 1: HTTP $RESPONSE2"

if [ "$RESPONSE1" = "200" ] && [ "$RESPONSE2" = "200" ]; then
  echo "✓ PASS: Different IPs work independently"
else
  echo "❌ FAIL: Different IPs should not interfere"
fi

echo ""
echo "Test 4: Health Check"
echo "-------------------"
HEALTH=$(curl -s http://localhost:8080/health | grep -o '"status":"ok"')
if [ ! -z "$HEALTH" ]; then
  echo "✓ PASS: Health check working"
else
  echo "❌ FAIL: Health check not working"
fi

echo ""
echo "================================"
echo "Load Test Completed"
echo "================================"

