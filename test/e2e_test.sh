#!/usr/bin/env bash
#
# End-to-end test for the kno CLI binary.
# Builds kno, creates a temp vault, and exercises every command.
#
# Usage: ./test/e2e_test.sh
# Exit code 0 = all passed, non-zero = failure.
#
set -euo pipefail

PASS=0
FAIL=0
KNO=""
VAULT=""

pass() { PASS=$((PASS + 1)); echo "  ✓ $1"; }
fail() { FAIL=$((FAIL + 1)); echo "  ✗ $1"; }

assert_exit_0() {
  local desc="$1"; shift
  if "$@" >/dev/null 2>&1; then pass "$desc"; else fail "$desc"; fi
}

assert_contains() {
  local desc="$1" expected="$2"; shift 2
  local output
  output=$("$@" 2>&1) || true
  if echo "$output" | grep -q "$expected"; then pass "$desc"; else fail "$desc: expected '$expected' in output"; echo "  got: $output"; fi
}

assert_not_contains() {
  local desc="$1" unexpected="$2"; shift 2
  local output
  output=$("$@" 2>&1) || true
  if echo "$output" | grep -q "$unexpected"; then fail "$desc: found '$unexpected' in output"; else pass "$desc"; fi
}

assert_exit_nonzero() {
  local desc="$1"; shift
  if "$@" >/dev/null 2>&1; then fail "$desc: expected non-zero exit"; else pass "$desc"; fi
}

cleanup() { rm -rf "$VAULT" "$KNO"; }
trap cleanup EXIT

# --- Build ---
echo "Building kno..."
KNO=$(mktemp /tmp/kno-e2e-bin.XXXXXX)
go build -o "$KNO" ./cmd/kno
echo ""

# --- Setup ---
VAULT=$(mktemp -d /tmp/kno-e2e-vault.XXXXXX)
rm -rf "$VAULT"  # setup expects to create it

echo "Setup"
assert_contains "setup creates vault" "Vault created" "$KNO" setup --vault "$VAULT" --no-claude-desktop
test -f "$VAULT/config.toml" && pass "config.toml exists" || fail "config.toml missing"

# --- Vault status (empty) ---
echo ""
echo "Vault status (empty)"
assert_contains "shows 0 notes" "Notes: 0" "$KNO" --vault "$VAULT" vault status
assert_contains "shows no pages" "No pages yet" "$KNO" --vault "$VAULT" vault status

# --- Note CRUD ---
echo ""
echo "Note operations"

NOTE1_OUT=$(echo "## TL;DR\nDebugged connection pool exhaustion." | "$KNO" --vault "$VAULT" note create --title "Connection pool debugging" --meta tags=aws --meta tags=rds --meta summary="Pool exhaustion fix" 2>&1)
echo "$NOTE1_OUT" | grep -q "Created" && pass "note create 1" || fail "note create 1"

NOTE2_OUT=$(echo "## TL;DR\nSpindle bearing diagnosis on CNC mill #3." | "$KNO" --vault "$VAULT" note create --title "Spindle bearing diagnosis" --meta tags=cnc --meta tags=maintenance --meta summary="Bearing replacement procedure" 2>&1)
echo "$NOTE2_OUT" | grep -q "Created" && pass "note create 2" || fail "note create 2"

# Extract IDs from JSON by title (list order is not guaranteed)
NOTE1=$("$KNO" --vault "$VAULT" note list --json 2>&1 | python3 -c "import sys,json; notes=json.load(sys.stdin); print(next(n['id'] for n in notes if 'Connection' in n['title']))")
NOTE2=$("$KNO" --vault "$VAULT" note list --json 2>&1 | python3 -c "import sys,json; notes=json.load(sys.stdin); print(next(n['id'] for n in notes if 'Spindle' in n['title']))")

assert_contains "note list shows 2 notes" "Connection pool" "$KNO" --vault "$VAULT" note list
assert_contains "note show" "Pool exhaustion fix" "$KNO" --vault "$VAULT" note show "$NOTE1"
assert_exit_0 "note update meta" "$KNO" --vault "$VAULT" note update "$NOTE1" --meta tags=databases

# Verify update
SHOW_OUT=$("$KNO" --vault "$VAULT" note show "$NOTE1" 2>&1)
echo "$SHOW_OUT" | grep -q "databases" && pass "note update reflected" || fail "note update reflected"

# --- Note filtering ---
echo ""
echo "Note filtering"
assert_contains "filter uncurated shows both" "$NOTE1" "$KNO" --vault "$VAULT" note list --filter curated_at=null

# Stamp as curated
assert_exit_0 "stamp curated" "$KNO" --vault "$VAULT" note update "$NOTE1" --meta curated_at=2026-03-06T00:00:00Z --meta curated_into=aws-infrastructure
assert_not_contains "filter uncurated excludes stamped" "$NOTE1" "$KNO" --vault "$VAULT" note list --filter curated_at=null
assert_contains "filter uncurated shows remaining" "$NOTE2" "$KNO" --vault "$VAULT" note list --filter curated_at=null

# --- Page CRUD ---
echo ""
echo "Page operations"

assert_contains "page create with content" "Created" bash -c "echo 'Focus on operational lessons.' | '$KNO' --vault '$VAULT' page create --name 'AWS Infrastructure'"
assert_contains "page create empty" "Created" "$KNO" --vault "$VAULT" page create --name "CNC Machine Maintenance"
assert_contains "page list shows 2" "AWS Infrastructure" "$KNO" --vault "$VAULT" page list
assert_contains "page show content" "operational lessons" "$KNO" --vault "$VAULT" page show aws-infrastructure

assert_exit_0 "page update" bash -c "echo 'Updated content.' | '$KNO' --vault '$VAULT' page update aws-infrastructure --meta last_curated_at=2026-03-06T00:00:00Z"
assert_contains "page update reflected" "Updated content" "$KNO" --vault "$VAULT" page show aws-infrastructure

# --- Search ---
echo ""
echo "Search"

# Search should work without index rebuild — index is created on first write
assert_contains "note search finds result" "Connection pool" "$KNO" --vault "$VAULT" note search "pool"
assert_contains "note search finds second" "Spindle" "$KNO" --vault "$VAULT" note search "spindle"
assert_contains "page search finds result" "aws-infrastructure" "$KNO" --vault "$VAULT" page search "AWS"

# --- Vault status (populated) ---
echo ""
echo "Vault status (populated)"
assert_contains "shows note count" "Notes: 2" "$KNO" --vault "$VAULT" vault status
assert_contains "shows pages" "AWS Infrastructure" "$KNO" --vault "$VAULT" vault status

# --- JSON output ---
echo ""
echo "JSON output"

JSON_OUT=$("$KNO" --vault "$VAULT" note list --json 2>&1)
echo "$JSON_OUT" | python3 -c "import sys,json; json.load(sys.stdin)" 2>/dev/null && pass "note list --json valid" || fail "note list --json invalid"

JSON_OUT=$("$KNO" --vault "$VAULT" page list --json 2>&1)
echo "$JSON_OUT" | python3 -c "import sys,json; json.load(sys.stdin)" 2>/dev/null && pass "page list --json valid" || fail "page list --json invalid"

JSON_OUT=$("$KNO" --vault "$VAULT" vault status --json 2>&1)
echo "$JSON_OUT" | python3 -c "import sys,json; d=json.load(sys.stdin); assert 'pages' in d" 2>/dev/null && pass "vault status --json has pages key" || fail "vault status --json missing pages key"

# --- Delete and prune ---
echo ""
echo "Delete and prune"

assert_contains "prune dry-run" "Would remove" "$KNO" --vault "$VAULT" note prune --count 1 --dry-run
assert_exit_0 "page delete" "$KNO" --vault "$VAULT" page delete cnc-machine-maintenance
assert_not_contains "page gone after delete" "CNC Machine Maintenance" "$KNO" --vault "$VAULT" page list

# --- Error cases ---
echo ""
echo "Error cases"

assert_exit_nonzero "show nonexistent note" "$KNO" --vault "$VAULT" note show nonexistent-id
assert_exit_nonzero "show nonexistent page" "$KNO" --vault "$VAULT" page show nonexistent-id
assert_exit_nonzero "delete nonexistent page" "$KNO" --vault "$VAULT" page delete nonexistent-id

# --- Version ---
echo ""
echo "Version"
assert_exit_0 "version output" "$KNO" --vault "$VAULT" version

# --- Summary ---
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Results: $PASS passed, $FAIL failed"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [ "$FAIL" -gt 0 ]; then exit 1; fi
