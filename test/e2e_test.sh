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

cleanup() { rm -rf "$VAULT" "$KNO" "${PUBLISH_DIR:-}"; }
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
assert_contains "setup creates vault" "Vault created" "$KNO" setup --vault "$VAULT" --no-register
test -f "$VAULT/config.toml" && pass "config.toml exists" || fail "config.toml missing"

# --- Setup --publish ---
echo ""
echo "Setup publish"
SETUP_PUB_VAULT=$(mktemp -d /tmp/kno-e2e-setup-pub.XXXXXX)
rm -rf "$SETUP_PUB_VAULT"
SETUP_PUB_DIR=$(mktemp -d /tmp/kno-e2e-setup-pubdir.XXXXXX)
assert_contains "setup --publish" "Publish target added" "$KNO" setup --vault "$SETUP_PUB_VAULT" --no-register --publish "$SETUP_PUB_DIR"
grep -q "frontmatter" "$SETUP_PUB_VAULT/config.toml" && pass "publish config written" || fail "publish config missing"
rm -rf "$SETUP_PUB_VAULT" "$SETUP_PUB_DIR"

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

# --- Multi-value metadata round-trip ---
echo ""
echo "Multi-value metadata"

MULTI_JSON=$("$KNO" --vault "$VAULT" note show "$NOTE1" --json 2>&1)
echo "$MULTI_JSON" | python3 -c "import sys,json; data=json.load(sys.stdin); n=data[0] if isinstance(data,list) else data; assert 'databases' in str(n.get('metadata',{})), f'metadata: {n.get(\"metadata\")}'" 2>/dev/null && pass "note metadata in JSON output" || fail "note metadata missing from JSON"

# --- Note content update via stdin ---
echo ""
echo "Note content update"

echo "## Updated TL;DR
New debugging insight." | "$KNO" --vault "$VAULT" note update "$NOTE1" >/dev/null 2>&1
UPDATED_NOTE=$("$KNO" --vault "$VAULT" note show "$NOTE1" 2>&1)
echo "$UPDATED_NOTE" | grep -q "New debugging insight" && pass "note content updated via stdin" || fail "note content update via stdin failed"

# --- Page rename ---
echo ""
echo "Page rename"

assert_exit_0 "page rename" "$KNO" --vault "$VAULT" page rename cnc-machine-maintenance --name "CNC Operations"
assert_contains "renamed page in list" "CNC Operations" "$KNO" --vault "$VAULT" page list
assert_not_contains "old name gone" "CNC Machine Maintenance" "$KNO" --vault "$VAULT" page list

# --- Delete and prune ---
echo ""
echo "Delete and prune"

assert_exit_0 "note delete" "$KNO" --vault "$VAULT" note delete "$NOTE2"
assert_not_contains "deleted note gone" "$NOTE2" "$KNO" --vault "$VAULT" note list

assert_contains "prune dry-run" "Would remove" "$KNO" --vault "$VAULT" note prune --count 1 --dry-run
assert_exit_0 "page delete" "$KNO" --vault "$VAULT" page delete cnc-operations
assert_not_contains "page gone after delete" "CNC Operations" "$KNO" --vault "$VAULT" page list

# --- Rebuild index ---
echo ""
echo "Rebuild index"

assert_contains "rebuild-index" "Indexed" "$KNO" --vault "$VAULT" vault rebuild-index
# Search should still work after rebuild
assert_contains "search after rebuild" "Connection pool" "$KNO" --vault "$VAULT" note search "pool"

# --- Error cases ---
echo ""
echo "Error cases"

assert_exit_nonzero "show nonexistent note" "$KNO" --vault "$VAULT" note show nonexistent-id
assert_exit_nonzero "show nonexistent page" "$KNO" --vault "$VAULT" page show nonexistent-id
assert_exit_nonzero "delete nonexistent page" "$KNO" --vault "$VAULT" page delete nonexistent-id

# --- Publish ---
echo ""
echo "Publish"

# No targets configured — should show help message
assert_contains "publish no targets" "No publish targets" "$KNO" --vault "$VAULT" publish

# Configure a publish target
PUBLISH_DIR=$(mktemp -d /tmp/kno-e2e-publish.XXXXXX)
cat >> "$VAULT/config.toml" << 'TOML'

[[publish.targets]]
TOML
echo "path = \"$PUBLISH_DIR\"" | sed "s|\$PUBLISH_DIR|$PUBLISH_DIR|" >> "$VAULT/config.toml"
echo 'format = "frontmatter"' >> "$VAULT/config.toml"

assert_exit_0 "publish with target" "$KNO" --vault "$VAULT" publish
test -f "$PUBLISH_DIR/aws-infrastructure.md" && pass "published file exists" || fail "published file missing"

# Check frontmatter structure
head -1 "$PUBLISH_DIR/aws-infrastructure.md" | grep -q "^---" && pass "frontmatter starts with ---" || fail "frontmatter missing opening ---"
grep -q "^title:" "$PUBLISH_DIR/aws-infrastructure.md" && pass "frontmatter has title" || fail "frontmatter missing title"
grep -q "^aliases:" "$PUBLISH_DIR/aws-infrastructure.md" && pass "frontmatter has aliases" || fail "frontmatter missing aliases"
grep -q "^created:" "$PUBLISH_DIR/aws-infrastructure.md" && pass "frontmatter has created date" || fail "frontmatter missing created"

# Check body content is present after frontmatter
grep -q "Updated content" "$PUBLISH_DIR/aws-infrastructure.md" && pass "published body has content" || fail "published body missing content"

# Publish single page
assert_exit_0 "publish single page" "$KNO" --vault "$VAULT" publish --page aws-infrastructure

# Publish nonexistent page
assert_exit_nonzero "publish nonexistent page" "$KNO" --vault "$VAULT" publish --page nonexistent-id

# Dry run
assert_contains "publish dry-run" "Would publish" "$KNO" --vault "$VAULT" publish --dry-run

# JSON output
PUB_JSON=$("$KNO" --vault "$VAULT" publish --json 2>&1)
echo "$PUB_JSON" | python3 -c "import sys,json; d=json.load(sys.stdin); assert len(d) > 0" 2>/dev/null && pass "publish --json valid" || fail "publish --json invalid"

# Invalid format
assert_exit_nonzero "publish invalid format" "$KNO" --vault "$VAULT" publish --format bogus

# Markdown format override
assert_exit_0 "publish markdown format" "$KNO" --vault "$VAULT" publish --format markdown
# Markdown format should not have frontmatter
grep -q "^---" "$PUBLISH_DIR/aws-infrastructure.md" && fail "markdown format has frontmatter" || pass "markdown format no frontmatter"

# Guidance stripping — add guidance then publish
echo '<!-- Guidance -->
Focus on ops.

## Real Content

The actual knowledge.' | "$KNO" --vault "$VAULT" page update aws-infrastructure >/dev/null 2>&1
"$KNO" --vault "$VAULT" publish >/dev/null 2>&1
grep -q "Guidance" "$PUBLISH_DIR/aws-infrastructure.md" && fail "guidance not stripped" || pass "guidance stripped from published"
grep -q "Real Content" "$PUBLISH_DIR/aws-infrastructure.md" && pass "content preserved after guidance strip" || fail "content lost after guidance strip"

# --- Curate-publish loop (real-world metadata flow) ---
echo ""
echo "Curate-publish loop"

# Simulate what curate does: stamp page with tags, summary, note_count, last_curated_at
assert_exit_0 "page update with curate metadata" bash -c "echo '## RDS

- Pin parameter groups before upgrades.

## ECS

- Drain window: 60s minimum.' | '$KNO' --vault '$VAULT' page update aws-infrastructure --meta tags=aws --meta tags=rds --meta tags=ecs --meta summary='Operational lessons for AWS infrastructure' --meta note_count=3 --meta last_curated_at=2026-03-07T10:00:00Z"

# Publish and verify rich frontmatter
"$KNO" --vault "$VAULT" publish >/dev/null 2>&1
grep -q "^tags:" "$PUBLISH_DIR/aws-infrastructure.md" && pass "frontmatter has tags after curate" || fail "frontmatter missing tags after curate"
grep -q "aws" "$PUBLISH_DIR/aws-infrastructure.md" && pass "frontmatter includes aws tag" || fail "frontmatter missing aws tag"
grep -q "^summary:" "$PUBLISH_DIR/aws-infrastructure.md" && pass "frontmatter has summary" || fail "frontmatter missing summary"
grep -q "^updated:" "$PUBLISH_DIR/aws-infrastructure.md" && pass "frontmatter has updated date" || fail "frontmatter missing updated date"
grep -q "2026-03-07" "$PUBLISH_DIR/aws-infrastructure.md" && pass "updated date from last_curated_at" || fail "updated date wrong"
# Body content should be present, guidance should not
grep -q "Pin parameter groups" "$PUBLISH_DIR/aws-infrastructure.md" && pass "curated body content published" || fail "curated body content missing"

# --- Prune actual execution ---
echo ""
echo "Prune execution"

# Create a throwaway note to prune
echo "Throwaway." | "$KNO" --vault "$VAULT" note create --title "Prune target" >/dev/null 2>&1
BEFORE_COUNT=$("$KNO" --vault "$VAULT" note list --json 2>&1 | python3 -c "import sys,json; print(len(json.load(sys.stdin)))")
assert_exit_0 "prune execute" "$KNO" --vault "$VAULT" note prune --count 1
AFTER_COUNT=$("$KNO" --vault "$VAULT" note list --json 2>&1 | python3 -c "import sys,json; print(len(json.load(sys.stdin)))")
[ "$AFTER_COUNT" -lt "$BEFORE_COUNT" ] && pass "prune reduced note count" || fail "prune did not reduce count"

# --- Setup idempotency ---
echo ""
echo "Setup idempotency"

IDEM_VAULT=$(mktemp -d /tmp/kno-e2e-idem.XXXXXX)
rm -rf "$IDEM_VAULT"
"$KNO" setup --vault "$IDEM_VAULT" --no-register >/dev/null 2>&1
# Add a note to the vault
echo "Test note." | "$KNO" --vault "$IDEM_VAULT" note create --title "Before re-setup" >/dev/null 2>&1
# Run setup again on the same vault
IDEM_OUT=$("$KNO" setup --vault "$IDEM_VAULT" --no-register 2>&1) || true
# Note should still be there
"$KNO" --vault "$IDEM_VAULT" note list 2>&1 | grep -q "Before re-setup" && pass "setup idempotent preserves data" || fail "setup idempotent lost data"
rm -rf "$IDEM_VAULT"

# --- Page multi-value metadata ---
echo ""
echo "Page multi-value metadata"

PAGE_JSON=$("$KNO" --vault "$VAULT" page show aws-infrastructure --json 2>&1)
echo "$PAGE_JSON" | python3 -c "import sys,json; p=json.load(sys.stdin); tags=p.get('metadata',{}).get('tags',[]); assert len(tags) >= 2, f'tags: {tags}'" 2>/dev/null && pass "page multi-value tags in JSON" || fail "page multi-value tags missing"

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
