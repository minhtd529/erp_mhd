#!/usr/bin/env bash
# Migration lint — enforce invariants on apps/api/migrations/

set -euo pipefail

MIGRATIONS_DIR="apps/api/migrations"
errors=0

cd "$(git rev-parse --show-toplevel)"

if [ ! -d "$MIGRATIONS_DIR" ]; then
    echo "ERROR: $MIGRATIONS_DIR not found"
    exit 1
fi

echo ">>> Migration lint starting..."
echo ">>> Directory: $MIGRATIONS_DIR"
echo ""

# ─── Check 1: Every .up.sql has a matching .down.sql ─────────────────
echo "[1/4] Checking .up.sql ↔ .down.sql pairs..."
pair_ok=true
for up in "$MIGRATIONS_DIR"/*.up.sql; do
    [ -e "$up" ] || continue
    down="${up%.up.sql}.down.sql"
    if [ ! -f "$down" ]; then
        echo "  ✗ Missing down migration: $down"
        errors=$((errors + 1))
        pair_ok=false
    fi
done
for down in "$MIGRATIONS_DIR"/*.down.sql; do
    [ -e "$down" ] || continue
    up="${down%.down.sql}.up.sql"
    if [ ! -f "$up" ]; then
        echo "  ✗ Orphan down migration (no up): $down"
        errors=$((errors + 1))
        pair_ok=false
    fi
done
$pair_ok && echo "  ✓ All pairs matched"

# ─── Check 2: Version sequence has no gaps ───────────────────────────
echo "[2/4] Checking version sequence..."
versions=$(ls "$MIGRATIONS_DIR"/*.up.sql 2>/dev/null | \
    sed -E 's|.*/([0-9]+)_.*|\1|' | \
    sort -n)

prev=0
gap_found=false
for v in $versions; do
    v_int=$((10#$v))
    expected=$((prev + 1))
    if [ "$v_int" -ne "$expected" ]; then
        echo "  ✗ Version gap: expected $(printf '%06d' $expected), got $(printf '%06d' $v_int)"
        errors=$((errors + 1))
        gap_found=true
    fi
    prev=$v_int
done
$gap_found || echo "  ✓ No gaps (1 → $prev)"

# ─── Check 3: No duplicate CREATE TABLE across files ─────────────────
echo "[3/4] Checking for duplicate CREATE TABLE..."
dupes=$(grep -h -oE 'CREATE TABLE[[:space:]]+(IF NOT EXISTS[[:space:]]+)?[a-zA-Z_][a-zA-Z0-9_]*' \
    "$MIGRATIONS_DIR"/*.up.sql 2>/dev/null | \
    sed -E 's/CREATE TABLE[[:space:]]+(IF NOT EXISTS[[:space:]]+)?//' | \
    tr '[:upper:]' '[:lower:]' | \
    sort | uniq -d || true)

if [ -n "$dupes" ]; then
    while IFS= read -r table; do
        echo "  ✗ Duplicate CREATE TABLE: $table"
        grep -ril "CREATE TABLE[[:space:]]\+\(IF NOT EXISTS[[:space:]]\+\)\?${table}" \
            "$MIGRATIONS_DIR"/*.up.sql | sed 's|^|      - |'
        errors=$((errors + 1))
    done <<< "$dupes"
else
    echo "  ✓ No duplicate CREATE TABLE"
fi

# ─── Check 4: Version numbers are zero-padded 6 digits ──────────────
echo "[4/4] Checking filename format..."
format_ok=true
for f in "$MIGRATIONS_DIR"/*.sql; do
    [ -e "$f" ] || continue
    basename_f=$(basename "$f")
    if ! echo "$basename_f" | grep -qE '^[0-9]{6}_[a-z0-9][a-z0-9_]*\.(up|down)\.sql$'; then
        echo "  ✗ Invalid filename format: $basename_f"
        echo "    Expected: 000NNN_snake_case_name.{up,down}.sql"
        errors=$((errors + 1))
        format_ok=false
    fi
done
$format_ok && echo "  ✓ All filenames valid"

echo ""
if [ "$errors" -gt 0 ]; then
    echo "FAILED: $errors violation(s) found"
    exit 1
fi

total=$(ls "$MIGRATIONS_DIR"/*.up.sql 2>/dev/null | wc -l | tr -d ' ')
echo "OK: $total migrations, all checks passed"
