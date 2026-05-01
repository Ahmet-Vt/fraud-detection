#!/usr/bin/env bash
set -euo pipefail

usage() {
	cat <<'EOF'
Usage: auto-test.sh [options]

Options:
  --duration=<seconds>            Script runtime in seconds (default: 30)
  --rate=<requests_per_second>    Requests per second to send (default: 20)
  --anomaly-chance=<percentage>   Fraud generation probability [0-100] (default: 25)
  --base-url=<url>                API base URL (default: http://localhost:8080)
  --user-id=<uuid>                Existing user id (creates one if omitted)
  -h, --help                      Show help

Examples:
  ./scripts/auto-test.sh --duration=15 --rate=50 --anomaly-chance=40
  ./scripts/auto-test.sh --duration=10 --rate=100 --anomaly-chance=20 --user-id=<uuid>
EOF
}

DURATION=30
RATE=20
ANOMALY_CHANCE=25
BASE_URL="${BASE_URL:-http://localhost:8080}"
USER_ID="${USER_ID:-}"

for arg in "$@"; do
	case "$arg" in
	--duration=*)
		DURATION="${arg#*=}"
		;;
	--rate=*)
		RATE="${arg#*=}"
		;;
	--anomaly-chance=*)
		ANOMALY_CHANCE="${arg#*=}"
		;;
	--base-url=*)
		BASE_URL="${arg#*=}"
		;;
	--user-id=*)
		USER_ID="${arg#*=}"
		;;
	-h | --help)
		usage
		exit 0
		;;
	*)
		echo "Unknown option: $arg" >&2
		usage >&2
		exit 1
		;;
	esac
done

if ! [[ "$DURATION" =~ ^[1-9][0-9]*$ ]]; then
	echo "duration must be a positive integer: $DURATION" >&2
	exit 1
fi

if ! [[ "$RATE" =~ ^[1-9][0-9]*$ ]]; then
	echo "rate must be a positive integer: $RATE" >&2
	exit 1
fi

if ! [[ "$ANOMALY_CHANCE" =~ ^([0-9]|[1-9][0-9]|100)$ ]]; then
	echo "anomaly-chance must be in the range [0-100]: $ANOMALY_CHANCE" >&2
	exit 1
fi

command -v curl >/dev/null 2>&1 || {
	echo "curl is required" >&2
	exit 1
}
command -v jq >/dev/null 2>&1 || {
	echo "jq is required" >&2
	exit 1
}

create_user() {
	local suffix
	suffix="$(date +%s)-$RANDOM"
	local payload
	payload=$(printf '{"name":"load","surname":"test","email":"load-%s@seed.local","status":"active"}' "$suffix")
	local response_file
	response_file="$(mktemp)"

	local status
	status=$(curl -sS -o "$response_file" -w '%{http_code}' \
		-X POST "${BASE_URL%/}/users" \
		-H 'Content-Type: application/json' \
		-d "$payload")

	if [[ "$status" != "201" ]]; then
		echo "failed to create user (HTTP $status)" >&2
		cat "$response_file" >&2
		rm -f "$response_file"
		exit 1
	fi

	local id
	id="$(jq -er '.id' "$response_file")"
	rm -f "$response_file"
	printf '%s\n' "$id"
}

send_transaction() {
	local user_id="$1"
	local amount="$2"
	local location="$3"
	local response_file
	response_file="$(mktemp)"
	local payload
	payload=$(printf '{"user_id":"%s","amount":%s,"location":"%s"}' "$user_id" "$amount" "$location")

	local status
	status=$(curl -sS -o "$response_file" -w '%{http_code}' \
		-X POST "${BASE_URL%/}/transactions" \
		-H 'Content-Type: application/json' \
		-d "$payload")

	if [[ "$status" != "201" ]]; then
		echo "POST /transactions failed (HTTP $status), payload=$payload" >&2
		cat "$response_file" >&2
		rm -f "$response_file"
		return 1
	fi

	rm -f "$response_file"
	return 0
}

if [[ -z "$USER_ID" ]]; then
	USER_ID="$(create_user)"
fi

TOTAL_REQUESTS=$((DURATION * RATE))
INTERVAL_SECONDS="$(awk "BEGIN { printf \"%.6f\", 1 / $RATE }")"

echo "Running auto test..."
echo "Base URL: ${BASE_URL%/}"
echo "Duration: ${DURATION}s"
echo "Rate: ${RATE} req/s"
echo "Anomaly chance: ${ANOMALY_CHANCE}%"
echo "User ID: $USER_ID"
echo "Total planned requests: $TOTAL_REQUESTS"
echo

anomaly_count=0
normal_count=0
success_count=0
failure_count=0

for ((i = 1; i <= TOTAL_REQUESTS; i++)); do
	rand=$((RANDOM % 100))
	if ((rand < ANOMALY_CHANCE)); then
		# Fraud candidate: high amount + far location
		amount=$((1501 + RANDOM % 500))
		location="Diyarbakir"
		((anomaly_count += 1))
	else
		amount=$((300 + RANDOM % 401))
		location="Bursa"
		((normal_count += 1))
	fi

	if send_transaction "$USER_ID" "$amount" "$location"; then
		((success_count += 1))
	else
		((failure_count += 1))
	fi

	if ((i < TOTAL_REQUESTS)); then
		sleep "$INTERVAL_SECONDS"
	fi
done

echo
echo "Finished."
echo "Sent successfully: $success_count"
echo "Failed: $failure_count"
echo "Normal requests: $normal_count"
echo "Anomaly requests: $anomaly_count"
echo "Auto test completed."
