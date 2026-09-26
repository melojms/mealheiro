package api

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/melojms/mm-budget/internal/reports"
)

// seedReports inserts several months of entries (now = 2026-03-15):
//
//	2025-06  Travel 1000,00 (B)
//	2025-12  Groceries 300 (A) · Salary 3000 (A)
//	2026-01  Groceries 300 (A) · Mortgage 800 recurring (Joint) · Salary 3000 (A) · ETFs 500 (B)
//	2026-02  Groceries 300 (B) · Mortgage 800 recurring · Eating out 50 (A) · Salary 3000 (A)
//	2026-03  Groceries 450 (A, 10th) · House itself 20 (Joint, 2nd) · Mortgage 800 recurring
//	         Electricity 60 pending recurring (template 1) · Eating out 30 (A, 20th, after today)
//	         Salary 3000 (A)
func seedReports(t *testing.T, s *Server) {
	t.Helper()
	stmts := []string{
		`INSERT INTO templates (id, type, category_id, payer_id, amount_cents, variable, start_month)
		 VALUES (1, 'expense', 2, 3, 6000, TRUE, '2026-03'),
		        (2, 'expense', 51, 3, 80000, FALSE, '2026-01')`,
		`INSERT INTO entries (id, type, date, amount_cents, category_id, payer_id, note, status, template_id, template_month) VALUES
		 (1,  'expense', '2025-06-10', 100000, 15, 2, 'Rome', 'confirmed', NULL, NULL),
		 (2,  'expense', '2025-12-05', 30000,  6,  1, '', 'confirmed', NULL, NULL),
		 (3,  'income',  '2025-12-25', 300000, 100, 1, '', 'confirmed', NULL, NULL),
		 (4,  'expense', '2026-01-07', 30000,  6,  1, '', 'confirmed', NULL, NULL),
		 (5,  'expense', '2026-01-01', 80000,  51, 3, '', 'confirmed', 2, '2026-01'),
		 (6,  'income',  '2026-01-25', 300000, 100, 1, '', 'confirmed', NULL, NULL),
		 (7,  'investment', '2026-01-26', 50000, 200, 2, '', 'confirmed', NULL, NULL),
		 (8,  'expense', '2026-02-09', 30000,  6,  2, '', 'confirmed', NULL, NULL),
		 (9,  'expense', '2026-02-01', 80000,  51, 3, '', 'confirmed', 2, '2026-02'),
		 (10, 'expense', '2026-02-14', 5000,   12, 1, 'dinner', 'confirmed', NULL, NULL),
		 (11, 'income',  '2026-02-25', 300000, 100, 1, '', 'confirmed', NULL, NULL),
		 (12, 'expense', '2026-03-10', 45000,  6,  1, 'big shop', 'confirmed', NULL, NULL),
		 (13, 'expense', '2026-03-02', 2000,   5,  3, '', 'confirmed', NULL, NULL),
		 (14, 'expense', '2026-03-01', 80000,  51, 3, '', 'confirmed', 2, '2026-03'),
		 (15, 'expense', '2026-03-01', 6000,   2,  3, '', 'pending', 1, '2026-03'),
		 (16, 'expense', '2026-03-20', 3000,   12, 1, '', 'confirmed', NULL, NULL),
		 (17, 'income',  '2026-03-25', 300000, 100, 1, '', 'confirmed', NULL, NULL)`,
	}
	for _, q := range stmts {
		if _, err := s.DB.ExecContext(t.Context(), q); err != nil {
			t.Fatal(err)
		}
	}
}

func TestReportsValidation(t *testing.T) {
	_, h := newTestServer(t)
	paths := []string{
		"/api/reports/month?month=2026-13",
		"/api/reports/month?month=2026-3",
		"/api/reports/month?month=march",
		"/api/reports/month?payer_id=abc",
		"/api/reports/month?payer_id=0",
		"/api/reports/month?payer_id=-1",
		"/api/reports/month?payer_id=99",
		"/api/reports/trends?end=2026",
		"/api/reports/trends?months=0",
		"/api/reports/trends?months=121",
		"/api/reports/trends?months=x",
		"/api/reports/trends?category_id=x",
		"/api/reports/trends?category_id=51",  // subcategory
		"/api/reports/trends?category_id=100", // income
		"/api/reports/trends?category_id=999", // unknown
		"/api/reports/trends?payer_id=99",
		"/api/reports/year?year=abc",
		"/api/reports/year?year=99",
		"/api/reports/year?year=10000",
		"/api/reports/year?payer_id=x",
		"/api/reports/years?payer_id=99",
		"/api/insights?month=2026-00",
		"/api/insights?payer_id=4",
	}
	for _, p := range paths {
		t.Run(p, func(t *testing.T) {
			rec := do(t, h, "GET", p, nil)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, body %s", rec.Code, rec.Body)
			}
			if body := decode[map[string]string](t, rec); body["error"] == "" {
				t.Fatalf("missing error message: %s", rec.Body)
			}
		})
	}
}

func TestMonthReport(t *testing.T) {
	s, h := newTestServer(t)
	seedReports(t, s)

	t.Run("current month, everyone", func(t *testing.T) {
		rec := do(t, h, "GET", "/api/reports/month", nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("status %d: %s", rec.Code, rec.Body)
		}
		rep := decode[reports.MonthReport](t, rec)
		if rep.Month != "2026-03" {
			t.Fatalf("month = %q", rep.Month)
		}
		wantKPIs := reports.KPIs{
			IncomeCents: 300000, ExpenseCents: 136000, LeftoverCents: 164000,
			SavingsRate: new(164000.0 / 300000.0), PendingCents: 6000,
		}
		if !reflect.DeepEqual(rep.KPIs, wantKPIs) {
			t.Fatalf("kpis = %+v", rep.KPIs)
		}
		wantPrev := reports.KPIs{
			IncomeCents: 300000, ExpenseCents: 115000, LeftoverCents: 185000,
			SavingsRate: new(185000.0 / 300000.0),
		}
		if !reflect.DeepEqual(rep.PrevKPIs, wantPrev) {
			t.Fatalf("prev kpis = %+v", rep.PrevKPIs)
		}

		var got []string
		for _, c := range rep.Expenses {
			got = append(got, fmt.Sprintf("%s:%d/%d/%d/%d", c.Name, c.AmountCents, c.PendingCents, c.PrevMonthCents, c.Avg3Cents))
		}
		want := []string{
			"House:82000/0/80000/53333",
			"Groceries:45000/0/30000/30000",
			"Electricity:6000/6000/0/0",
			"Eating out:3000/0/5000/1667",
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("expenses = %v, want %v", got, want)
		}
		wantSubs := []reports.Subcategory{
			{CategoryID: 51, Name: "Mortgage", Color: "#8b5cf6", AmountCents: 80000},
			{CategoryID: 5, Name: "(general)", Color: "#8b5cf6", AmountCents: 2000},
		}
		if !reflect.DeepEqual(rep.Expenses[0].Subcategories, wantSubs) {
			t.Fatalf("house subcategories = %+v", rep.Expenses[0].Subcategories)
		}
		if len(rep.Income) != 1 || rep.Income[0].Name != "Salary" || rep.Income[0].Avg3Cents != 300000 {
			t.Fatalf("income = %+v", rep.Income)
		}
		if rep.Investments == nil || len(rep.Investments) != 0 {
			t.Fatalf("investments = %+v", rep.Investments)
		}
		wantRec := reports.Recurring{CommittedCents: 86000, EntryCount: 2, PendingCount: 1, PendingCents: 6000}
		if rep.Recurring != wantRec {
			t.Fatalf("recurring = %+v", rep.Recurring)
		}
	})

	t.Run("payer filter", func(t *testing.T) {
		rep := decode[reports.MonthReport](t, do(t, h, "GET", "/api/reports/month?month=2026-02&payer_id=1", nil))
		if rep.KPIs.ExpenseCents != 5000 || rep.KPIs.IncomeCents != 300000 || rep.PrevKPIs.InvestmentCents != 0 {
			t.Fatalf("kpis = %+v prev %+v", rep.KPIs, rep.PrevKPIs)
		}
		if len(rep.Expenses) != 1 || rep.Expenses[0].Name != "Eating out" {
			t.Fatalf("expenses = %+v", rep.Expenses)
		}
		if rep.Recurring != (reports.Recurring{}) {
			t.Fatalf("recurring = %+v", rep.Recurring)
		}
		joint := decode[reports.MonthReport](t, do(t, h, "GET", "/api/reports/month?month=2026-03&payer_id=3", nil))
		if joint.KPIs.ExpenseCents != 88000 || joint.KPIs.SavingsRate != nil || joint.Recurring.EntryCount != 2 {
			t.Fatalf("joint = %+v", joint)
		}
	})

	t.Run("month across year boundary", func(t *testing.T) {
		rep := decode[reports.MonthReport](t, do(t, h, "GET", "/api/reports/month?month=2026-01", nil))
		// prev = Dec 2025, avg3 = Oct–Dec 2025.
		if rep.PrevKPIs.ExpenseCents != 30000 || rep.PrevKPIs.IncomeCents != 300000 {
			t.Fatalf("prev = %+v", rep.PrevKPIs)
		}
		if g := rep.Expenses[1]; g.Name != "Groceries" || g.PrevMonthCents != 30000 || g.Avg3Cents != 10000 {
			t.Fatalf("groceries = %+v", g)
		}
		if rep.Investments[0].AmountCents != 50000 {
			t.Fatalf("investments = %+v", rep.Investments)
		}
	})

	t.Run("empty month", func(t *testing.T) {
		rec := do(t, h, "GET", "/api/reports/month?month=2020-05", nil)
		body := rec.Body.String()
		rep := decode[reports.MonthReport](t, rec)
		if rep.KPIs != (reports.KPIs{}) || len(rep.Expenses) != 0 {
			t.Fatalf("rep = %+v", rep)
		}
		for _, frag := range []string{`"expenses":[]`, `"income":[]`, `"investments":[]`, `"savings_rate":null`} {
			if !strings.Contains(body, frag) {
				t.Fatalf("body %s missing %s", body, frag)
			}
		}
	})
}

func TestTrendsReport(t *testing.T) {
	s, h := newTestServer(t)
	seedReports(t, s)

	t.Run("default window", func(t *testing.T) {
		rep := decode[reports.TrendsReport](t, do(t, h, "GET", "/api/reports/trends", nil))
		if len(rep.Months) != 12 || rep.Months[0] != "2025-04" || rep.Months[11] != "2026-03" || len(rep.Series) != 12 {
			t.Fatalf("months = %v", rep.Months)
		}
		if rep.Series[2].Month != "2025-06" || rep.Series[2].ByCategory["15"] != 100000 {
			t.Fatalf("june = %+v", rep.Series[2])
		}
	})

	t.Run("dec to feb", func(t *testing.T) {
		rec := do(t, h, "GET", "/api/reports/trends?end=2026-02&months=3", nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("status %d: %s", rec.Code, rec.Body)
		}
		rep := decode[reports.TrendsReport](t, rec)
		want := reports.TrendsReport{
			Months: []string{"2025-12", "2026-01", "2026-02"},
			Series: []reports.TrendPoint{
				{Month: "2025-12", IncomeCents: 300000, ExpenseCents: 30000, SavingsCents: 270000, ByCategory: map[string]int64{"6": 30000}},
				{Month: "2026-01", IncomeCents: 300000, ExpenseCents: 110000, InvestmentCents: 50000, SavingsCents: 190000, ByCategory: map[string]int64{"5": 80000, "6": 30000}},
				{Month: "2026-02", IncomeCents: 300000, ExpenseCents: 115000, SavingsCents: 185000, ByCategory: map[string]int64{"5": 80000, "6": 30000, "12": 5000}},
			},
			Categories: []reports.TrendCategory{
				{ID: 5, Name: "House", Color: "#8b5cf6", Icon: "house"},
				{ID: 6, Name: "Groceries", Color: "#10b981", Icon: "shopping-cart"},
				{ID: 12, Name: "Eating out", Color: "#f59e0b", Icon: "utensils"},
			},
		}
		if !reflect.DeepEqual(rep, want) {
			t.Fatalf("got\n%+v\nwant\n%+v", rep, want)
		}
	})

	t.Run("single category and payer", func(t *testing.T) {
		rep := decode[reports.TrendsReport](t, do(t, h, "GET", "/api/reports/trends?end=2026-02&months=3&category_id=6&payer_id=1", nil))
		var exp []int64
		for _, p := range rep.Series {
			exp = append(exp, p.ExpenseCents)
		}
		if !reflect.DeepEqual(exp, []int64{30000, 30000, 0}) || len(rep.Categories) != 1 || rep.Categories[0].ID != 6 {
			t.Fatalf("series = %+v categories = %+v", rep.Series, rep.Categories)
		}
		// Savings for payer A in Feb: 3000 income − 50 eating out.
		if rep.Series[2].SavingsCents != 295000 {
			t.Fatalf("savings = %d", rep.Series[2].SavingsCents)
		}
	})

	t.Run("one month", func(t *testing.T) {
		rep := decode[reports.TrendsReport](t, do(t, h, "GET", "/api/reports/trends?end=2020-01&months=1", nil))
		if len(rep.Months) != 1 || rep.Categories == nil || len(rep.Series[0].ByCategory) != 0 {
			t.Fatalf("rep = %+v", rep)
		}
	})
}

func TestYearReport(t *testing.T) {
	s, h := newTestServer(t)
	seedReports(t, s)

	rec := do(t, h, "GET", "/api/reports/year", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}
	rep := decode[reports.YearReport](t, rec)
	if rep.Year != 2026 {
		t.Fatalf("year = %d", rep.Year)
	}
	wantTotals := reports.Totals{
		IncomeCents: 900000, ExpenseCents: 361000, InvestmentCents: 50000,
		LeftoverCents: 489000, SavingsRate: new(539000.0 / 900000.0),
	}
	if !reflect.DeepEqual(rep.Totals, wantTotals) {
		t.Fatalf("totals = %+v", rep.Totals)
	}
	wantPrev := reports.Totals{IncomeCents: 300000, ExpenseCents: 130000, LeftoverCents: 170000, SavingsRate: new(170000.0 / 300000.0)}
	if !reflect.DeepEqual(rep.PrevTotals, wantPrev) {
		t.Fatalf("prev = %+v", rep.PrevTotals)
	}
	var got []string
	for _, c := range rep.Categories {
		got = append(got, fmt.Sprintf("%s/%s:%d/%d", c.Type, c.Name, c.AmountCents, c.PrevYearCents))
	}
	want := []string{
		"expense/House:242000/0",
		"expense/Groceries:105000/30000",
		"expense/Eating out:8000/0",
		"expense/Electricity:6000/0",
		"expense/Travel:0/100000",
		"income/Salary:900000/300000",
		"investment/ETFs/Stocks:50000/0",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("categories = %v", got)
	}
	if len(rep.Monthly) != 12 || rep.Monthly[2] != (reports.YearMonth{Month: "2026-03", IncomeCents: 300000, ExpenseCents: 136000}) ||
		rep.Monthly[3] != (reports.YearMonth{Month: "2026-04"}) {
		t.Fatalf("monthly = %+v", rep.Monthly)
	}

	prev := decode[reports.YearReport](t, do(t, h, "GET", "/api/reports/year?year=2025&payer_id=2", nil))
	if prev.Totals.ExpenseCents != 100000 || prev.Totals.SavingsRate != nil || len(prev.Categories) != 1 ||
		prev.Monthly[5].ExpenseCents != 100000 {
		t.Fatalf("2025/B = %+v", prev)
	}
}

func TestReportYears(t *testing.T) {
	s, h := newTestServer(t)
	if got := decode[[]int64](t, do(t, h, "GET", "/api/reports/years", nil)); !reflect.DeepEqual(got, []int64{2026}) {
		t.Fatalf("empty db years = %v", got)
	}
	seedReports(t, s)
	if _, err := s.DB.ExecContext(t.Context(), `INSERT INTO entries (type, date, amount_cents, category_id, payer_id) VALUES ('expense','2023-02-01',100,6,3),('expense','2027-01-01',100,6,3)`); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		query string
		want  []int64
	}{
		{"", []int64{2027, 2026, 2025, 2023}},
		{"?payer_id=1", []int64{2026, 2025}},
		{"?payer_id=2", []int64{2026, 2025}},
		{"?payer_id=3", []int64{2027, 2026, 2023}},
	}
	for _, tt := range tests {
		if got := decode[[]int64](t, do(t, h, "GET", "/api/reports/years"+tt.query, nil)); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("years%s = %v, want %v", tt.query, got, tt.want)
		}
	}
	// Current year is always present, even when the payer only has other years.
	if _, err := s.DB.ExecContext(t.Context(), `UPDATE entries SET payer_id = 2 WHERE payer_id = 1`); err != nil {
		t.Fatal(err)
	}
	if got := decode[[]int64](t, do(t, h, "GET", "/api/reports/years?payer_id=1", nil)); !reflect.DeepEqual(got, []int64{2026}) {
		t.Fatalf("years = %v", got)
	}
}

func TestInsights(t *testing.T) {
	s, h := newTestServer(t)
	seedReports(t, s)

	summary := func(ins []reports.Insight) []string {
		out := []string{}
		for _, i := range ins {
			out = append(out, i.Kind+"/"+i.Severity+"/"+i.Title)
		}
		return out
	}

	t.Run("current month", func(t *testing.T) {
		rec := do(t, h, "GET", "/api/insights", nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("status %d: %s", rec.Code, rec.Body)
		}
		got := decode[[]reports.Insight](t, rec)
		want := []string{
			"category_up/warning/House up 54%",
			"category_up/warning/Groceries up 50%",
			// (450 + 20) / 15 days × 31; the 20th's dinner is not "so far".
			"pace/warning/On pace for 971,33 €",
			"savings_rate/warning/Savings rate down to 55%",
			"top_expense/info/Big expense: Mortgage",
			"top_expense/info/Big expense: Groceries",
			"top_expense/info/Big expense: Electricity",
			"new_recurring/info/New recurring: Electricity",
		}
		if s := summary(got); !reflect.DeepEqual(s, want) {
			t.Fatalf("got\n%v\nwant\n%v", s, want)
		}
		if got[4].EntryID == nil || *got[4].EntryID != 14 || *got[4].AmountCents != 80000 || *got[4].CategoryID != 51 {
			t.Fatalf("top expense = %+v", got[4])
		}
		if got[2].Detail != "470,00 € non-recurring spent in 15 days · 350,00 € last month" {
			t.Fatalf("pace detail = %q", got[2].Detail)
		}
	})

	t.Run("past month has no pace", func(t *testing.T) {
		got := decode[[]reports.Insight](t, do(t, h, "GET", "/api/insights?month=2026-02", nil))
		want := []string{
			"category_up/warning/House up 200%",
			"category_up/warning/Groceries up 50%",
			"top_expense/info/Big expense: Mortgage",
			"top_expense/info/Big expense: Groceries",
			"top_expense/info/Big expense: Eating out",
			"savings_rate/info/Savings rate down to 62%",
		}
		if s := summary(got); !reflect.DeepEqual(s, want) {
			t.Fatalf("got\n%v\nwant\n%v", s, want)
		}
	})

	t.Run("payer filter", func(t *testing.T) {
		got := decode[[]reports.Insight](t, do(t, h, "GET", "/api/insights?payer_id=3", nil))
		want := []string{
			// Joint pays House and the new Electricity template; no income so no savings rate; pace = 20 / 15 × 31.
			"category_up/warning/House up 54%",
			"pace/info/On pace for 41,33 €",
			"top_expense/info/Big expense: Mortgage",
			"top_expense/info/Big expense: Electricity",
			"top_expense/info/Big expense: House",
			"new_recurring/info/New recurring: Electricity",
		}
		if s := summary(got); !reflect.DeepEqual(s, want) {
			t.Fatalf("got\n%v\nwant\n%v", s, want)
		}
		none := decode[[]reports.Insight](t, do(t, h, "GET", "/api/insights?payer_id=2", nil))
		if len(none) != 0 {
			t.Fatalf("payer B = %v", summary(none))
		}
	})

	t.Run("empty month is []", func(t *testing.T) {
		rec := do(t, h, "GET", "/api/insights?month=2020-01", nil)
		if rec.Code != http.StatusOK || rec.Body.String() != "[]\n" {
			t.Fatalf("got %d %q", rec.Code, rec.Body)
		}
	})
}
