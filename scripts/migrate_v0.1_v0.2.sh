#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DB_PATH="${HISTREE_DB:-$HOME/.histree.db}"

if ! command -v sqlite3 &> /dev/null; then
    echo "Error: sqlite3 is not installed. Please install it with your system package manager."
    exit 1
fi

# Check if database exists
if [ ! -f "$DB_PATH" ]; then
    echo "Database file not found at $DB_PATH"
    exit 1
fi

# Check if the table exists
if ! sqlite3 "$DB_PATH" "SELECT name FROM sqlite_master WHERE type='table' AND name='history';" | grep -q "history"; then
    echo "history table not found in database"
    exit 1
fi

# Create backup
backup_path="${DB_PATH}.bak.$(date +%Y%m%d_%H%M%S)"
cp "$DB_PATH" "$backup_path"
echo "Created backup at $backup_path"

# Check if migration is needed by looking at the table schema
if ! sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM pragma_table_info('history') WHERE name='session_label';" | grep -q "1"; then
    echo "Migration already applied (session_label column not found)"
    exit 0
fi

echo "Applying migration..."
if ! sqlite3 "$DB_PATH" < "$SCRIPT_DIR/../migration/v0.1_v0.2.sql"; then
    echo "Migration failed. Restoring backup..."
    cp "$backup_path" "$DB_PATH"
    echo "Backup restored from $backup_path"
    exit 1
fi

echo "Verifying migration..."
if sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM pragma_table_info('history') WHERE name IN ('hostname', 'process_id');" | grep -q "2"; then
    echo "Migration completed successfully"
else
    echo "Migration verification failed. Restoring backup..."
    cp "$backup_path" "$DB_PATH"
    echo "Backup restored from $backup_path"
    exit 1
fi