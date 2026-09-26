package reports

import (
	"cmp"
	"fmt"
	"slices"
	"strconv"
)

type Totals struct {
	IncomeCents     int64    `json:"income_cents"`
	ExpenseCents    int64    `json:"expense_cents"`
	InvestmentCents int64    `json:"investment_cents"`
	LeftoverCents   int64    `json:"leftover_cents"`
	SavingsRate     *float64 `json:"savings_rate"`
}

type YearCategory struct {
	CategoryID    int64  `json:"category_id"`
	Name          string `json:"name"`
	Icon          string `json:"icon"`
	Color         string `json:"color"`
	Type          string `json:"type"`
	AmountCents   int64  `json:"amount_cents"`
	PrevYearCents int64  `json:"prev_year_cents"`
}

type YearMonth struct {
	Month           string `json:"month"`
	IncomeCents     int64  `json:"income_cents"`
	ExpenseCents    int64  `json:"expense_cents"`
	InvestmentCents int64  `json:"investment_cents"`
}

type YearReport struct {
	Year       int            `json:"year"`
	Totals     Totals         `json:"totals"`
	PrevTotals Totals         `json:"prev_totals"`
	Categories []YearCategory `json:"categories"`
	Monthly    []YearMonth    `json:"monthly"`
}

// YearWindow returns the date range (inclusive) BuildYear needs: the previous
// year through the end of year.
func YearWindow(year int) (from, to string) {
	return fmt.Sprintf("%04d-01-01", year-1), fmt.Sprintf("%04d-12-31", year)
}

// BuildYear assembles the year report. rows must cover YearWindow(year).
func BuildYear(year int, rows []Row, cats Categories) YearReport {
	cur, prev := strconv.Itoa(year), strconv.Itoa(year-1)
	rep := YearReport{Year: year, Categories: []YearCategory{}, Monthly: make([]YearMonth, 12)}
	for i := range rep.Monthly {
		rep.Monthly[i].Month = fmt.Sprintf("%04d-%02d", year, i+1)
	}

	var totals, prevTotals Totals
	byCat := map[int64]*YearCategory{}
	for _, r := range rows {
		y := r.Month[:4]
		var t *Totals
		switch y {
		case cur:
			t = &totals
		case prev:
			t = &prevTotals
		default:
			continue
		}
		addTotals(t, r)

		c, ok := byCat[r.TopCategoryID]
		if !ok {
			cat := cats.get(r.TopCategoryID)
			c = &YearCategory{CategoryID: cat.ID, Name: cat.Name, Icon: cat.Icon, Color: cat.Color, Type: r.Type}
			byCat[r.TopCategoryID] = c
		}
		if y == cur {
			c.AmountCents += r.AmountCents
			if m, err := strconv.Atoi(r.Month[5:7]); err == nil && m >= 1 && m <= 12 {
				addMonthly(&rep.Monthly[m-1], r)
			}
		} else {
			c.PrevYearCents += r.AmountCents
		}
	}
	rep.Totals = finishTotals(totals)
	rep.PrevTotals = finishTotals(prevTotals)

	for _, c := range byCat {
		if c.AmountCents != 0 || c.PrevYearCents != 0 {
			rep.Categories = append(rep.Categories, *c)
		}
	}
	slices.SortFunc(rep.Categories, func(a, b YearCategory) int {
		return cmp.Or(
			cmp.Compare(typeRank(a.Type), typeRank(b.Type)),
			cmp.Compare(b.AmountCents, a.AmountCents),
			cmp.Compare(b.PrevYearCents, a.PrevYearCents),
			cmp.Compare(a.Name, b.Name),
			cmp.Compare(a.CategoryID, b.CategoryID),
		)
	})
	return rep
}

func addTotals(t *Totals, r Row) {
	switch r.Type {
	case TypeIncome:
		t.IncomeCents += r.AmountCents
	case TypeExpense:
		t.ExpenseCents += r.AmountCents
	case TypeInvestment:
		t.InvestmentCents += r.AmountCents
	}
}

func addMonthly(m *YearMonth, r Row) {
	switch r.Type {
	case TypeIncome:
		m.IncomeCents += r.AmountCents
	case TypeExpense:
		m.ExpenseCents += r.AmountCents
	case TypeInvestment:
		m.InvestmentCents += r.AmountCents
	}
}

func finishTotals(t Totals) Totals {
	t.LeftoverCents = t.IncomeCents - t.ExpenseCents - t.InvestmentCents
	t.SavingsRate = savingsRate(t.IncomeCents, t.ExpenseCents)
	return t
}
