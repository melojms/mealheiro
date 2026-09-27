// Package recurring generates monthly entries from templates.
package recurring

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/melojms/mealheiro/internal/clock"
	"github.com/melojms/mealheiro/internal/store"
)

// Generate creates entries for every active template for each month from the
// template's start_month up to the current month (inclusive), skipping months
// already present in template_runs. Returns the number of entries created.
func Generate(ctx context.Context, db *sql.DB, c clock.Clock) (int, error) {
	templates, err := store.New(db).ListActiveTemplates(ctx)
	if err != nil {
		return 0, fmt.Errorf("list templates: %w", err)
	}
	current := clock.CurrentMonth(c)
	created := 0
	for _, t := range templates {
		n, err := generateTemplate(ctx, db, t, current)
		created += n
		if err != nil {
			return created, fmt.Errorf("template %d: %w", t.ID, err)
		}
	}
	return created, nil
}

// generateTemplate catches up one template in its own transaction.
func generateTemplate(ctx context.Context, db *sql.DB, t store.Template, current string) (int, error) {
	last := current
	if t.EndMonth != nil && *t.EndMonth < last {
		last = *t.EndMonth
	}
	months, err := clock.MonthsBetween(t.StartMonth, last)
	if err != nil || len(months) == 0 {
		return 0, err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()
	q := store.New(tx)

	done, err := q.ListTemplateRunMonths(ctx, t.ID)
	if err != nil {
		return 0, err
	}
	seen := make(map[string]bool, len(done))
	for _, m := range done {
		seen[m] = true
	}

	status := "confirmed"
	if t.Variable {
		status = "pending"
	}
	created := 0
	for _, month := range months {
		if seen[month] {
			continue
		}
		amount := t.AmountCents
		if t.Variable {
			// Prefill with the most recent entry of this template (incl. ones just generated).
			switch last, err := q.LatestTemplateEntryAmount(ctx, &t.ID); {
			case err == nil:
				amount = last
			case !errors.Is(err, sql.ErrNoRows):
				return 0, err
			}
		}
		m := month
		if err := q.InsertGeneratedEntry(ctx, store.InsertGeneratedEntryParams{
			Type:          t.Type,
			Date:          month + "-01",
			AmountCents:   amount,
			CategoryID:    t.CategoryID,
			PayerID:       t.PayerID,
			Personal:      t.Personal,
			Note:          t.Note,
			Status:        status,
			TemplateID:    &t.ID,
			TemplateMonth: &m,
		}); err != nil {
			return 0, err
		}
		if err := q.InsertTemplateRun(ctx, store.InsertTemplateRunParams{TemplateID: t.ID, Month: month}); err != nil {
			return 0, err
		}
		created++
	}
	if created == 0 {
		return 0, nil
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return created, nil
}
