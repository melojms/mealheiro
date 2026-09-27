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
