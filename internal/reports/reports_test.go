package reports

import (
	"reflect"
	"testing"
)

// testCats mirrors part of the seed data.
var testCats = Categories{
	2:   {ID: 2, Type: TypeExpense, Name: "Electricity", Icon: "zap", Color: "#eab308"},
	5:   {ID: 5, Type: TypeExpense, Name: "House", Icon: "house", Color: "#8b5cf6"},
	6:   {ID: 6, Type: TypeExpense, Name: "Groceries", Icon: "shopping-cart", Color: "#10b981"},
	9:   {ID: 9, Type: TypeExpense, Name: "Transport", Icon: "bus", Color: "#14b8a6"},
	12:  {ID: 12, Type: TypeExpense, Name: "Eating out", Icon: "utensils", Color: "#f59e0b"},
	50:  {ID: 50, ParentID: new(int64(5)), Type: TypeExpense, Name: "Rent", Icon: "key-round", Color: "#8b5cf6"},
	51:  {ID: 51, ParentID: new(int64(5)), Type: TypeExpense, Name: "Mortgage", Icon: "landmark", Color: "#8b5cf6"},
	55:  {ID: 55, ParentID: new(int64(9)), Type: TypeExpense, Name: "Public", Icon: "train-front", Color: "#14b8a6"},
	100: {ID: 100, Type: TypeIncome, Name: "Salary", Icon: "briefcase", Color: "#10b981"},
	101: {ID: 101, Type: TypeIncome, Name: "Bonus", Icon: "award", Color: "#22c55e"},
	200: {ID: 200, Type: TypeInvestment, Name: "ETFs/Stocks", Icon: "trending-up", Color: "#6366f1"},
}

// row builds a confirmed, non-recurring single-entry row; top is derived from testCats.
func row(month, typ string, cat, cents int64) Row {
	top := cat
	if p := testCats[cat].ParentID; p != nil {
		top = *p
	}
	return Row{Month: month, Type: typ, CategoryID: cat, TopCategoryID: top, AmountCents: cents, Count: 1}
}

func pending(r Row) Row   { r.Pending = true; return r }
func recurring(r Row) Row { r.Recurring = true; return r }

func TestFormatCents(t *testing.T) {
	tests := []struct {
		cents int64
		want  string
	}{
		{0, "0,00 €"},
		{5, "0,05 €"},
		{41230, "412,30 €"},
		{123456, "1234,56 €"},
		{999999, "9999,99 €"},
		{1000000, "10 000,00 €"},
		{1234567, "12 345,67 €"},
		{123456789, "1 234 567,89 €"},
		{-1250, "-12,50 €"},
		{-1234567, "-12 345,67 €"},
	}
	for _, tt := range tests {
		if got := FormatCents(tt.cents); got != tt.want {
			t.Errorf("FormatCents(%d) = %q, want %q", tt.cents, got, tt.want)
		}
	}
}

func TestDivRound(t *testing.T) {
	tests := []struct{ num, den, want int64 }{
		{0, 3, 0},
		{1, 3, 0},
		{2, 3, 1},
		{3, 2, 2}, // half rounds up
		{160000, 3, 53333},
		{160001, 3, 53334},
		{-3, 2, -2},
		{-1, 3, 0},
		{5, 0, 0},
	}
	for _, tt := range tests {
		if got := divRound(tt.num, tt.den); got != tt.want {
			t.Errorf("divRound(%d, %d) = %d, want %d", tt.num, tt.den, got, tt.want)
		}
	}
}

func TestKPIsFor(t *testing.T) {
	rows := []Row{
		row("2026-03", TypeIncome, 100, 300000),
		pending(row("2026-03", TypeIncome, 101, 50000)),
		row("2026-03", TypeExpense, 6, 100000),
		pending(row("2026-03", TypeExpense, 2, 7000)),
		row("2026-03", TypeInvestment, 200, 50000),
		row("2026-02", TypeExpense, 6, 99999), // other month ignored
	}
	tests := []struct {
		name  string
		rows  []Row
		month string
		want  KPIs
	}{
		{
			name: "full month", rows: rows, month: "2026-03",
			want: KPIs{
				IncomeCents: 350000, ExpenseCents: 107000, InvestmentCents: 50000,
				LeftoverCents: 193000, SavingsRate: new(243000.0 / 350000.0), PendingCents: 57000,
			},
		},
		{
			name: "no income", rows: rows, month: "2026-02",
			want: KPIs{ExpenseCents: 99999, LeftoverCents: -99999},
		},
		{name: "empty", month: "2026-01", want: KPIs{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := KPIsFor(tt.rows, tt.month); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %+v (rate %v), want %+v", got, deref(got.SavingsRate), tt.want)
			}
		})
	}
}

func deref(p *float64) any {
	if p == nil {
		return nil
	}
	return *p
}

func TestMonthWindow(t *testing.T) {
	from, to, err := MonthWindow("2026-02")
	if err != nil || from != "2025-11" || to != "2026-02" {
		t.Fatalf("MonthWindow = %q %q %v", from, to, err)
	}
	if _, _, err := MonthWindow("nope"); err == nil {
		t.Fatal("want error")
	}
}

func TestBuildMonth(t *testing.T) {
	rows := []Row{
		// Three months before (Dec, Jan, Feb) for avg3; Nov must not count.
		row("2025-11", TypeExpense, 6, 900000),
		row("2025-12", TypeExpense, 6, 30000),
		row("2026-01", TypeExpense, 6, 30000),
		recurring(row("2026-01", TypeExpense, 51, 80000)),
		row("2026-02", TypeExpense, 6, 30001),
		recurring(row("2026-02", TypeExpense, 51, 80000)),
		row("2026-02", TypeIncome, 100, 300000),
		// March.
		row("2026-03", TypeExpense, 6, 45000),
		row("2026-03", TypeExpense, 5, 2000), // booked on the parent -> "(general)"
		recurring(row("2026-03", TypeExpense, 51, 80000)),
		row("2026-03", TypeExpense, 50, 2000),
		pending(recurring(row("2026-03", TypeExpense, 2, 6000))),
		row("2026-03", TypeExpense, 12, 6000), // ties with Electricity: name decides
		row("2026-03", TypeIncome, 100, 300000),
		pending(row("2026-03", TypeIncome, 101, 1000)),
		row("2026-03", TypeInvestment, 200, 50000),
	}
	rep, err := BuildMonth("2026-03", rows, testCats)
	if err != nil {
		t.Fatal(err)
	}

	if rep.Month != "2026-03" {
		t.Fatalf("month = %q", rep.Month)
	}
	if rep.KPIs.ExpenseCents != 141000 || rep.KPIs.IncomeCents != 301000 || rep.KPIs.PendingCents != 7000 ||
		rep.KPIs.LeftoverCents != 301000-141000-50000 {
		t.Fatalf("kpis = %+v", rep.KPIs)
	}
	if rep.PrevKPIs.ExpenseCents != 110001 || rep.PrevKPIs.IncomeCents != 300000 {
		t.Fatalf("prev kpis = %+v", rep.PrevKPIs)
	}

	wantExp := []CategoryAmount{
		{
			CategoryID: 5, Name: "House", Icon: "house", Color: "#8b5cf6", Type: TypeExpense,
			AmountCents: 84000, PrevMonthCents: 80000, Avg3Cents: 53333,
			Subcategories: []Subcategory{
				{CategoryID: 51, Name: "Mortgage", Color: "#8b5cf6", AmountCents: 80000},
				{CategoryID: 5, Name: GeneralSubcategory, Color: "#8b5cf6", AmountCents: 2000},
				{CategoryID: 50, Name: "Rent", Color: "#8b5cf6", AmountCents: 2000},
			},
		},
		{
			CategoryID: 6, Name: "Groceries", Icon: "shopping-cart", Color: "#10b981", Type: TypeExpense,
			AmountCents: 45000, PrevMonthCents: 30001, Avg3Cents: 30000,
			Subcategories: []Subcategory{{CategoryID: 6, Name: GeneralSubcategory, Color: "#10b981", AmountCents: 45000}},
		},
		{
			CategoryID: 12, Name: "Eating out", Icon: "utensils", Color: "#f59e0b", Type: TypeExpense,
			AmountCents:   6000,
			Subcategories: []Subcategory{{CategoryID: 12, Name: GeneralSubcategory, Color: "#f59e0b", AmountCents: 6000}},
		},
		{
			CategoryID: 2, Name: "Electricity", Icon: "zap", Color: "#eab308", Type: TypeExpense,
			AmountCents: 6000, PendingCents: 6000,
			Subcategories: []Subcategory{{CategoryID: 2, Name: GeneralSubcategory, Color: "#eab308", AmountCents: 6000}},
		},
	}
	if !reflect.DeepEqual(rep.Expenses, wantExp) {
		t.Fatalf("expenses =\n%+v\nwant\n%+v", rep.Expenses, wantExp)
	}
	if len(rep.Income) != 2 || rep.Income[0].CategoryID != 100 || rep.Income[0].PrevMonthCents != 300000 ||
		rep.Income[0].Avg3Cents != 100000 || rep.Income[1].PendingCents != 1000 {
		t.Fatalf("income = %+v", rep.Income)
	}
	if len(rep.Investments) != 1 || rep.Investments[0].AmountCents != 50000 {
		t.Fatalf("investments = %+v", rep.Investments)
	}
	wantRec := Recurring{CommittedCents: 86000, EntryCount: 2, PendingCount: 1, PendingCents: 6000}
	if rep.Recurring != wantRec {
		t.Fatalf("recurring = %+v", rep.Recurring)
	}
}

func TestBuildMonthEmpty(t *testing.T) {
	rep, err := BuildMonth("2026-03", nil, testCats)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Expenses == nil || rep.Income == nil || rep.Investments == nil || len(rep.Expenses) != 0 {
		t.Fatalf("lists must be empty, non-nil: %+v", rep)
	}
	if rep.KPIs.SavingsRate != nil || rep.PrevKPIs.SavingsRate != nil {
		t.Fatal("savings rate must be nil without income")
	}
	if _, err := BuildMonth("bad", nil, testCats); err == nil {
		t.Fatal("want error for bad month")
	}
}

func TestBuildMonthRecurringCountsEntries(t *testing.T) {
	r := recurring(row("2026-03", TypeExpense, 51, 160000))
	r.Count = 2
	inc := recurring(row("2026-03", TypeIncome, 100, 300000)) // not an expense: ignored
	rep, err := BuildMonth("2026-03", []Row{r, inc}, testCats)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Recurring != (Recurring{CommittedCents: 160000, EntryCount: 2}) {
		t.Fatalf("recurring = %+v", rep.Recurring)
	}
}

func TestBuildTrends(t *testing.T) {
	months := []string{"2025-12", "2026-01", "2026-02"}
	rows := []Row{
		row("2025-11", TypeExpense, 6, 99999), // outside the window
		row("2025-12", TypeExpense, 6, 30000),
		row("2025-12", TypeIncome, 100, 300000),
		row("2026-01", TypeExpense, 51, 80000),
		row("2026-01", TypeExpense, 5, 1000),
		row("2026-01", TypeInvestment, 200, 50000),
		row("2026-02", TypeExpense, 6, 30000),
		row("2026-02", TypeExpense, 55, 500),
		row("2026-02", TypeIncome, 100, 300000),
	}

	t.Run("all categories", func(t *testing.T) {
		rep := BuildTrends(months, rows, testCats, nil)
		want := TrendsReport{
			Months: months,
			Series: []TrendPoint{
				{Month: "2025-12", IncomeCents: 300000, ExpenseCents: 30000, SavingsCents: 270000, ByCategory: map[string]int64{"6": 30000}},
				{Month: "2026-01", ExpenseCents: 81000, InvestmentCents: 50000, SavingsCents: -81000, ByCategory: map[string]int64{"5": 81000}},
				{Month: "2026-02", IncomeCents: 300000, ExpenseCents: 30500, SavingsCents: 269500, ByCategory: map[string]int64{"6": 30000, "9": 500}},
			},
			Categories: []TrendCategory{
				{ID: 5, Name: "House", Color: "#8b5cf6", Icon: "house"},
				{ID: 6, Name: "Groceries", Color: "#10b981", Icon: "shopping-cart"},
				{ID: 9, Name: "Transport", Color: "#14b8a6", Icon: "bus"},
			},
		}
		if !reflect.DeepEqual(rep, want) {
			t.Fatalf("got\n%+v\nwant\n%+v", rep, want)
		}
	})

	t.Run("single category", func(t *testing.T) {
		rep := BuildTrends(months, rows, testCats, new(int64(6)))
		if rep.Series[1].ExpenseCents != 0 || len(rep.Series[1].ByCategory) != 0 || rep.Series[2].ExpenseCents != 30000 {
			t.Fatalf("series = %+v", rep.Series)
		}
		// Savings keeps using every expense.
		if rep.Series[2].SavingsCents != 269500 || rep.Series[1].SavingsCents != -81000 {
			t.Fatalf("savings = %+v", rep.Series)
		}
		if len(rep.Categories) != 1 || rep.Categories[0].ID != 6 {
			t.Fatalf("categories = %+v", rep.Categories)
		}
	})

	t.Run("empty", func(t *testing.T) {
		rep := BuildTrends(months, nil, testCats, nil)
		if len(rep.Series) != 3 || rep.Categories == nil || len(rep.Categories) != 0 || rep.Series[0].ByCategory == nil {
			t.Fatalf("got %+v", rep)
		}
	})
}

func TestBuildYear(t *testing.T) {
	rows := []Row{
		row("2024-12", TypeExpense, 6, 99999), // outside
		row("2025-03", TypeExpense, 12, 100000),
		row("2025-12", TypeIncome, 100, 200000),
		row("2025-12", TypeExpense, 6, 20000),
		row("2026-01", TypeExpense, 6, 30000),
		row("2026-01", TypeExpense, 51, 80000),
		row("2026-01", TypeIncome, 100, 300000),
		row("2026-02", TypeIncome, 101, 30000),
		pending(row("2026-03", TypeExpense, 2, 6000)),
		row("2026-12", TypeInvestment, 200, 50000),
	}
	rep := BuildYear(2026, rows, testCats)
	wantTotals := Totals{IncomeCents: 330000, ExpenseCents: 116000, InvestmentCents: 50000, LeftoverCents: 164000, SavingsRate: new(214000.0 / 330000.0)}
	if !reflect.DeepEqual(rep.Totals, wantTotals) {
		t.Fatalf("totals = %+v", rep.Totals)
	}
	wantPrev := Totals{IncomeCents: 200000, ExpenseCents: 120000, LeftoverCents: 80000, SavingsRate: new(0.4)}
	if !reflect.DeepEqual(rep.PrevTotals, wantPrev) {
		t.Fatalf("prev totals = %+v", rep.PrevTotals)
	}
	var got []int64
	for _, c := range rep.Categories {
		got = append(got, c.CategoryID, c.AmountCents, c.PrevYearCents)
	}
	want := []int64{
		5, 80000, 0,
		6, 30000, 20000,
		2, 6000, 0,
		12, 0, 100000,
		100, 300000, 200000,
		101, 30000, 0,
		200, 50000, 0,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("categories = %v, want %v", got, want)
	}
	if len(rep.Monthly) != 12 || rep.Monthly[0] != (YearMonth{Month: "2026-01", IncomeCents: 300000, ExpenseCents: 110000}) ||
		rep.Monthly[2].ExpenseCents != 6000 || rep.Monthly[11] != (YearMonth{Month: "2026-12", InvestmentCents: 50000}) ||
		rep.Monthly[5] != (YearMonth{Month: "2026-06"}) {
		t.Fatalf("monthly = %+v", rep.Monthly)
	}

	empty := BuildYear(2030, nil, testCats)
	if empty.Categories == nil || len(empty.Monthly) != 12 || empty.Totals.SavingsRate != nil {
		t.Fatalf("empty = %+v", empty)
	}
}

func TestYearWindow(t *testing.T) {
	if from, to := YearWindow(2026); from != "2025-01-01" || to != "2026-12-31" {
		t.Fatalf("YearWindow = %q %q", from, to)
	}
}
