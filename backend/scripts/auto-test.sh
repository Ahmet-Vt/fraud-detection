#!/usr/bin/env bash
set -euo pipefail

usage() {
	cat <<'EOF'
Usage: auto-test.sh [options]

Options:
  --duration=<seconds>                 Script runtime in seconds (default: 30)
  --rate=<requests_per_second>         Transactions to send per second (default: 5)
  --anomaly=<percentage>               Anomaly request probability [0-100] (default: 30)
  -h, --help                           Show help

Examples:
  ./scripts/auto-test.sh --duration=20 --rate=5 --anomaly=50
  BASE_URL=http://localhost:8080 ./scripts/auto-test.sh --anomaly=100
EOF
}

DURATION=30
RATE=5
ANOMALY_CHANCE=30
BASE_URL="${BASE_URL:-http://localhost:8080}"
HOME_CITY="Istanbul"
FAR_CITIES_CSV="Van,Hakkari,Artvin,Kars,Edirne"
NORMAL_AMOUNT_MIN=300
NORMAL_AMOUNT_MAX=900
ANOMALY_AMOUNT=45000
BURST_SIZE=6
SETTLE_SECONDS=2

for arg in "$@"; do
	case "$arg" in
	--duration=*)
		DURATION="${arg#*=}"
		;;
	--rate=*)
		RATE="${arg#*=}"
		;;
	--anomaly=*)
		ANOMALY_CHANCE="${arg#*=}"
		;;
	--anomaly-chance=*)
		ANOMALY_CHANCE="${arg#*=}"
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

require_positive_int() {
	local name="$1"
	local value="$2"
	if ! [[ "$value" =~ ^[1-9][0-9]*$ ]]; then
		echo "$name must be a positive integer: $value" >&2
		exit 1
	fi
}

require_non_negative_int() {
	local name="$1"
	local value="$2"
	if ! [[ "$value" =~ ^[0-9]+$ ]]; then
		echo "$name must be a non-negative integer: $value" >&2
		exit 1
	fi
}

require_percentage() {
	local name="$1"
	local value="$2"
	if ! [[ "$value" =~ ^([0-9]|[1-9][0-9]|100)$ ]]; then
		echo "$name must be in the range [0-100]: $value" >&2
		exit 1
	fi
}

require_positive_int "duration" "$DURATION"
require_positive_int "rate" "$RATE"
require_percentage "anomaly" "$ANOMALY_CHANCE"

if ((NORMAL_AMOUNT_MIN > NORMAL_AMOUNT_MAX)); then
	echo "normal-amount-min cannot be greater than normal-amount-max" >&2
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

RESULTS_FILE="$(mktemp)"
FAILURES_FILE="$(mktemp)"
trap 'rm -f "$RESULTS_FILE" "$FAILURES_FILE"' EXIT

IFS=',' read -r -a FAR_CITIES <<<"$FAR_CITIES_CSV"
if ((${#FAR_CITIES[@]} == 0)); then
	echo "far-cities cannot be empty" >&2
	exit 1
fi

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

random_amount() {
	local min="$1"
	local max="$2"
	printf '%s\n' $((min + RANDOM % (max - min + 1)))
}

random_far_city() {
	local idx
	idx=$((RANDOM % ${#FAR_CITIES[@]}))
	printf '%s\n' "${FAR_CITIES[$idx]}"
}

now_ms() {
	date +%s%3N
}

pick_anomaly_scenario() {
	case $((RANDOM % 4)) in
	0) printf 'amount\n' ;;
	1) printf 'velocity\n' ;;
	2) printf 'location\n' ;;
	*) printf 'combined\n' ;;
	esac
}

send_transaction() {
	local user_id="$1"
	local amount="$2"
	local location="$3"
	local scenario="$4"
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
		printf '%s\t%s\t%s\t%s\n' "$scenario" "$amount" "$location" "$status" >>"$FAILURES_FILE"
		rm -f "$response_file"
		return 1
	fi

	local transaction_id
	local transaction_status
	local fraud_status
	transaction_id="$(jq -er '.id' "$response_file")"
	transaction_status="$(jq -r '.transaction_status // "unknown"' "$response_file")"
	fraud_status="$(jq -r '.fraud_status // "unknown"' "$response_file")"
	printf '%s\t%s\t%s\t%s\t%s\n' "$scenario" "$transaction_id" "$amount" "$location" "$status" >>"$RESULTS_FILE"
	printf '[%s] id=%s amount=%s location=%s response_status=%s fraud_status=%s\n' \
		"$scenario" "$transaction_id" "$amount" "$location" "$transaction_status" "$fraud_status"

	rm -f "$response_file"
	return 0
}

send_normal() {
	local amount
	amount="$(random_amount "$NORMAL_AMOUNT_MIN" "$NORMAL_AMOUNT_MAX")"
	send_transaction "$USER_ID" "$amount" "$HOME_CITY" "normal"
}

send_amount_anomaly() {
	send_transaction "$USER_ID" "$ANOMALY_AMOUNT" "$HOME_CITY" "amount"
}

send_velocity_anomaly() {
	local amount
	amount="$(random_amount "$NORMAL_AMOUNT_MIN" "$NORMAL_AMOUNT_MAX")"
	send_transaction "$USER_ID" "$amount" "$HOME_CITY" "velocity"
}

send_location_anomaly() {
	local amount
	local far_city
	amount="$(random_amount "$NORMAL_AMOUNT_MIN" "$NORMAL_AMOUNT_MAX")"
	far_city="$(random_far_city)"
	send_transaction "$USER_ID" "$amount" "$far_city" "location"
}

send_combined_anomaly() {
	local amount
	local far_city

	amount="$ANOMALY_AMOUNT"
	far_city="$(random_far_city)"
	send_transaction "$USER_ID" "$amount" "$far_city" "combined"
}

print_final_statuses() {
	local total
	total="$(wc -l <"$RESULTS_FILE" | tr -d ' ')"
	if ((total == 0)); then
		return
	fi

	if ((SETTLE_SECONDS > 0)); then
		echo
		echo "Waiting ${SETTLE_SECONDS}s for queue processing..."
		sleep "$SETTLE_SECONDS"
	fi

	echo
	echo "Final transaction statuses:"
	local scenario transaction_id amount location http_status
	while IFS=$'\t' read -r scenario transaction_id amount location http_status; do
		local response_file
		response_file="$(mktemp)"
		local status
		status=$(curl -sS -o "$response_file" -w '%{http_code}' \
			-X GET "${BASE_URL%/}/transactions/${transaction_id}")

		if [[ "$status" == "200" ]]; then
			local transaction_status
			local fraud_status
			transaction_status="$(jq -r '.transaction_status // "unknown"' "$response_file")"
			fraud_status="$(jq -r '.fraud_status // "unknown"' "$response_file")"
			printf '[%s] id=%s amount=%s location=%s final_status=%s fraud_status=%s\n' \
				"$scenario" "$transaction_id" "$amount" "$location" "$transaction_status" "$fraud_status"
		else
			printf '[%s] id=%s final lookup failed HTTP %s\n' "$scenario" "$transaction_id" "$status" >&2
			cat "$response_file" >&2
		fi

		rm -f "$response_file"
	done <"$RESULTS_FILE"
}

USER_ID="$(create_user)"

TOTAL_REQUESTS=$((DURATION * RATE))

echo "Running auto test..."
echo "Base URL: ${BASE_URL%/}"
echo "Duration: ${DURATION}s"
echo "Rate: ${RATE} transactions/s"
echo "Anomaly chance: ${ANOMALY_CHANCE}%"
echo "User ID: $USER_ID"
echo "Home city: $HOME_CITY"
echo "Far cities: $FAR_CITIES_CSV"
echo "Normal amount range: ${NORMAL_AMOUNT_MIN}-${NORMAL_AMOUNT_MAX}"
echo "Anomaly amount: $ANOMALY_AMOUNT"
echo "Burst size: $BURST_SIZE"
echo "Planned requests: $TOTAL_REQUESTS"
echo

anomaly_count=0
normal_count=0
velocity_remaining=0

for ((second = 1; second <= DURATION; second++)); do
	worker_pids=()
	batch_started_ms="$(now_ms)"

	for ((slot = 1; slot <= RATE; slot++)); do
		if ((velocity_remaining > 0)); then
			send_velocity_anomaly &
			worker_pids+=("$!")
			velocity_remaining=$((velocity_remaining - 1))
			((anomaly_count += 1))
		else
			rand=$((RANDOM % 100))
			if ((rand < ANOMALY_CHANCE)); then
				scenario="$(pick_anomaly_scenario)"
				case "$scenario" in
				amount)
					send_amount_anomaly &
					;;
				velocity)
					send_velocity_anomaly &
					velocity_remaining=$((BURST_SIZE - 1))
					;;
				location)
					send_location_anomaly &
					;;
				combined)
					send_combined_anomaly &
					;;
				esac
				worker_pids+=("$!")
				((anomaly_count += 1))
			else
				send_normal &
				worker_pids+=("$!")
				((normal_count += 1))
			fi
		fi
	done

	for pid in "${worker_pids[@]}"; do
		wait "$pid" || true
	done

	if ((second < DURATION)); then
		batch_elapsed_ms=$(($(now_ms) - batch_started_ms))
		sleep_ms=$((1000 - batch_elapsed_ms))
		if ((sleep_ms > 0)); then
			sleep "$(awk "BEGIN { printf \"%.3f\", $sleep_ms / 1000 }")"
		fi
	fi
done

print_final_statuses

success_count="$(wc -l <"$RESULTS_FILE" | tr -d ' ')"
failure_count="$(wc -l <"$FAILURES_FILE" | tr -d ' ')"

echo
echo "Finished."
echo "Planned requests: $TOTAL_REQUESTS"
echo "Sent successfully: $success_count"
echo "Failed: $failure_count"
echo "Normal requests: $normal_count"
echo "Anomaly requests: $anomaly_count"
echo "Auto test completed."
