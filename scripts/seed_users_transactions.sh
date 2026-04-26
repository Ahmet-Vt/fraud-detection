#!/usr/bin/env bash
set -euo pipefail

usage() {
	cat <<'EOF'
Usage: seed_users_transactions.sh [options]

Creates users named user1, user2, ... and POSTs transactions per user.

Options:
  -u, --users N                 Number of users to create (default: 3)
  -t, --transactions-per-user M  Transactions per user (default: 2)
      --base-url URL            API base URL (default: env BASE_URL or http://localhost:8080)

Environment:
  BASE_URL   Same as --base-url if flag omitted

Requires: curl, jq

Example:
  ./scripts/seed_users_transactions.sh -u 5 -t 10
  BASE_URL=http://127.0.0.1:8080 ./scripts/seed_users_transactions.sh --users 2 --transactions-per-user 1
EOF
}

BASE_URL="${BASE_URL:-http://localhost:8080}"
USERS=3
TX_PER_USER=2

while [[ $# -gt 0 ]]; do
	case "$1" in
	-h | --help)
		usage
		exit 0
		;;
	-u | --users)
		USERS="${2:?missing value for $1}"
		shift 2
		;;
	-t | --transactions-per-user)
		TX_PER_USER="${2:?missing value for $1}"
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

if ! [[ "$USERS" =~ ^[1-9][0-9]*$ ]]; then
	echo "users must be a positive integer, got: $USERS" >&2
	exit 1
fi
if ! [[ "$TX_PER_USER" =~ ^[0-9]+$ ]] || [[ "$TX_PER_USER" -lt 0 ]]; then
	echo "transactions-per-user must be a non-negative integer, got: $TX_PER_USER" >&2
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

resp_body="$(mktemp)"
trap 'rm -f "$resp_body"' EXIT

create_user() {
	local idx="$1"
	local name="user${idx}"
	local body
	body=$(jq -n \
		--arg name "$name" \
		--arg email "user${idx}@seed.local" \
		'{name: $name, surname: "User", email: $email, status: "active"}')

	local http_code
	http_code=$(curl -sS -o "$resp_body" -w '%{http_code}' -X POST "${BASE_URL}/users" \
		-H 'Content-Type: application/json' \
		-d "$body")

	if [[ "$http_code" != "201" ]]; then
		echo "POST /users failed for ${name} (HTTP ${http_code}):" >&2
		cat "$resp_body" >&2
		exit 1
	fi

	jq -er '.id' "$resp_body"
}

create_transaction() {
	local user_id="$1"
	local amount="$2"
	local location="$3"

	local body
	body=$(jq -n \
		--arg uid "$user_id" \
		--argjson amount "$amount" \
		--arg loc "$location" \
		'{user_id: $uid, amount: $amount, location: $loc}')

	local http_code
	http_code=$(curl -sS -o "$resp_body" -w '%{http_code}' -X POST "${BASE_URL}/transactions" \
		-H 'Content-Type: application/json' \
		-d "$body")

	if [[ "$http_code" != "201" ]]; then
		echo "POST /transactions failed (HTTP ${http_code}):" >&2
		cat "$resp_body" >&2
		exit 1
	fi
}

echo "Base URL: ${BASE_URL}"
echo "Users: ${USERS}, transactions per user: ${TX_PER_USER}"
echo

for ((i = 1; i <= USERS; i++)); do
	user_id=$(create_user "$i")
	echo "Created user${i} id=${user_id}"

	for ((j = 1; j <= TX_PER_USER; j++)); do
		amount=$((1000 + i * 100 + j))
		loc_idx=$(((i + j - 1) % ${#LOCATIONS[@]}))
		location="${LOCATIONS[$loc_idx]}"
		create_transaction "$user_id" "$amount" "$location"
		echo "  transaction ${j}/${TX_PER_USER} amount=${amount} location=${location}"
	done
	echo
done

echo "Done."
