#!/usr/bin/env bash
# Manuel veri girişi: belirli kullanıcı için tek işlem oluşturur (REST).
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"

usage() {
	cat <<'EOF'
Usage: manual-input.sh <user_id> <amount> <location> [options]

Creates one transaction for the given user via REST API (POST /transactions).

Arguments:
  user_id     Target user UUID
  amount      Positive integer (amount value accepted by the API, min 1)
  location    Location string (max 100 characters)

Options:
  --base-url URL   API base URL (default: env BASE_URL or http://localhost:8080)
  -h, --help       Show this help

Environment:
  BASE_URL         Same as --base-url when the flag is omitted

Requires: curl, jq

Examples:
  ./scripts/manual-input.sh 4ca5e74c-3a07-4010-9f07-14e90a35868b 5000 Istanbul
  BASE_URL=http://127.0.0.1:8080 ./scripts/manual-input.sh "$USER_UUID" 12000 "Ankara"
EOF
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
	usage
	exit 0
fi

while [[ $# -gt 0 ]]; do
	case "$1" in
	--base-url)
		BASE_URL="${2:?missing value for $1}"
		shift 2
		;;
	-h | --help)
		usage
		exit 0
		;;
	--)
		shift
		break
		;;
	-*)
		echo "Unknown option: $1" >&2
		usage >&2
		exit 1
		;;
	*)
		break
		;;
	esac
done

if [[ $# -lt 3 ]]; then
	echo "Error: expected <user_id> <amount> <location>" >&2
	usage >&2
	exit 1
fi

USER_ID="$1"
AMOUNT="$2"
LOCATION="$3"

command -v curl >/dev/null 2>&1 || {
	echo "curl is required" >&2
	exit 1
}
command -v jq >/dev/null 2>&1 || {
	echo "jq is required" >&2
	exit 1
}

if ! [[ "$USER_ID" =~ ^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$ ]]; then
	echo "user_id must be a UUID, got: $USER_ID" >&2
	exit 1
fi

if ! [[ "$AMOUNT" =~ ^[1-9][0-9]*$ ]]; then
	echo "amount must be a positive integer, got: $AMOUNT" >&2
	exit 1
fi

if [[ ${#LOCATION} -gt 100 ]]; then
	echo "location must be at most 100 characters (got ${#LOCATION})" >&2
	exit 1
fi
if [[ -z "${LOCATION//[[:space:]]/}" ]]; then
	echo "location must not be empty" >&2
	exit 1
fi

BASE_URL="${BASE_URL%/}"
BODY="$(jq -n \
	--arg user_id "$USER_ID" \
	--argjson amount "$AMOUNT" \
	--arg location "$LOCATION" \
	'{user_id: $user_id, amount: $amount, location: $location}')"

resp_body="$(mktemp)"
trap 'rm -f "$resp_body"' EXIT

http_code=""
http_code="$(curl -sS -o "$resp_body" -w '%{http_code}' -X POST "${BASE_URL}/transactions" \
	-H 'Content-Type: application/json' \
	-d "$BODY")"

if [[ "$http_code" != "201" ]]; then
	echo "POST ${BASE_URL}/transactions failed (HTTP ${http_code}):" >&2
	cat "$resp_body" >&2
	exit 1
fi

echo "Created transaction:"
jq '.' "$resp_body"
