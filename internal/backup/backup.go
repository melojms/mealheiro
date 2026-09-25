// Package backup snapshots the SQLite database with VACUUM INTO.
package backup

import (
	"context"
	"database/sql"
)

// Snapshot writes a consistent copy of the database to dst (must not exist).
func Snapshot(ctx context.Context, db *sql.DB, dst string) error {
	_, _, _ = ctx, db, dst
	return nil // TODO(be-ops): implement
}

// Nightly writes dir/mm-budget-YYYYMMDD.db for today (if missing) and keeps
// only the newest keep snapshots.
func Nightly(ctx context.Context, db *sql.DB, dir, today string, keep int) error {
	_, _, _, _, _ = ctx, db, dir, today, keep
	return nil // TODO(be-ops): implement
}
