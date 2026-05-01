#!/usr/bin/env bash
set -euo pipefail

usage() {
	cat <<'EOF'
Usage: auto-test.sh [options]

Options:
  --duration=<seconds>                 Script runtime in seconds (default: 30)
  --rate=<requests_per_second>         Transactions to send per second (default: 5)
  --anomaly-chance=<percentage>        Anomaly request probability [0-100] (default: 30)
  --base-url=<url>                     API base URL (default: env BASE_URL or http://localhost:8080)
  --user-id=<uuid>                     Existing user id (creates one if omitted)
  --home-city=<city>                   Normal traffic city from city.go (default: Istanbul)
  --far-cities=<city,city>             Far city pool from city.go (default: Van,Hakkari,Artvin,Kars,Edirne)
  --normal-amount-min=<amount>         Normal minimum amount (default: 300)
  --normal-amount-max=<amount>         Normal maximum amount (default: 900)
  --anomaly-amount=<amount>            Amount used by amount anomalies (default: 45000)
  --burst-size=<count>                 Requests sent by velocity scenarios (default: 6)
  --settle-seconds=<seconds>           Wait before reading final statuses (default: 2)
  --amount-weight=<weight>             Weight inside anomaly scenarios (default: 35)
  --velocity-weight=<weight>           Weight inside anomaly scenarios (default: 25)
  --location-weight=<weight>           Weight inside anomaly scenarios (default: 25)
  --combined-weight=<weight>           Weight inside anomaly scenarios (default: 15)
  -h, --help                           Show help

Examples:
  ./scripts/auto-test.sh --duration=20 --rate=5 --anomaly-chance=50 --anomaly-amount=45000
  ./scripts/auto-test.sh --user-id=<uuid> --home-city=Istanbul --far-cities=Van,Hakkari,Kars
EOF
}

DURATION=30
RATE=5
ANOMALY_CHANCE=30
BASE_URL="${BASE_URL:-http://localhost:8080}"
USER_ID="${USER_ID:-}"
HOME_CITY="Istanbul"
FAR_CITIES_CSV="Van,Hakkari,Artvin,Kars,Edirne"
NORMAL_AMOUNT_MIN=300
NORMAL_AMOUNT_MAX=900
ANOMALY_AMOUNT=45000
BURST_SIZE=6
SETTLE_SECONDS=2
AMOUNT_WEIGHT=35
VELOCITY_WEIGHT=25
LOCATION_WEIGHT=25
COMBINED_WEIGHT=15

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
	--home-city=*)
		HOME_CITY="${arg#*=}"
		;;
	--far-cities=*)
		FAR_CITIES_CSV="${arg#*=}"
		;;
	--normal-amount-min=*)
		NORMAL_AMOUNT_MIN="${arg#*=}"
		;;
	--normal-amount-max=*)
		NORMAL_AMOUNT_MAX="${arg#*=}"
		;;
	--anomaly-amount=*)
		ANOMALY_AMOUNT="${arg#*=}"
		;;
	--burst-size=*)
		BURST_SIZE="${arg#*=}"
		;;
	--settle-seconds=*)
		SETTLE_SECONDS="${arg#*=}"
		;;
	--amount-weight=*)
		AMOUNT_WEIGHT="${arg#*=}"
		;;
	--velocity-weight=*)
		VELOCITY_WEIGHT="${arg#*=}"
		;;
	--location-weight=*)
		LOCATION_WEIGHT="${arg#*=}"
		;;
	--combined-weight=*)
		COMBINED_WEIGHT="${arg#*=}"
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
require_percentage "anomaly-chance" "$ANOMALY_CHANCE"
require_positive_int "normal-amount-min" "$NORMAL_AMOUNT_MIN"
require_positive_int "normal-amount-max" "$NORMAL_AMOUNT_MAX"
require_positive_int "anomaly-amount" "$ANOMALY_AMOUNT"
require_positive_int "burst-size" "$BURST_SIZE"
require_non_negative_int "settle-seconds" "$SETTLE_SECONDS"
require_non_negative_int "amount-weight" "$AMOUNT_WEIGHT"
require_non_negative_int "velocity-weight" "$VELOCITY_WEIGHT"
require_non_negative_int "location-weight" "$LOCATION_WEIGHT"
require_non_negative_int "combined-weight" "$COMBINED_WEIGHT"

if ((NORMAL_AMOUNT_MIN > NORMAL_AMOUNT_MAX)); then
	echo "normal-amount-min cannot be greater than normal-amount-max" >&2
	exit 1
fi

TOTAL_ANOMALY_WEIGHT=$((AMOUNT_WEIGHT + VELOCITY_WEIGHT + LOCATION_WEIGHT + COMBINED_WEIGHT))
if ((TOTAL_ANOMALY_WEIGHT <= 0)); then
	echo "at least one anomaly scenario weight must be greater than zero" >&2
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
	local roll
	roll=$((RANDOM % TOTAL_ANOMALY_WEIGHT))

	if ((roll < AMOUNT_WEIGHT)); then
		printf 'amount\n'
		return
	fi
	roll=$((roll - AMOUNT_WEIGHT))

	if ((roll < VELOCITY_WEIGHT)); then
		printf 'velocity\n'
		return
	fi
	roll=$((roll - VELOCITY_WEIGHT))

	if ((roll < LOCATION_WEIGHT)); then
		printf 'location\n'
		return
	fi

	printf 'combined\n'
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

if [[ -z "$USER_ID" ]]; then
	USER_ID="$(create_user)"
fi

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
