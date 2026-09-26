package api

import (
	"math"
	"net/http"
	"testing"
)

func opsI64(n int64) *int64 { return &n }

// insertBudget adds a budget row directly (to set up past months).
func opsInsertBudget(t *testing.T, s *Server, categoryID *int64, from string, cents *int64) {
	t.Helper()
	if _, err := s.DB.ExecContext(t.Context(), `INSERT INTO budgets (category_id, effective_from, amount_cents) VALUES (?, ?, ?)`, categoryID, from, cents); err != nil {
		t.Fatal(err)
	}
}

func TestPutBudgetValidation(t *testing.T) {
	s, h := newTestServer(t)
	if _, err := s.DB.ExecContext(t.Context(), `UPDATE categories SET archived = TRUE WHERE id = 19`); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name string
		body any
		want int
	}{
		{"top-level expense", map[string]any{"category_id": 6, "amount_cents": 40000}, 204},
		{"overall cap", map[string]any{"category_id": nil, "amount_cents": 200000}, 204},
		{"remove", map[string]any{"category_id": 6, "amount_cents": nil}, 204},
		{"subcategory", map[string]any{"category_id": 51, "amount_cents": 100}, 400},
		{"income category", map[string]any{"category_id": 100, "amount_cents": 100}, 400},
		{"unknown category", map[string]any{"category_id": 999, "amount_cents": 100}, 400},
		{"zero amount", map[string]any{"category_id": 6, "amount_cents": 0}, 400},
		{"negative amount", map[string]any{"category_id": nil, "amount_cents": -1}, 400},
		{"archived set", map[string]any{"category_id": 19, "amount_cents": 100}, 400},
		{"archived remove", map[string]any{"category_id": 19, "amount_cents": nil}, 204},
		{"unknown field", map[string]any{"category_id": 6, "amount": 1}, 400},
		{"empty body", nil, 400},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if rec := do(t, h, "PUT", "/api/budgets", tt.body); rec.Code != tt.want {
				t.Fatalf("status = %d %s", rec.Code, rec.Body)
			}
		})
	}
}

func TestPutBudgetVersioning(t *testing.T) {
	s, h := newTestServer(t)
	opsInsertBudget(t, s, opsI64(6), "2026-01", opsI64(30000))

	// Setting twice in the same month replaces the row.
	for _, amount := range []int64{40000, 45000} {
		if rec := do(t, h, "PUT", "/api/budgets", map[string]any{"category_id": 6, "amount_cents": amount}); rec.Code != 204 {
			t.Fatalf("put = %d", rec.Code)
		}
	}
	budgetOf := func(month string) *BudgetLine {
		resp := decode[budgetsResponse](t, do(t, h, "GET", "/api/budgets?month="+month, nil))
		for _, l := range resp.Categories {
			if *l.CategoryID == 6 {
				return &l
			}
		}
		return nil
	}
	if l := budgetOf("2026-02"); l == nil || l.BudgetCents != 30000 {
		t.Fatalf("february = %+v (history must be kept)", l)
	}
	if l := budgetOf("2026-03"); l == nil || l.BudgetCents != 45000 {
		t.Fatalf("march = %+v", l)
	}
	if l := budgetOf("2027-01"); l == nil || l.BudgetCents != 45000 {
		t.Fatalf("future = %+v", l)
	}

	// Removing keeps the past but hides it from March on.
	if rec := do(t, h, "PUT", "/api/budgets", map[string]any{"category_id": 6, "amount_cents": nil}); rec.Code != 204 {
		t.Fatalf("remove = %d", rec.Code)
	}
	if l := budgetOf("2026-03"); l != nil {
		t.Fatalf("march after remove = %+v", l)
	}
	if l := budgetOf("2026-02"); l == nil || l.BudgetCents != 30000 {
		t.Fatalf("february after remove = %+v", l)
	}
}

func TestPutBudgetRemoveWithoutHistoryLeavesNoRow(t *testing.T) {
	s, h := newTestServer(t)
	do(t, h, "PUT", "/api/budgets", map[string]any{"category_id": 6, "amount_cents": 100})
	do(t, h, "PUT", "/api/budgets", map[string]any{"category_id": 6, "amount_cents": nil})
	do(t, h, "PUT", "/api/budgets", map[string]any{"category_id": nil, "amount_cents": nil})
	var n int
	if err := s.DB.QueryRowContext(t.Context(), `SELECT count(*) FROM budgets`).Scan(&n); err != nil || n != 0 {
		t.Fatalf("rows = %d %v", n, err)
	}
}

func TestListBudgets(t *testing.T) {
	s, h := newTestServer(t)
	opsInsertBudget(t, s, nil, "2026-01", opsI64(100000))
	opsInsertBudget(t, s, opsI64(6), "2026-01", opsI64(40000))  // Groceries
	opsInsertBudget(t, s, opsI64(5), "2026-02", opsI64(100000)) // House
	opsInsertBudget(t, s, opsI64(12), "2026-01", opsI64(10000)) // Eating out, removed in March
	opsInsertBudget(t, s, opsI64(12), "2026-03", nil)

	opsInsertEntry(t, s, "expense", "2026-03-02", 30000, 6, "confirmed")
	opsInsertEntry(t, s, "expense", "2026-03-20", 2000, 6, "pending")
	opsInsertEntry(t, s, "expense", "2026-03-01", 80000, 51, "confirmed") // Mortgage → House
	opsInsertEntry(t, s, "expense", "2026-03-05", 5000, 5, "pending")     // House itself
	opsInsertEntry(t, s, "expense", "2026-03-06", 1500, 12, "confirmed")  // no budget
	opsInsertEntry(t, s, "expense", "2026-02-28", 9999, 6, "confirmed")   // other month
	opsInsertEntry(t, s, "income", "2026-03-01", 300000, 100, "confirmed")

	rec := do(t, h, "GET", "/api/budgets", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	resp := decode[budgetsResponse](t, rec)
	if resp.Month != "2026-03" {
		t.Fatalf("month = %s", resp.Month)
	}
	o := resp.Overall
	if o == nil || o.CategoryID != nil || o.Name != "Overall" || o.Icon != "wallet" || o.Color != "#64748b" ||
		o.BudgetCents != 100000 || o.SpentCents != 118500 || o.PendingCents != 7000 || math.Abs(o.Ratio-1.185) > 1e-9 {
		t.Fatalf("overall = %+v", o)
	}
	if len(resp.Categories) != 2 {
		t.Fatalf("categories = %+v", resp.Categories)
	}
	house, groceries := resp.Categories[0], resp.Categories[1]
	if house.Name != "House" || house.Icon != "house" || house.SpentCents != 85000 || house.PendingCents != 5000 || house.Ratio != 0.85 {
		t.Fatalf("house = %+v", house)
	}
	if groceries.Name != "Groceries" || groceries.SpentCents != 32000 || groceries.PendingCents != 2000 || groceries.Ratio != 0.8 {
		t.Fatalf("groceries = %+v", groceries)
	}

	// January: House budget not yet effective, Eating out still is.
	resp = decode[budgetsResponse](t, do(t, h, "GET", "/api/budgets?month=2026-01", nil))
	if len(resp.Categories) != 2 || resp.Categories[0].Name != "Eating out" || resp.Categories[0].SpentCents != 0 {
		t.Fatalf("january = %+v", resp.Categories)
	}
}

func TestListBudgetsEmptyAndErrors(t *testing.T) {
	_, h := newTestServer(t)
	rec := do(t, h, "GET", "/api/budgets?month=2026-03", nil)
	if rec.Code != 200 {
		t.Fatalf("status = %d", rec.Code)
	}
	if got := rec.Body.String(); got != `{"month":"2026-03","overall":null,"categories":[]}`+"\n" {
		t.Fatalf("body = %s", got)
	}
	for _, q := range []string{"2026-3", "march", "2026-13"} {
		if rec := do(t, h, "GET", "/api/budgets?month="+q, nil); rec.Code != 400 {
			t.Errorf("month=%s = %d", q, rec.Code)
		}
	}
}

func TestBudgetStatus(t *testing.T) {
	s, h := newTestServer(t)
	opsInsertBudget(t, s, nil, "2026-01", opsI64(200000))
	opsInsertBudget(t, s, opsI64(5), "2026-01", opsI64(100000))
	opsInsertEntry(t, s, "expense", "2026-03-01", 80000, 51, "confirmed")

	tests := []struct {
		name         string
		query        string
		code         int
		wantCategory string // "" = null
		wantOverall  bool
	}{
		{"subcategory resolves to parent", "?category_id=51", 200, "House", true},
		{"top-level", "?category_id=5", 200, "House", true},
		{"no budget", "?category_id=6", 200, "", true},
		{"income category", "?category_id=100", 200, "", true},
		{"no category", "", 200, "", true},
		{"before budgets", "?category_id=5&month=2025-12", 200, "", false},
		{"unknown category", "?category_id=999", 404, "", false},
		{"bad category", "?category_id=x", 400, "", false},
		{"bad month", "?category_id=5&month=x", 400, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := do(t, h, "GET", "/api/budgets/status"+tt.query, nil)
			if rec.Code != tt.code {
				t.Fatalf("status = %d %s", rec.Code, rec.Body)
			}
			if tt.code != 200 {
				return
			}
			resp := decode[budgetStatusResponse](t, rec)
			if (resp.Category == nil) != (tt.wantCategory == "") || (resp.Category != nil && resp.Category.Name != tt.wantCategory) {
				t.Fatalf("category = %+v", resp.Category)
			}
			if (resp.Overall != nil) != tt.wantOverall {
				t.Fatalf("overall = %+v", resp.Overall)
			}
		})
	}
	resp := decode[budgetStatusResponse](t, do(t, h, "GET", "/api/budgets/status?category_id=51", nil))
	if resp.Category.SpentCents != 80000 || resp.Category.BudgetCents != 100000 || resp.Overall.SpentCents != 80000 {
		t.Fatalf("amounts = %+v %+v", resp.Category, resp.Overall)
	}
}
