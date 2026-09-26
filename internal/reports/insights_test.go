package reports

import (
	"reflect"
	"testing"
)

// kinds summarises insights as "kind/severity/category" for compact assertions.
func kinds(ins []Insight) []string {
	out := []string{}
	for _, i := range ins {
		s := i.Kind + "/" + i.Severity
		if i.CategoryID != nil && (i.Kind == KindCategoryUp || i.Kind == KindCategoryDown) {
			s += "/" + testCats[*i.CategoryID].Name
		}
		out = append(out, s)
	}
	return out
}

// hist gives category 6 an average of avg over Dec–Feb and cur in March.
func hist(cat, avg, cur int64) []Row {
	rows := []Row{
		row("2025-12", TypeExpense, cat, avg),
		row("2026-01", TypeExpense, cat, avg),
		row("2026-02", TypeExpense, cat, avg),
	}
	if cur > 0 {
		rows = append(rows, row("2026-03", TypeExpense, cat, cur))
	}
	return rows
}

func TestCategoryInsights(t *testing.T) {
	tests := []struct {
		name  string
		rows  []Row
		today string
		want  []string
	}{
		{name: "up past month", rows: hist(6, 30000, 45000), today: "2026-04-10", want: []string{"category_up/warning/Groceries"}},
		{name: "down past month", rows: hist(6, 30000, 20000), today: "2026-04-10", want: []string{"category_down/good/Groceries"}},
		{name: "drop to zero past month", rows: hist(6, 30000, 0), today: "2026-04-10", want: []string{"category_down/good/Groceries"}},
		{name: "down ignored in current month", rows: hist(6, 30000, 20000), today: "2026-03-20", want: []string{}},
		{name: "down ignored in future month", rows: hist(6, 30000, 20000), today: "2026-02-20", want: []string{}},
		{name: "up in current month", rows: hist(6, 30000, 45000), today: "2026-03-02", want: []string{"category_up/warning/Groceries"}},
		// 20% exactly is not "more than 20%".
		{name: "exactly 20 pct", rows: hist(6, 30000, 36000), today: "2026-04-10", want: []string{}},
		{name: "just over 20 pct", rows: hist(6, 30000, 36001), today: "2026-04-10", want: []string{"category_up/warning/Groceries"}},
		// Big relative change but only 20,00 € absolute.
		{name: "exactly 2000 cents", rows: hist(6, 5000, 7000), today: "2026-04-10", want: []string{}},
		{name: "2001 cents", rows: hist(6, 5000, 7001), today: "2026-04-10", want: []string{"category_up/warning/Groceries"}},
		{name: "no history", rows: []Row{row("2026-03", TypeExpense, 6, 90000)}, today: "2026-04-10", want: []string{}},
		{name: "income ignored", rows: []Row{
			row("2026-01", TypeIncome, 100, 100000), row("2026-03", TypeIncome, 100, 900000),
		}, today: "2026-04-10", want: []string{}},
		{name: "subcategories roll up", rows: []Row{
			row("2026-02", TypeExpense, 51, 30000), row("2026-03", TypeExpense, 51, 30000), row("2026-03", TypeExpense, 50, 30000),
		}, today: "2026-04-10", want: []string{"category_up/warning/House"}},
		{name: "ordered by absolute change", rows: append(hist(6, 30000, 40000), hist(12, 10000, 40000)...), today: "2026-04-10", want: []string{
			"category_up/warning/Eating out", "category_up/warning/Groceries",
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildInsights(InsightInput{Month: "2026-03", Today: tt.today, Rows: tt.rows, Cats: testCats})
			if err != nil {
				t.Fatal(err)
			}
			if k := kinds(got); !reflect.DeepEqual(k, tt.want) {
				t.Fatalf("got %v, want %v", k, tt.want)
			}
		})
	}
}

func TestCategoryInsightText(t *testing.T) {
	got, err := BuildInsights(InsightInput{Month: "2026-03", Today: "2026-04-01", Rows: hist(6, 31200, 41230), Cats: testCats})
	if err != nil {
		t.Fatal(err)
	}
	want := Insight{
		Kind: KindCategoryUp, Severity: SeverityWarning,
		Title: "Groceries up 32%", Detail: "412,30 € vs 312,00 € 3-month average",
		CategoryID: new(int64(6)), AmountCents: new(int64(41230)),
	}
	if len(got) != 1 || !reflect.DeepEqual(got[0], want) {
		t.Fatalf("got %+v", got)
	}
}

func TestPaceInsight(t *testing.T) {
	prev := []Row{
		row("2026-02", TypeExpense, 6, 100000),
		recurring(row("2026-02", TypeExpense, 51, 500000)), // recurring: not in the baseline
	}
	tests := []struct {
		name      string
		month     string
		today     string
		rows      []Row
		soFar     int64
		want      *Insight
		wantTitle string
	}{
		{name: "before day 5", month: "2026-03", today: "2026-03-04", rows: prev, soFar: 50000},
		{name: "past month", month: "2026-02", today: "2026-03-15", rows: prev, soFar: 50000},
		{name: "nothing spent", month: "2026-03", today: "2026-03-15", rows: prev, soFar: 0},
		{
			// 50000 / 15 × 31 = 103333,3 -> within 10% of 100000.
			name: "on track", month: "2026-03", today: "2026-03-15", rows: prev, soFar: 50000,
			want: &Insight{
				Kind: KindPace, Severity: SeverityInfo, Title: "On pace for 1033,33 €",
				Detail: "500,00 € non-recurring spent in 15 days · 1000,00 € last month", AmountCents: new(int64(103333)),
			},
		},
		{
			// 5 days: 22000 / 5 × 31 = 136400 > 110000.
			name: "over pace on day 5", month: "2026-03", today: "2026-03-05", rows: prev, soFar: 22000,
			want: &Insight{
				Kind: KindPace, Severity: SeverityWarning, Title: "On pace for 1364,00 €",
				Detail: "220,00 € non-recurring spent in 5 days · 1000,00 € last month", AmountCents: new(int64(136400)),
			},
		},
		{
			// Exactly +10% is not more than 10%: 110000 × 28 / 28.
			name: "exactly 10 pct over", month: "2026-02", today: "2026-02-28",
			rows: []Row{row("2026-01", TypeExpense, 6, 100000)}, soFar: 110000,
			want: &Insight{
				Kind: KindPace, Severity: SeverityInfo, Title: "On pace for 1100,00 €",
				Detail: "1100,00 € non-recurring spent in 28 days · 1000,00 € last month", AmountCents: new(int64(110000)),
			},
		},
		{
			name: "no previous month", month: "2026-03", today: "2026-03-10", soFar: 10000,
			want: &Insight{
				Kind: KindPace, Severity: SeverityInfo, Title: "On pace for 310,00 €",
				Detail: "100,00 € non-recurring spent in 10 days", AmountCents: new(int64(31000)),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildInsights(InsightInput{Month: tt.month, Today: tt.today, Rows: tt.rows, Cats: testCats, NonRecurringSoFar: tt.soFar})
			if err != nil {
				t.Fatal(err)
			}
			var pace []Insight
			for _, i := range got {
				if i.Kind == KindPace {
					pace = append(pace, i)
				}
			}
			switch {
			case tt.want == nil && len(pace) != 0:
				t.Fatalf("unexpected pace %+v", pace)
			case tt.want != nil && (len(pace) != 1 || !reflect.DeepEqual(pace[0], *tt.want)):
				t.Fatalf("pace = %+v, want %+v", pace, *tt.want)
			}
		})
	}
}

func TestTopExpenseInsights(t *testing.T) {
	top := []TopEntry{
		{ID: 7, Date: "2026-03-04", AmountCents: 80000, CategoryID: 51, CategoryName: "Mortgage", Note: " March "},
		{ID: 3, Date: "2026-03-12", AmountCents: 45000, CategoryID: 6, CategoryName: "Groceries"},
		{ID: 9, Date: "2026-03-01", AmountCents: 6000, CategoryID: 2, CategoryName: "Electricity"},
		{ID: 1, Date: "2026-03-01", AmountCents: 10, CategoryID: 2, CategoryName: "Electricity"}, // beyond the limit
	}
	got, err := BuildInsights(InsightInput{Month: "2026-03", Today: "2026-04-01", Cats: testCats, Top: top})
	if err != nil {
		t.Fatal(err)
	}
	want := []Insight{
		{Kind: KindTopExpense, Severity: SeverityInfo, Title: "Big expense: Mortgage", Detail: "800,00 € · 4 Mar · March", CategoryID: new(int64(51)), EntryID: new(int64(7)), AmountCents: new(int64(80000))},
		{Kind: KindTopExpense, Severity: SeverityInfo, Title: "Big expense: Groceries", Detail: "450,00 € · 12 Mar", CategoryID: new(int64(6)), EntryID: new(int64(3)), AmountCents: new(int64(45000))},
		{Kind: KindTopExpense, Severity: SeverityInfo, Title: "Big expense: Electricity", Detail: "60,00 € · 1 Mar", CategoryID: new(int64(2)), EntryID: new(int64(9)), AmountCents: new(int64(6000))},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v", got)
	}
}

func TestSavingsRateInsight(t *testing.T) {
	months := func(prevInc, prevExp, inc, exp int64) []Row {
		var rows []Row
		add := func(m, typ string, cat, c int64) {
			if c > 0 {
				rows = append(rows, row(m, typ, cat, c))
			}
		}
		add("2026-02", TypeIncome, 100, prevInc)
		add("2026-02", TypeExpense, 19, prevExp)
		add("2026-03", TypeIncome, 100, inc)
		add("2026-03", TypeExpense, 19, exp)
		return rows
	}
	tests := []struct {
		name  string
		rows  []Row
		want  string // severity, "" = no insight
		title string
		det   string
	}{
		{name: "up", rows: months(100000, 70000, 100000, 60000), want: SeverityGood, title: "Savings rate up to 40%", det: "30% last month (+10 pts)"},
		{name: "steady", rows: months(100000, 70000, 200000, 140000), want: SeverityInfo, title: "Savings rate steady at 30%", det: "30% last month (+0 pts)"},
		{name: "small drop", rows: months(100000, 70000, 100000, 74000), want: SeverityInfo, title: "Savings rate down to 26%", det: "30% last month (-4 pts)"},
		{name: "exactly 5 pts drop", rows: months(100000, 70000, 100000, 75000), want: SeverityInfo, title: "Savings rate down to 25%", det: "30% last month (-5 pts)"},
		{name: "big drop", rows: months(100000, 70000, 100000, 75100), want: SeverityWarning, title: "Savings rate down to 25%", det: "30% last month (-5 pts)"},
		{name: "negative rate", rows: months(100000, 70000, 100000, 150000), want: SeverityWarning, title: "Savings rate down to -50%", det: "30% last month (-80 pts)"},
		{name: "no income this month", rows: months(100000, 70000, 0, 50000)},
		{name: "no income last month", rows: months(0, 70000, 100000, 50000)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildInsights(InsightInput{Month: "2026-03", Today: "2026-04-01", Rows: tt.rows, Cats: testCats})
			if err != nil {
				t.Fatal(err)
			}
			var sr []Insight
			for _, i := range got {
				if i.Kind == KindSavingsRate {
					sr = append(sr, i)
				}
			}
			if tt.want == "" {
				if len(sr) != 0 {
					t.Fatalf("unexpected %+v", sr)
				}
				return
			}
			if len(sr) != 1 || sr[0].Severity != tt.want || sr[0].Title != tt.title || sr[0].Detail != tt.det ||
				sr[0].AmountCents != nil || sr[0].CategoryID != nil {
				t.Fatalf("got %+v", sr)
			}
		})
	}
}

func TestNewRecurringInsight(t *testing.T) {
	tpls := []NewTemplate{
		{ID: 1, Type: TypeExpense, CategoryID: 2, CategoryName: "Electricity", AmountCents: 6000, Variable: true},
		{ID: 2, Type: TypeInvestment, CategoryID: 200, CategoryName: "ETFs/Stocks", AmountCents: 50000, Note: "IWDA"},
	}
	got, err := BuildInsights(InsightInput{Month: "2026-03", Today: "2026-04-01", Cats: testCats, NewTemplates: tpls})
	if err != nil {
		t.Fatal(err)
	}
	want := []Insight{
		{Kind: KindNewRecurring, Severity: SeverityInfo, Title: "New recurring: Electricity", Detail: "60,00 € / month (variable)", CategoryID: new(int64(2)), AmountCents: new(int64(6000))},
		{Kind: KindNewRecurring, Severity: SeverityInfo, Title: "New recurring: ETFs/Stocks", Detail: "500,00 € / month · investment · IWDA", CategoryID: new(int64(200)), AmountCents: new(int64(50000))},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v", got)
	}
}

func TestInsightOrdering(t *testing.T) {
	rows := append(hist(6, 30000, 20000), hist(12, 10000, 40000)...)
	rows = append(rows,
		row("2026-02", TypeIncome, 100, 100000),
		row("2026-03", TypeIncome, 100, 100000),
	)
	got, err := BuildInsights(InsightInput{
		Month: "2026-03", Today: "2026-04-02", Rows: rows, Cats: testCats,
		Top:          []TopEntry{{ID: 1, Date: "2026-03-02", AmountCents: 40000, CategoryID: 12, CategoryName: "Eating out"}},
		NewTemplates: []NewTemplate{{ID: 1, Type: TypeExpense, CategoryID: 2, CategoryName: "Electricity", AmountCents: 6000}},
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"category_up/warning/Eating out",
		"savings_rate/warning", // 60% -> 40%
		"category_down/good/Groceries",
		"top_expense/info",
		"new_recurring/info",
	}
	if k := kinds(got); !reflect.DeepEqual(k, want) {
		t.Fatalf("got %v, want %v", k, want)
	}
}

func TestBuildInsightsEmptyAndErrors(t *testing.T) {
	got, err := BuildInsights(InsightInput{Month: "2026-03", Today: "2026-03-15", Cats: testCats})
	if err != nil || got == nil || len(got) != 0 {
		t.Fatalf("got %v, %v; want empty non-nil", got, err)
	}
	if _, err := BuildInsights(InsightInput{Month: "bad", Today: "2026-03-15"}); err == nil {
		t.Fatal("want error")
	}
}
