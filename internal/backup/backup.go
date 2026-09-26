// Package backup snapshots the SQLite database with VACUUM INTO.
package backup

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// Snapshot writes a consistent copy of the database to dst (must not exist).
func Snapshot(ctx context.Context, db *sql.DB, dst string) error {
	if _, err := os.Stat(dst); err == nil {
		return fmt.Errorf("snapshot %s: %w", dst, fs.ErrExist)
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	if _, err := db.ExecContext(ctx, "VACUUM INTO ?", dst); err != nil {
		return fmt.Errorf("vacuum into %s: %w", dst, err)
	}
	return nil
}

// Nightly writes dir/mm-budget-YYYYMMDD.db for today (if missing) and keeps
// only the newest keep snapshots.
func Nightly(ctx context.Context, db *sql.DB, dir, today string, keep int) error {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("create backup dir: %w", err)
	}
	name := filepath.Join(dir, "mm-budget-"+strings.ReplaceAll(today, "-", "")+".db")
	if _, err := os.Stat(name); errors.Is(err, fs.ErrNotExist) {
		// Write to a hidden temp name first so a crash never leaves a partial
		// file that would be mistaken for today's backup.
		tmp := filepath.Join(dir, ".tmp-"+filepath.Base(name))
		_ = os.Remove(tmp)
		if err := Snapshot(ctx, db, tmp); err != nil {
			_ = os.Remove(tmp)
			return err
		}
		if err := os.Rename(tmp, name); err != nil {
			_ = os.Remove(tmp)
			return err
		}
	} else if err != nil {
		return err
	}
	return prune(dir, keep)
}

// prune deletes all but the newest keep mm-budget-*.db files (names sort by date).
func prune(dir string, keep int) error {
	if keep <= 0 {
		return nil
	}
	files, err := filepath.Glob(filepath.Join(dir, "mm-budget-*.db"))
	if err != nil {
		return err
	}
	if len(files) <= keep {
		return nil
	}
	slices.Sort(files)
	var errs []error
	for _, f := range files[:len(files)-keep] {
		errs = append(errs, os.Remove(f))
	}
	return errors.Join(errs...)
}
