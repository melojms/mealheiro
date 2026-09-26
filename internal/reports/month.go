package reports

import (
	"slices"

	"github.com/melojms/mm-budget/internal/clock"
)

// GeneralSubcategory names the pseudo-subcategory for entries booked on a parent.
const GeneralSubcategory = "(general)"

// KPIs are the headline numbers of one month.
type KPIs struct {
	IncomeCents     int64    `json:"income_cents"`
	ExpenseCents    int64    `json:"expense_cents"`
	InvestmentCents int64    `json:"investment_cents"`
	LeftoverCents   int64    `json:"leftover_cents"`
	SavingsRate     *float64 `json:"savings_rate"`
	PendingCents    int64    `json:"pending_cents"`
}

type Subcategory struct {
	CategoryID  int64  `json:"category_id"`
	Name        string `json:"name"`
	Color       string `json:"color"`
	AmountCents int64  `json:"amount_cents"`
}

type CategoryAmount struct {
	CategoryID     int64         `json:"category_id"`
	Name           string        `json:"name"`
	Icon           string        `json:"icon"`
	Color          string        `json:"color"`
	Type           string        `json:"type"`
	AmountCents    int64         `json:"amount_cents"`
	PendingCents   int64         `json:"pending_cents"`
	PrevMonthCents int64         `json:"prev_month_cents"`
	Avg3Cents      int64         `json:"avg3_cents"`
	Subcategories  []Subcategory `json:"subcategories"`
}

type Recurring struct {
	CommittedCents int64 `json:"committed_cents"`
	EntryCount     int64 `json:"entry_count"`
	PendingCount   int64 `json:"pending_count"`
	PendingCents   int64 `json:"pending_cents"`
}

type MonthReport struct {
	Month       string           `json:"month"`
	KPIs        KPIs             `json:"kpis"`
	PrevKPIs    KPIs             `json:"prev_kpis"`
	Expenses    []CategoryAmount `json:"expenses"`
	Income      []CategoryAmount `json:"income"`
	Investments []CategoryAmount `json:"investments"`
	Recurring   Recurring        `json:"recurring"`
}

// MonthWindow returns the first and last month whose rows BuildMonth and
// BuildInsights need: the three months before month, through month.
func MonthWindow(month string) (from, to string, err error) {
	from, err = clock.AddMonths(month, -3)
	return from, month, err
}

// KPIsFor sums rows of one month into KPIs.
func KPIsFor(rows []Row, month string) KPIs {
	var k KPIs
	for _, r := range rows {
		if r.Month != month {
			continue
		}
		switch r.Type {
		case TypeIncome:
			k.IncomeCents += r.AmountCents
		case TypeExpense:
			k.ExpenseCents += r.AmountCents
		case TypeInvestment:
			k.InvestmentCents += r.AmountCents
		}
		if r.Pending {
			k.PendingCents += r.AmountCents
		}
	}
	k.LeftoverCents = k.IncomeCents - k.ExpenseCents - k.InvestmentCents
	k.SavingsRate = savingsRate(k.IncomeCents, k.ExpenseCents)
	return k
}

// topTotals sums rows of the given type per month and top-level category.
type topKey struct {
	month string
	top   int64
}

func topTotals(rows []Row, typ string) map[topKey]int64 {
	out := map[topKey]int64{}
	for _, r := range rows {
		if r.Type == typ {
			out[topKey{r.Month, r.TopCategoryID}] += r.AmountCents
		}
	}
	return out
}

// avg3 is the rounded average of the three months before month for one category.
func avg3(totals map[topKey]int64, month string, top int64) int64 {
	var sum int64
	for i := 1; i <= 3; i++ {
		m, _ := clock.AddMonths(month, -i)
		sum += totals[topKey{m, top}]
	}
	return divRound(sum, 3)
}

// BuildMonth assembles the month report. rows must cover MonthWindow(month).
func BuildMonth(month string, rows []Row, cats Categories) (MonthReport, error) {
	prev, err := clock.AddMonths(month, -1)
	if err != nil {
		return MonthReport{}, err
	}
	rep := MonthReport{
		Month:       month,
		KPIs:        KPIsFor(rows, month),
		PrevKPIs:    KPIsFor(rows, prev),
		Expenses:    categoryAmounts(rows, cats, month, prev, TypeExpense),
		Income:      categoryAmounts(rows, cats, month, prev, TypeIncome),
		Investments: categoryAmounts(rows, cats, month, prev, TypeInvestment),
	}
	for _, r := range rows {
		if r.Month != month || r.Type != TypeExpense || !r.Recurring {
			continue
		}
		rep.Recurring.CommittedCents += r.AmountCents
		rep.Recurring.EntryCount += r.Count
		if r.Pending {
			rep.Recurring.PendingCents += r.AmountCents
			rep.Recurring.PendingCount += r.Count
		}
	}
	return rep, nil
}

func categoryAmounts(rows []Row, cats Categories, month, prev, typ string) []CategoryAmount {
	totals := topTotals(rows, typ)
	byTop := map[int64]*CategoryAmount{}
	leaves := map[int64]map[int64]int64{} // top -> leaf -> cents
	for _, r := range rows {
		if r.Month != month || r.Type != typ {
			continue
		}
		ca, ok := byTop[r.TopCategoryID]
		if !ok {
			c := cats.get(r.TopCategoryID)
			ca = &CategoryAmount{
				CategoryID:     c.ID,
				Name:           c.Name,
				Icon:           c.Icon,
				Color:          c.Color,
				Type:           typ,
				PrevMonthCents: totals[topKey{prev, c.ID}],
				Avg3Cents:      avg3(totals, month, c.ID),
				Subcategories:  []Subcategory{},
			}
			byTop[r.TopCategoryID] = ca
			leaves[r.TopCategoryID] = map[int64]int64{}
		}
		ca.AmountCents += r.AmountCents
		if r.Pending {
			ca.PendingCents += r.AmountCents
		}
		leaves[r.TopCategoryID][r.CategoryID] += r.AmountCents
	}

	out := make([]CategoryAmount, 0, len(byTop))
	for top, ca := range byTop {
		if ca.AmountCents <= 0 {
			continue
		}
		for leaf, cents := range leaves[top] {
			if cents <= 0 {
				continue
			}
			sub := Subcategory{CategoryID: leaf, AmountCents: cents}
			if leaf == top {
				sub.Name, sub.Color = GeneralSubcategory, ca.Color
			} else {
				c := cats.get(leaf)
				sub.Name, sub.Color = c.Name, c.Color
			}
			ca.Subcategories = append(ca.Subcategories, sub)
		}
		slices.SortFunc(ca.Subcategories, func(a, b Subcategory) int {
			return byAmountDesc(a.AmountCents, b.AmountCents, a.Name, b.Name, a.CategoryID, b.CategoryID)
		})
		out = append(out, *ca)
	}
	slices.SortFunc(out, func(a, b CategoryAmount) int {
		return byAmountDesc(a.AmountCents, b.AmountCents, a.Name, b.Name, a.CategoryID, b.CategoryID)
	})
	return out
}
