// Package reports turns pre-aggregated entry totals into the month, trends and
// year reports and the month insights described in docs/API.md. It is pure: the
// caller loads rows from the database and this package only does arithmetic.
package reports

import "cmp"

// Entry types.
const (
	TypeExpense    = "expense"
	TypeIncome     = "income"
	TypeInvestment = "investment"
)

// Row is the total of entries sharing month, type, leaf category, status and
// recurring flag (see the ReportAggregate query).
type Row struct {
	Month         string // YYYY-MM
	Type          string
	CategoryID    int64 // leaf (or top-level when booked on the parent)
	TopCategoryID int64
	Pending       bool
	Recurring     bool
	AmountCents   int64
	Count         int64
}

// Category is the category metadata reports need.
type Category struct {
	ID       int64
	ParentID *int64
	Type     string
	Name     string
	Icon     string
	Color    string
}

// Categories indexes categories by id.
type Categories map[int64]Category

func (c Categories) get(id int64) Category {
	if cat, ok := c[id]; ok {
		return cat
	}
	return Category{ID: id, Name: "Unknown", Icon: "circle", Color: "#64748b"}
}

// typeRank orders entry types: expense, income, investment.
func typeRank(t string) int {
	switch t {
	case TypeExpense:
		return 0
	case TypeIncome:
		return 1
	default:
		return 2
	}
}

// divRound divides non-negative-denominator integers rounding half away from zero.
func divRound(num, den int64) int64 {
	if den == 0 {
		return 0
	}
	if num < 0 {
		return -((-num*2 + den) / (den * 2))
	}
	return (num*2 + den) / (den * 2)
}

// savingsRate is (income − expense) / income, nil when income is 0.
func savingsRate(income, expense int64) *float64 {
	if income == 0 {
		return nil
	}
	r := float64(income-expense) / float64(income)
	return &r
}

// byAmountDesc orders by amount desc, then name, then id.
func byAmountDesc(a, b int64, nameA, nameB string, idA, idB int64) int {
	return cmp.Or(cmp.Compare(b, a), cmp.Compare(nameA, nameB), cmp.Compare(idA, idB))
}
