package store

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/pressly/goose/v3"
)

func TestMigrateRenamesCondominioToManagement(t *testing.T) {
	db := OpenTest(t)

	cat, err := New(db).GetCategory(context.Background(), 52)
	if err != nil {
		t.Fatalf("get category: %v", err)
	}
	if cat.Name != "Management" {
		t.Fatalf("category 52 name = %q, want %q", cat.Name, "Management")
	}
}

func TestMigrateKeepsUserRenamedCondominio(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	provider, err := goose.NewProvider(goose.DialectSQLite3, db, mustSub(migrations, "migrations"))
	if err != nil {
		t.Fatalf("goose provider: %v", err)
	}
	if _, err := provider.DownTo(ctx, 2); err != nil {
		t.Fatalf("down to seed: %v", err)
	}
	if _, err := db.ExecContext(ctx, `UPDATE categories SET name = 'Condo fees' WHERE id = 52`); err != nil {
		t.Fatalf("rename: %v", err)
	}
	if err := Migrate(ctx, db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	cat, err := New(db).GetCategory(ctx, 52)
	if err != nil {
		t.Fatalf("get category: %v", err)
	}
	if cat.Name != "Condo fees" {
		t.Fatalf("category 52 name = %q, want %q", cat.Name, "Condo fees")
	}
}

func TestMigrateAddsPersonalFlagDefaultingToShared(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	provider, err := goose.NewProvider(goose.DialectSQLite3, db, mustSub(migrations, "migrations"))
	if err != nil {
		t.Fatalf("goose provider: %v", err)
	}
	if _, err := provider.DownTo(ctx, 3); err != nil {
		t.Fatalf("down to 3: %v", err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO entries (id, type, date, amount_cents, category_id, payer_id) VALUES (1, 'expense', '2026-09-01', 100, 50, 1)`); err != nil {
		t.Fatalf("insert entry: %v", err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO templates (id, type, category_id, payer_id, amount_cents, start_month) VALUES (1, 'expense', 50, 1, 100, '2026-09')`); err != nil {
		t.Fatalf("insert template: %v", err)
	}
	if err := Migrate(ctx, db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	q := New(db)
	v, err := q.GetEntryView(ctx, 1)
	if err != nil {
		t.Fatalf("get entry view: %v", err)
	}
	if v.Personal {
		t.Fatalf("existing entry personal = true, want false")
	}
	if v.PayerKind != "person" {
		t.Fatalf("payer_kind = %q, want %q", v.PayerKind, "person")
	}
	tpl, err := q.GetTemplate(ctx, 1)
	if err != nil {
		t.Fatalf("get template: %v", err)
	}
	if tpl.Personal {
		t.Fatalf("existing template personal = true, want false")
	}
}
