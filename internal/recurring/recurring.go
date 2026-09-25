// Package recurring generates monthly entries from templates.
package recurring

import (
	"context"
	"database/sql"

	"github.com/melojms/mm-budget/internal/clock"
)

// Generate creates entries for every active template for each month from the
// template's start_month up to the current month (inclusive), skipping months
// already present in template_runs. Returns the number of entries created.
func Generate(ctx context.Context, db *sql.DB, c clock.Clock) (int, error) {
	_, _, _ = ctx, db, c
	return 0, nil // TODO(be-ops): implement
}
