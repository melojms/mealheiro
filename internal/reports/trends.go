package reports

import (
	"slices"
	"strconv"
)

type TrendPoint struct {
	Month           string           `json:"month"`
	IncomeCents     int64            `json:"income_cents"`
	ExpenseCents    int64            `json:"expense_cents"`
	InvestmentCents int64            `json:"investment_cents"`
	SavingsCents    int64            `json:"savings_cents"`
	ByCategory      map[string]int64 `json:"by_category"`
}

type TrendCategory struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
	Icon  string `json:"icon"`
}

type TrendsReport struct {
	Months     []string        `json:"months"`
	Series     []TrendPoint    `json:"series"`
	Categories []TrendCategory `json:"categories"`
}

// BuildTrends assembles the trends report for the given ascending months.
// A non-nil categoryID (top-level expense) restricts expense_cents and
// by_category to that category; savings_cents always uses all expenses.
func BuildTrends(months []string, rows []Row, cats Categories, categoryID *int64) TrendsReport {
	idx := make(map[string]int, len(months))
	rep := TrendsReport{
		Months:     slices.Clone(months),
		Series:     make([]TrendPoint, len(months)),
		Categories: []TrendCategory{},
	}
	for i, m := range months {
		idx[m] = i
		rep.Series[i] = TrendPoint{Month: m, ByCategory: map[string]int64{}}
	}

	allExpense := make([]int64, len(months))
	catTotals := map[int64]int64{}
	for _, r := range rows {
		i, ok := idx[r.Month]
		if !ok {
			continue
		}
		p := &rep.Series[i]
		switch r.Type {
		case TypeIncome:
			p.IncomeCents += r.AmountCents
		case TypeInvestment:
			p.InvestmentCents += r.AmountCents
		case TypeExpense:
			allExpense[i] += r.AmountCents
			if categoryID != nil && r.TopCategoryID != *categoryID {
				continue
			}
			p.ExpenseCents += r.AmountCents
			p.ByCategory[strconv.FormatInt(r.TopCategoryID, 10)] += r.AmountCents
			catTotals[r.TopCategoryID] += r.AmountCents
		}
	}
	for i := range rep.Series {
		rep.Series[i].SavingsCents = rep.Series[i].IncomeCents - allExpense[i]
	}

	for id, total := range catTotals {
		if total == 0 {
			continue
		}
		c := cats.get(id)
		rep.Categories = append(rep.Categories, TrendCategory{ID: c.ID, Name: c.Name, Color: c.Color, Icon: c.Icon})
	}
	slices.SortFunc(rep.Categories, func(a, b TrendCategory) int {
		return byAmountDesc(catTotals[a.ID], catTotals[b.ID], a.Name, b.Name, a.ID, b.ID)
	})
	return rep
}
