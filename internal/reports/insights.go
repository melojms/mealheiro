package reports

import (
	"cmp"
	"fmt"
	"math"
	"slices"
	"strings"
	"time"

	"github.com/melojms/mealheiro/internal/clock"
)

// Insight thresholds (docs/API.md "Insights").
const (
	// A category change vs its 3-month average must exceed both of these.
	CategoryChangeMinPct   = 20
	CategoryChangeMinCents = 2000
	// Pace is only projected from this day of the current month on.
	PaceMinDay = 5
	// Pace warns when the projection exceeds last month's non-recurring spend by more than this.
	PaceWarnPct = 10
	// Number of largest expenses reported.
	TopExpenseCount = 3
	// Savings rate drops larger than this many percentage points warn.
	SavingsRateWarnDropPts = 5
)

const (
	KindCategoryUp   = "category_up"
	KindCategoryDown = "category_down"
	KindPace         = "pace"
	KindTopExpense   = "top_expense"
	KindSavingsRate  = "savings_rate"
	KindNewRecurring = "new_recurring"

	SeverityWarning = "warning"
	SeverityGood    = "good"
	SeverityInfo    = "info"
)

type Insight struct {
	Kind        string `json:"kind"`
	Severity    string `json:"severity"`
	Title       string `json:"title"`
	Detail      string `json:"detail"`
	CategoryID  *int64 `json:"category_id"`
	EntryID     *int64 `json:"entry_id"`
	AmountCents *int64 `json:"amount_cents"`
}

// TopEntry is one of the month's largest expenses.
type TopEntry struct {
	ID           int64
	Date         string
	AmountCents  int64
	CategoryID   int64
	CategoryName string
	Note         string
}

// NewTemplate is a recurring template starting in the month.
type NewTemplate struct {
	ID           int64
	Type         string
	CategoryID   int64
	CategoryName string
	AmountCents  int64
	Variable     bool
	Note         string
}

// InsightInput is everything BuildInsights needs for one month.
type InsightInput struct {
	Month string
	Today string // current local date YYYY-MM-DD
	Rows  []Row  // must cover MonthWindow(Month)
	Cats  Categories
	// NonRecurringSoFar is the month's non-recurring expense dated up to Today
	// (only used when Month is the current month).
	NonRecurringSoFar int64
	Top               []TopEntry // largest expenses, amount desc
	NewTemplates      []NewTemplate
}

// BuildInsights applies the insight rules and returns them warnings first,
// then good, then info (rule order within a severity).
func BuildInsights(in InsightInput) ([]Insight, error) {
	prev, err := clock.AddMonths(in.Month, -1)
	if err != nil {
		return nil, err
	}
	current := in.Month == clock.MonthOf(in.Today)
	// The current (or a future) month is incomplete: drops are not meaningful yet.
	incomplete := in.Month >= clock.MonthOf(in.Today)

	out := categoryInsights(in, incomplete)
	if current {
		if ins, ok := paceInsight(in, prev); ok {
			out = append(out, ins)
		}
	}
	out = append(out, topExpenseInsights(in.Top)...)
	if ins, ok := savingsRateInsight(KPIsFor(in.Rows, in.Month), KPIsFor(in.Rows, prev)); ok {
		out = append(out, ins)
	}
	for _, t := range in.NewTemplates {
		out = append(out, newRecurringInsight(t))
	}

	slices.SortStableFunc(out, func(a, b Insight) int {
		return cmp.Compare(severityRank(a.Severity), severityRank(b.Severity))
	})
	return out, nil
}

func severityRank(s string) int {
	switch s {
	case SeverityWarning:
		return 0
	case SeverityGood:
		return 1
	default:
		return 2
	}
}

func categoryInsights(in InsightInput, incomplete bool) []Insight {
	totals := topTotals(in.Rows, TypeExpense)
	tops := map[int64]bool{}
	for k := range totals {
		tops[k.top] = true
	}

	type change struct {
		ins  Insight
		diff int64
	}
	var changes []change
	for top := range tops {
		avg := avg3(totals, in.Month, top)
		if avg <= 0 {
			continue
		}
		cur := totals[topKey{in.Month, top}]
		diff := cur - avg
		absDiff := max(diff, -diff)
		if absDiff*100 <= avg*CategoryChangeMinPct || absDiff <= CategoryChangeMinCents {
			continue
		}
		if diff < 0 && incomplete {
			continue
		}
		c := in.Cats.get(top)
		pct := divRound(absDiff*100, avg)
		ins := Insight{
			Kind:        KindCategoryUp,
			Severity:    SeverityWarning,
			Title:       fmt.Sprintf("%s up %d%%", c.Name, pct),
			Detail:      fmt.Sprintf("%s vs %s 3-month average", FormatCents(cur), FormatCents(avg)),
			CategoryID:  new(top),
			AmountCents: new(cur),
		}
		if diff < 0 {
			ins.Kind, ins.Severity = KindCategoryDown, SeverityGood
			ins.Title = fmt.Sprintf("%s down %d%%", c.Name, pct)
		}
		changes = append(changes, change{ins, absDiff})
	}
	// Biggest moves first; category id breaks ties deterministically.
	slices.SortFunc(changes, func(a, b change) int {
		return cmp.Or(cmp.Compare(b.diff, a.diff), cmp.Compare(*a.ins.CategoryID, *b.ins.CategoryID))
	})
	out := make([]Insight, len(changes))
	for i, c := range changes {
		out[i] = c.ins
	}
	return out
}

func paceInsight(in InsightInput, prev string) (Insight, bool) {
	today, err := clock.ParseDate(in.Today)
	if err != nil || today.Day() < PaceMinDay || in.NonRecurringSoFar <= 0 {
		return Insight{}, false
	}
	days, err := clock.DaysIn(in.Month)
	if err != nil {
		return Insight{}, false
	}
	projection := divRound(in.NonRecurringSoFar*int64(days), int64(today.Day()))

	var prevSpend int64
	for _, r := range in.Rows {
		if r.Month == prev && r.Type == TypeExpense && !r.Recurring {
			prevSpend += r.AmountCents
		}
	}

	ins := Insight{
		Kind:        KindPace,
		Severity:    SeverityInfo,
		Title:       "On pace for " + FormatCents(projection),
		Detail:      fmt.Sprintf("%s non-recurring spent in %d days", FormatCents(in.NonRecurringSoFar), today.Day()),
		AmountCents: new(projection),
	}
	if prevSpend > 0 {
		ins.Detail += fmt.Sprintf(" · %s last month", FormatCents(prevSpend))
		if projection*100 > prevSpend*(100+PaceWarnPct) {
			ins.Severity = SeverityWarning
		}
	}
	return ins, true
}

func topExpenseInsights(top []TopEntry) []Insight {
	out := make([]Insight, 0, TopExpenseCount)
	for i, e := range top {
		if i == TopExpenseCount {
			break
		}
		detail := FormatCents(e.AmountCents)
		if d, err := time.Parse(clock.DateLayout, e.Date); err == nil {
			detail += " · " + d.Format("2 Jan")
		}
		if note := strings.TrimSpace(e.Note); note != "" {
			detail += " · " + note
		}
		out = append(out, Insight{
			Kind:        KindTopExpense,
			Severity:    SeverityInfo,
			Title:       "Big expense: " + e.CategoryName,
			Detail:      detail,
			CategoryID:  new(e.CategoryID),
			EntryID:     new(e.ID),
			AmountCents: new(e.AmountCents),
		})
	}
	return out
}

func savingsRateInsight(cur, prev KPIs) (Insight, bool) {
	if cur.SavingsRate == nil || prev.SavingsRate == nil {
		return Insight{}, false
	}
	now, before := *cur.SavingsRate*100, *prev.SavingsRate*100
	delta := now - before
	const eps = 1e-9

	ins := Insight{Kind: KindSavingsRate, Severity: SeverityInfo}
	switch {
	case delta > eps:
		ins.Severity = SeverityGood
		ins.Title = fmt.Sprintf("Savings rate up to %s", pct(now))
	case delta < -SavingsRateWarnDropPts-eps:
		ins.Severity = SeverityWarning
		ins.Title = fmt.Sprintf("Savings rate down to %s", pct(now))
	case delta < -eps:
		ins.Title = fmt.Sprintf("Savings rate down to %s", pct(now))
	default:
		ins.Title = fmt.Sprintf("Savings rate steady at %s", pct(now))
	}
	ins.Detail = fmt.Sprintf("%s last month (%+d pts)", pct(before), int(math.Round(delta)))
	return ins, true
}

func pct(p float64) string { return fmt.Sprintf("%d%%", int(math.Round(p))) }

func newRecurringInsight(t NewTemplate) Insight {
	detail := FormatCents(t.AmountCents) + " / month"
	if t.Variable {
		detail += " (variable)"
	}
	if t.Type != TypeExpense {
		detail += " · " + t.Type
	}
	if note := strings.TrimSpace(t.Note); note != "" {
		detail += " · " + note
	}
	return Insight{
		Kind:        KindNewRecurring,
		Severity:    SeverityInfo,
		Title:       "New recurring: " + t.CategoryName,
		Detail:      detail,
		CategoryID:  new(t.CategoryID),
		AmountCents: new(t.AmountCents),
	}
}
