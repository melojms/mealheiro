package backup

import (
	"database/sql"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/melojms/mealheiro/internal/store"
)

func countPeople(t *testing.T, path string) int {
	t.Helper()
	db, err := sql.Open("sqlite", "file:"+path+"?mode=ro")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	var n int
	if err := db.QueryRowContext(t.Context(), `SELECT count(*) FROM people`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestSnapshot(t *testing.T) {
	db := store.OpenTest(t)
	dst := filepath.Join(t.TempDir(), "snap.db")
	if err := Snapshot(t.Context(), db, dst); err != nil {
		t.Fatal(err)
	}
	if n := countPeople(t, dst); n != 3 {
		t.Fatalf("people in snapshot = %d", n)
	}
	if err := Snapshot(t.Context(), db, dst); !errors.Is(err, fs.ErrExist) {
		t.Fatalf("second snapshot err = %v, want ErrExist", err)
	}
}

func listDir(t *testing.T, dir string) []string {
	t.Helper()
	des, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, d := range des {
		names = append(names, d.Name())
	}
	return names
}

func TestNightly(t *testing.T) {
	db := store.OpenTest(t)
	dir := filepath.Join(t.TempDir(), "backups") // created on demand

	if err := Nightly(t.Context(), db, dir, "2026-03-15", 3); err != nil {
		t.Fatal(err)
	}
	today := filepath.Join(dir, "mealheiro-20260315.db")
	if n := countPeople(t, today); n != 3 {
		t.Fatalf("people = %d", n)
	}

	// Same day again: existing file is kept untouched.
	before, _ := os.Stat(today)
	if err := Nightly(t.Context(), db, dir, "2026-03-15", 3); err != nil {
		t.Fatal(err)
	}
	after, _ := os.Stat(today)
	if !before.ModTime().Equal(after.ModTime()) {
		t.Fatal("today's backup was rewritten")
	}

	// Older snapshots plus unrelated files; only mealheiro-*.db are pruned.
	for _, name := range []string{"mealheiro-20260101.db", "mealheiro-20260201.db", "mealheiro-20260301.db", "notes.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := Nightly(t.Context(), db, dir, "2026-03-16", 3); err != nil {
		t.Fatal(err)
	}
	want := []string{"mealheiro-20260301.db", "mealheiro-20260315.db", "mealheiro-20260316.db", "notes.txt"}
	if got := listDir(t, dir); !slices.Equal(got, want) {
		t.Fatalf("dir = %v, want %v", got, want)
	}
}

func TestNightlyKeepZeroKeepsAll(t *testing.T) {
	db := store.OpenTest(t)
	dir := t.TempDir()
	for _, day := range []string{"2026-03-14", "2026-03-15"} {
		if err := Nightly(t.Context(), db, dir, day, 0); err != nil {
			t.Fatal(err)
		}
	}
	if got := listDir(t, dir); len(got) != 2 {
		t.Fatalf("dir = %v", got)
	}
}
