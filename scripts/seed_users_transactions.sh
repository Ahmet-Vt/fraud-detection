#!/usr/bin/env bash
set -euo pipefail

usage() {
	cat <<'EOF'
Usage: seed_users_transactions.sh [options]

Creates load by posting transactions for a fixed duration and rate.

Options:
  -d, --duration S              Script runtime in seconds (default: 30)
  -r, --rate RPS                Target requests per second (default: 100)
  -c, --concurrency K           Parallel channels (default: 4)
      --base-url URL            API base URL (default: env BASE_URL or http://localhost:8080)

Environment:
  BASE_URL   Same as --base-url if flag omitted

Requires: curl, jq

Example:
  ./scripts/seed_users_transactions.sh --duration 60 --rate 200 --concurrency 8
  BASE_URL=http://127.0.0.1:8080 ./scripts/seed_users_transactions.sh -d 30 -r 100 -c 4
EOF
}

BASE_URL="${BASE_URL:-http://localhost:8080}"
DURATION=30
RATE=100
CONCURRENCY=4
USER_ROTATE_SECONDS=10

while [[ $# -gt 0 ]]; do
	case "$1" in
	-h | --help)
		usage
		exit 0
		;;
	-d | --duration)
		DURATION="${2:?missing value for $1}"
		shift 2
		;;
	-r | --rate)
		RATE="${2:?missing value for $1}"
		shift 2
		;;
	-c | --concurrency)
		CONCURRENCY="${2:?missing value for $1}"
		shift 2
		;;
	--base-url)
		BASE_URL="${2:?missing value for $1}"
		shift 2
		;;
	*)
		echo "Unknown option: $1" >&2
		usage >&2
		exit 1
		;;
	esac
done

if ! [[ "$DURATION" =~ ^[1-9][0-9]*$ ]]; then
	echo "duration must be a positive integer, got: $DURATION" >&2
	exit 1
fi
if ! [[ "$RATE" =~ ^[1-9][0-9]*$ ]]; then
	echo "rate must be a positive integer, got: $RATE" >&2
	exit 1
fi
if ! [[ "$CONCURRENCY" =~ ^[1-9][0-9]*$ ]]; then
	echo "concurrency must be a positive integer, got: $CONCURRENCY" >&2
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

LOCATIONS=(Istanbul Ankara Izmir Bursa)

create_user() {
	local idx="$1"
	local name="user${idx}"
	local resp_body=""
	resp_body="$(mktemp)"
	local body
	body=$(printf '{"name":"%s","surname":"User","email":"%s","status":"active"}' "$name" "user${idx}@seed.local")

	local http_code
	http_code=$(curl -sS -o "$resp_body" -w '%{http_code}' -X POST "${BASE_URL}/users" \
		-H 'Content-Type: application/json' \
		-d "$body")

	if [[ "$http_code" != "201" ]]; then
		echo "POST /users failed for ${name} (HTTP ${http_code}):" >&2
		cat "$resp_body" >&2
		rm -f "$resp_body"
		exit 1
	fi

	local user_id
	user_id="$(jq -er '.id' "$resp_body")"
	rm -f "$resp_body"
	printf '%s\n' "$user_id"
}

create_transaction() {
	local user_id="$1"
	local amount="$2"
	local location="$3"
	local resp_body=""
	resp_body="$(mktemp)"

	local body
	body=$(printf '{"user_id":"%s","amount":%s,"location":"%s"}' "$user_id" "$amount" "$location")

	local http_code
	http_code=$(curl -sS -o "$resp_body" -w '%{http_code}' -X POST "${BASE_URL}/transactions" \
		-H 'Content-Type: application/json' \
		-d "$body")

	if [[ "$http_code" != "201" ]]; then
		echo "POST /transactions failed (HTTP ${http_code}):" >&2
		cat "$resp_body" >&2
		rm -f "$resp_body"
		exit 1
	fi

	rm -f "$resp_body"
}

echo "Base URL: ${BASE_URL}"
echo "Duration: ${DURATION}s"
echo "Rate: ${RATE} req/s"
echo "Parallel channels: ${CONCURRENCY}"
echo

process_channel() {
	local channel_idx="$1"
	local channel_rate="$2"
	local user_idx="$channel_idx"
	local user_id
	local interval_s
	local end_at
	local next_user_rotation_at
	local sent=0
	local seq=1
	local amount
	local loc_idx
	local location

	user_id=$(create_user "$user_idx")
	echo "[ch${channel_idx}] created user${user_idx} id=${user_id}, channel_rate=${channel_rate} req/s"

	if (( channel_rate <= 0 )); then
		echo "[ch${channel_idx}] idle (assigned rate=0)"
		return 0
	fi

	interval_s=$(awk "BEGIN { printf \"%.6f\", 1/${channel_rate} }")
	end_at=$((SECONDS + DURATION))
	next_user_rotation_at=$((SECONDS + USER_ROTATE_SECONDS))

	while (( SECONDS < end_at )); do
		if (( SECONDS >= next_user_rotation_at )); then
			((user_idx += CONCURRENCY))
			user_id=$(create_user "$user_idx")
			echo "[ch${channel_idx}] switched to user${user_idx} id=${user_id}"
			next_user_rotation_at=$((SECONDS + USER_ROTATE_SECONDS))
		fi

		amount=$((1000 + channel_idx * 100000 + seq))
		loc_idx=$(((channel_idx + seq - 1) % ${#LOCATIONS[@]}))
		location="${LOCATIONS[$loc_idx]}"
		create_transaction "$user_id" "$amount" "$location"
		((sent += 1))
		((seq += 1))
		sleep "$interval_s"
	done

	echo "[ch${channel_idx}] sent ${sent} transactions in ${DURATION}s"
}

worker_pids=()
base_rate=$((RATE / CONCURRENCY))
remainder=$((RATE % CONCURRENCY))

for ((channel = 1; channel <= CONCURRENCY; channel++)); do
	channel_rate=$base_rate
	if (( channel <= remainder )); then
		((channel_rate += 1))
	fi
	process_channel "$channel" "$channel_rate" &
	worker_pids+=("$!")
done

for pid in "${worker_pids[@]}"; do
	wait "$pid"
done

echo "Done."
