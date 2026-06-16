#!/usr/bin/env bash
set -euo pipefail

# Example:
# mysql -u root -p < migrations/001_create_transactions.sql
mysql -u "${DB_USER:-root}" -p"${DB_PASSWORD:-}" -h "${DB_HOST:-127.0.0.1}" -P "${DB_PORT:-3306}" < migrations/001_create_transactions.sql
