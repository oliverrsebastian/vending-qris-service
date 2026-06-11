#!/usr/bin/env bash
# Integration test curls for vending-qris-service.
#
# Usage:
#   export PAYMENT_SERVICE_BASE_URL=http://localhost:8080
#   export PAYMENT_SERVICE_AUTH_KEY=your-admin-key
#   bash scripts/integration/curl-examples.sh [command]
#
# Commands: all | health | qris | callback-stub | gateways | priority | failover | poll

set -euo pipefail

BASE_URL="${PAYMENT_SERVICE_BASE_URL:-http://localhost:8080}"
AUTH_KEY="${PAYMENT_SERVICE_AUTH_KEY:-}"

cmd="${1:-all}"

curl_health() {
  echo "== GET /health (public) =="
  curl -sS -w "\nHTTP %{http_code}\n" \
    -H "Accept: application/json" \
    "${BASE_URL}/health"
  echo
}

curl_callback_stub() {
  echo "== POST /v1/callbacks/stub (public webhook) =="
  curl -sS -w "\nHTTP %{http_code}\n" \
    -X POST "${BASE_URL}/v1/callbacks/stub" \
    -H "Content-Type: application/json" \
    -H "Accept: application/json" \
    -d '{
      "transaction_id": 1,
      "status": "PAID",
      "reference_id": "stub-ref-001"
    }'
  echo
}

curl_generate_qris() {
  echo "== POST /v1/payments/qris/dynamic (public) =="
  curl -sS -w "\nHTTP %{http_code}\n" \
    -X POST "${BASE_URL}/v1/payments/qris/dynamic" \
    -H "Content-Type: application/json" \
    -H "Accept: application/json" \
    -d '{
      "description": "integration test purchase",
      "invoice_number": "INV-2026-0001",
      "products": [
        {
          "name": "Mineral Water",
          "quantity": 2,
          "item_price": 5000
        },
        {
          "name": "Snack Bar",
          "quantity": 1,
          "item_price": 12000
        }
      ]
    }'
  echo
}

curl_list_gateways() {
  echo "== GET /v1/admin/payment-gateways (authenticated) =="
  curl -sS -w "\nHTTP %{http_code}\n" \
    -H "Accept: application/json" \
    -H "Authorization: ${AUTH_KEY}" \
    "${BASE_URL}/v1/admin/payment-gateways"
  echo
}

curl_set_priority() {
  echo "== PUT /v1/admin/payment-gateways/priority (authenticated) =="
  curl -sS -w "\nHTTP %{http_code}\n" \
    -X PUT "${BASE_URL}/v1/admin/payment-gateways/priority" \
    -H "Content-Type: application/json" \
    -H "Accept: application/json" \
    -H "Authorization: ${AUTH_KEY}" \
    -d '{
      "gateways": ["stub", "stub_fallback", "stub_down"]
    }'
  echo
}

curl_failover() {
  echo "== POST /v1/admin/payment-gateways/failover (authenticated) =="
  curl -sS -w "\nHTTP %{http_code}\n" \
    -X POST "${BASE_URL}/v1/admin/payment-gateways/failover" \
    -H "Accept: application/json" \
    -H "Authorization: ${AUTH_KEY}" \
    "${BASE_URL}/v1/admin/payment-gateways/failover"
  echo
}

curl_poll_communications() {
  echo "== POST /v1/admin/payment-communications/poll (authenticated) =="
  curl -sS -w "\nHTTP %{http_code}\n" \
    -X POST "${BASE_URL}/v1/admin/payment-communications/poll" \
    -H "Accept: application/json" \
    -H "Authorization: ${AUTH_KEY}" \
    "${BASE_URL}/v1/admin/payment-communications/poll"
  echo
}

require_auth() {
  if [[ -z "${AUTH_KEY}" ]]; then
    echo "PAYMENT_SERVICE_AUTH_KEY is required for admin endpoints" >&2
    exit 1
  fi
}

case "${cmd}" in
  health) curl_health ;;
  qris) curl_generate_qris ;;
  callback-stub) curl_callback_stub ;;
  gateways) require_auth; curl_list_gateways ;;
  priority) require_auth; curl_set_priority ;;
  failover) require_auth; curl_failover ;;
  poll) require_auth; curl_poll_communications ;;
  all)
    curl_health
    curl_generate_qris
    curl_callback_stub
    require_auth
    curl_list_gateways
    curl_set_priority
    curl_failover
    curl_poll_communications
    ;;
  *)
    echo "unknown command: ${cmd}" >&2
    echo "usage: $0 [all|health|qris|callback-stub|gateways|priority|failover|poll]" >&2
    exit 1
    ;;
esac
