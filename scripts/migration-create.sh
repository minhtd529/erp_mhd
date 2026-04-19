#!/usr/bin/env bash
# Create new migration with auto-incremented version number.
# Usage: ./scripts/migration-create.sh <name>

set -euo pipefail

if [ -z "${1:-}" ]; then
    echo "Usage: $0 <migration_name>"
    echo "Example: $0 add_users_status_column"
    exit 1
fi

NAME="$1"

if ! echo "$NAME" | grep -qE '^[a-z][a-z0-9_]*$'; then
    echo "ERROR: migration name must be snake_case (lowercase, underscores, starts with letter)"
    echo "Got: $NAME"
    exit 1
fi

cd "$(git rev-parse --show-toplevel)"
MIGRATIONS_DIR="apps/api/migrations"

# Find highest version + 1
latest=$(ls "$MIGRATIONS_DIR"/*.up.sql 2>/dev/null | \
    sed -E 's|.*/([0-9]+)_.*|\1|' | \
    sort -n | tail -1 || echo "0")
next=$(printf "%06d" $((10#$latest + 1)))

UP="$MIGRATIONS_DIR/${next}_${NAME}.up.sql"
DOWN="$MIGRATIONS_DIR/${next}_${NAME}.down.sql"

if [ -f "$UP" ] || [ -f "$DOWN" ]; then
    echo "ERROR: files already exist: $UP"
    exit 1
fi

cat > "$UP" <<EOF
-- Migration: ${NAME}
-- Version: ${next}
-- Created: $(date +%Y-%m-%d)
-- Purpose: <describe what this migration does>
-- Dependencies: <list previous migrations this depends on, if any>
-- Tables affected: <CREATE|ALTER|DROP> <table_name>, ...
--
-- BEFORE WRITING THIS MIGRATION:
--   1. Verify no existing migration creates the same table:
--      grep -l "CREATE TABLE.*<table_name>" apps/api/migrations/*.up.sql
--   2. If adding a column to existing table: use ALTER TABLE, not CREATE.
--   3. Run: make migrate-lint  before committing.

-- Your SQL here
EOF

cat > "$DOWN" <<EOF
-- Down migration for ${next}_${NAME}
-- MUST be symmetric: if up creates X, down drops X.

-- Your reversal SQL here
EOF

echo "Created:"
echo "  $UP"
echo "  $DOWN"
echo ""
echo "Next: edit both files, then run 'make migrate-lint' to validate."
