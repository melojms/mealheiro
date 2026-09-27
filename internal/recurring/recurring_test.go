package recurring

import (
	"database/sql"
	"testing"
	"time"

	"github.com/melojms/mealheiro/internal/clock"
	"github.com/melojms/mealheiro/internal/store"
)

var lisbon = func() *time.Location {
	loc, err := time.LoadLocation("Europe/Lisbon")
	if err != nil {
		panic(err)
	}
	return loc
}()

func at(y int, m time.Month, d int) clock.Clock {
	return clock.Fixed{T: time.Date(y, m, d, 12, 0, 0, 0, lisbon)}
}

type genEntry struct {
	Date, Status, Month string
	Amount              int64
}

func addTemplate(t *testing.T, db *sql.DB, typ string, cat int64, amount int64, variable bool, start string, end *string, active bool) int64 {
	t.Helper()
	id, err := store.New(db).CreateTemplate(t.Context(), store.CreateTemplateParams{
		Type: typ, CategoryID: cat, PayerID: 3, AmountCents: amount, Variable: variable,
		Note: "n", StartMonth: start, EndMonth: end, Active: active,
	})
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func entriesOf(t *testing.T, db *sql.DB, templateID int64) []genEntry {
	t.Helper()
	rows, err := db.QueryContext(t.Context(), `SELECT date, status, template_month, amount_cents FROM entries WHERE template_id = ? ORDER BY date`, templateID)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	var out []genEntry
	for rows.Next() {
		var e genEntry
		if err := rows.Scan(&e.Date, &e.Status, &e.Month, &e.Amount); err != nil {
			t.Fatal(err)
		}
		out = append(out, e)
	}
	return out
}

func ptr(s string) *string { return &s }

func TestGenerate(t *testing.T) {
	tests := []struct {
		name     string
		typ      string
		cat      int64
		variable bool
		start    string
		end      *string
		active   bool
		want     []genEntry
	}{
		{
			name: "fixed catches up to current month", typ: "expense", cat: 51, start: "2026-01", active: true,
			want: []genEntry{
				{"2026-01-01", "confirmed", "2026-01", 1000},
				{"2026-02-01", "confirmed", "2026-02", 1000},
				{"2026-03-01", "confirmed", "2026-03", 1000},
			},
		},
		{
			name: "variable is pending", typ: "income", cat: 100, variable: true, start: "2026-03", active: true,
			want: []genEntry{{"2026-03-01", "pending", "2026-03", 1000}},
		},
		{
			name: "end month in the past stops generation", typ: "investment", cat: 200, start: "2025-12", end: ptr("2026-01"), active: true,
			want: []genEntry{
				{"2025-12-01", "confirmed", "2025-12", 1000},
				{"2026-01-01", "confirmed", "2026-01", 1000},
			},
		},
		{name: "future start generates nothing", typ: "expense", cat: 7, start: "2026-04", active: true},
		{name: "inactive generates nothing", typ: "expense", cat: 7, start: "2026-01", active: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := store.OpenTest(t)
			id := addTemplate(t, db, tt.typ, tt.cat, 1000, tt.variable, tt.start, tt.end, tt.active)
			n, err := Generate(t.Context(), db, at(2026, 3, 15))
			if err != nil {
				t.Fatal(err)
			}
			got := entriesOf(t, db, id)
			if n != len(tt.want) || len(got) != len(tt.want) {
				t.Fatalf("created %d, entries %+v, want %+v", n, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("entry %d = %+v, want %+v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestGenerateIdempotentEvenAfterDelete(t *testing.T) {
	db := store.OpenTest(t)
	id := addTemplate(t, db, "expense", 7, 999, false, "2026-02", nil, true)
	c := at(2026, 3, 15)
	if n, err := Generate(t.Context(), db, c); err != nil || n != 2 {
		t.Fatalf("first run = %d, %v", n, err)
	}
	if n, err := Generate(t.Context(), db, c); err != nil || n != 0 {
		t.Fatalf("second run = %d, %v", n, err)
	}
	if _, err := db.ExecContext(t.Context(), `DELETE FROM entries WHERE template_month = '2026-02'`); err != nil {
		t.Fatal(err)
	}
	if n, err := Generate(t.Context(), db, c); err != nil || n != 0 {
		t.Fatalf("after delete = %d, %v", n, err)
	}
	if got := entriesOf(t, db, id); len(got) != 1 || got[0].Month != "2026-03" {
		t.Fatalf("entries = %+v", got)
	}
	// Next month the job picks up only the new month.
	if n, err := Generate(t.Context(), db, at(2026, 4, 1)); err != nil || n != 1 {
		t.Fatalf("april run = %d, %v", n, err)
	}
}

func TestGenerateVariableUsesLatestAmount(t *testing.T) {
	db := store.OpenTest(t)
	id := addTemplate(t, db, "expense", 2, 5000, true, "2026-01", nil, true)
	if _, err := Generate(t.Context(), db, at(2026, 1, 10)); err != nil {
		t.Fatal(err)
	}
	// User confirms January with the real bill.
	if _, err := db.ExecContext(t.Context(), `UPDATE entries SET amount_cents = 6234, status = 'confirmed' WHERE template_id = ?`, id); err != nil {
		t.Fatal(err)
	}
	if n, err := Generate(t.Context(), db, at(2026, 3, 2)); err != nil || n != 2 {
		t.Fatalf("catch-up = %d, %v", n, err)
	}
	want := []genEntry{
		{"2026-01-01", "confirmed", "2026-01", 6234},
		{"2026-02-01", "pending", "2026-02", 6234},
		{"2026-03-01", "pending", "2026-03", 6234},
	}
	got := entriesOf(t, db, id)
	if len(got) != len(want) {
		t.Fatalf("entries = %+v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("entry %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestGenerateCopiesTemplateFields(t *testing.T) {
	db := store.OpenTest(t)
	id := addTemplate(t, db, "expense", 51, 80000, false, "2026-03", nil, true)
	if _, err := Generate(t.Context(), db, at(2026, 3, 15)); err != nil {
		t.Fatal(err)
	}
	var typ, note string
	var cat, payer int64
	err := db.QueryRowContext(t.Context(), `SELECT type, category_id, payer_id, note FROM entries WHERE template_id = ?`, id).Scan(&typ, &cat, &payer, &note)
	if err != nil {
		t.Fatal(err)
	}
	if typ != "expense" || cat != 51 || payer != 3 || note != "n" {
		t.Fatalf("got %s %d %d %q", typ, cat, payer, note)
	}
	var runs int
	if err := db.QueryRowContext(t.Context(), `SELECT count(*) FROM template_runs WHERE template_id = ?`, id).Scan(&runs); err != nil || runs != 1 {
		t.Fatalf("runs = %d, %v", runs, err)
	}
}

func TestGenerateCopiesPersonalFlag(t *testing.T) {
	db := store.OpenTest(t)
	id, err := store.New(db).CreateTemplate(t.Context(), store.CreateTemplateParams{
		Type: "expense", CategoryID: 51, PayerID: 1, Personal: true, AmountCents: 1000,
		StartMonth: "2026-03", Active: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Generate(t.Context(), db, at(2026, 3, 15)); err != nil {
		t.Fatal(err)
	}
	var personal bool
	if err := db.QueryRowContext(t.Context(), `SELECT personal FROM entries WHERE template_id = ?`, id).Scan(&personal); err != nil {
		t.Fatal(err)
	}
	if !personal {
		t.Fatal("generated entry personal = false, want true")
	}
}
