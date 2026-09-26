package api

import (
	"database/sql"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func opsSeedExport(t *testing.T, s *Server) {
	t.Helper()
	ctx := t.Context()
	mortgage := opsInsertEntry(t, s, "expense", "2026-03-01", 80000, 51, "confirmed")
	opsInsertEntry(t, s, "income", "2026-02-25", 250050, 100, "pending")
	groceries := opsInsertEntry(t, s, "expense", "2026-03-02", 1250, 6, "confirmed")
	opsInsertEntry(t, s, "investment", "2026-04-01", 10000, 200, "confirmed")
	if _, err := s.DB.ExecContext(ctx, `UPDATE entries SET template_month = '2026-03', payer_id = 3 WHERE id = ?`, mortgage); err != nil {
		t.Fatal(err)
	}
	if _, err := s.DB.ExecContext(ctx, `UPDATE entries SET note = 'Lidl; "big" shop' WHERE id = ?`, groceries); err != nil {
		t.Fatal(err)
	}
	if _, err := s.DB.ExecContext(ctx, `INSERT INTO tags (id, name) VALUES (1, 'weekly'), (2, 'food'); INSERT INTO entry_tags VALUES (?, 1), (?, 2)`, groceries, groceries); err != nil {
		t.Fatal(err)
	}
}

const opsCSVHeader = "\xef\xbb\xbfdate;type;category;subcategory;amount;payer;note;tags;recurring;status\r\n"

func TestExportCSV(t *testing.T) {
	s, h := newTestServer(t)
	opsSeedExport(t, s)

	var (
		income    = "2026-02-25;income;Salary;;2500,50;Person A;;;no;pending\r\n"
		mortgage  = "2026-03-01;expense;House;Mortgage;800,00;Joint;;;yes;confirmed\r\n"
		groceries = "2026-03-02;expense;Groceries;;12,50;Person A;\"Lidl; \"\"big\"\" shop\";food,weekly;no;confirmed\r\n"
		etf       = "2026-04-01;investment;ETFs/Stocks;;100,00;Person A;;;no;confirmed\r\n"
	)
	tests := []struct {
		name, query, filename, body string
	}{
		{"everything", "", "mealheiro_all_all.csv", income + mortgage + groceries + etf},
		{"date range", "?from=2026-03-01&to=2026-03-31", "mealheiro_2026-03-01_2026-03-31.csv", mortgage + groceries},
		{"open start", "?to=2026-03-01", "mealheiro_all_2026-03-01.csv", income + mortgage},
		{"types", "?types=income,investment", "mealheiro_all_all.csv", income + etf},
		{"empty range", "?from=2030-01-01", "mealheiro_2030-01-01_all.csv", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := do(t, h, "GET", "/api/export.csv"+tt.query, nil)
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d %s", rec.Code, rec.Body)
			}
			if ct := rec.Header().Get("Content-Type"); ct != "text/csv; charset=utf-8" {
				t.Errorf("content-type = %s", ct)
			}
			if cd := rec.Header().Get("Content-Disposition"); cd != `attachment; filename="`+tt.filename+`"` {
				t.Errorf("disposition = %s", cd)
			}
			if got, want := rec.Body.String(), opsCSVHeader+tt.body; got != want {
				t.Fatalf("body\n%q\nwant\n%q", got, want)
			}
		})
	}
}

func TestExportCSVValidation(t *testing.T) {
	_, h := newTestServer(t)
	for _, q := range []string{"?from=2026-3-01", "?to=yesterday", "?from=2026-04-01&to=2026-03-01", "?types=expense,gift", "?types=,"} {
		rec := do(t, h, "GET", "/api/export.csv"+q, nil)
		if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "error") {
			t.Errorf("%s = %d %s", q, rec.Code, rec.Body)
		}
	}
}

func TestBackupDownload(t *testing.T) {
	s, h := newTestServer(t)
	opsInsertEntry(t, s, "expense", "2026-03-02", 1250, 6, "confirmed")

	rec := do(t, h, "GET", "/api/backup", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d %s", rec.Code, rec.Body)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/octet-stream" {
		t.Errorf("content-type = %s", ct)
	}
	if cd := rec.Header().Get("Content-Disposition"); cd != `attachment; filename="mealheiro-20260315-120000.db"` {
		t.Errorf("disposition = %s", cd)
	}
	if !strings.HasPrefix(rec.Body.String(), "SQLite format 3\x00") {
		t.Fatal("body is not a SQLite database")
	}

	// The snapshot opens and contains the data.
	path := filepath.Join(t.TempDir(), "dl.db")
	if err := os.WriteFile(path, rec.Body.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", "file:"+path+"?mode=ro")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	var n int
	if err := db.QueryRowContext(t.Context(), `SELECT count(*) FROM entries`).Scan(&n); err != nil || n != 1 {
		t.Fatalf("entries = %d %v", n, err)
	}

	// Temp files are cleaned up.
	if left, _ := os.ReadDir(s.BackupDir); len(left) != 0 {
		t.Fatalf("leftovers in backup dir: %v", left)
	}
}
