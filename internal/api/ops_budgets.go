package api

import (
	"cmp"
	"context"
	"database/sql"
	"errors"
	"net/http"
	"slices"

	"github.com/melojms/mealheiro/internal/clock"
	"github.com/melojms/mealheiro/internal/store"
)

// Overall cap presentation (docs/API.md "BudgetLine").
const (
	overallBudgetName  = "Overall"
	overallBudgetIcon  = "wallet"
	overallBudgetColor = "#64748b"
)

// BudgetLine is one budget with its usage for a month. CategoryID nil = overall cap.
type BudgetLine struct {
	CategoryID   *int64  `json:"category_id"`
	Name         string  `json:"name"`
	Icon         string  `json:"icon"`
	Color        string  `json:"color"`
	BudgetCents  int64   `json:"budget_cents"`
	SpentCents   int64   `json:"spent_cents"`
	PendingCents int64   `json:"pending_cents"`
	Ratio        float64 `json:"ratio"`
}

type budgetsResponse struct {
	Month      string       `json:"month"`
	Overall    *BudgetLine  `json:"overall"`
	Categories []BudgetLine `json:"categories"`
}

type budgetStatusResponse struct {
	Category *BudgetLine `json:"category"`
	Overall  *BudgetLine `json:"overall"`
}

type budgetInput struct {
	CategoryID  *int64 `json:"category_id"`
	AmountCents *int64 `json:"amount_cents"`
}

// monthBudgets holds everything needed to build budget lines for one month.
type monthBudgets struct {
	overall    *int64          // effective overall cap
	byCategory map[int64]int64 // top-level category id -> effective budget
	spent      map[int64]store.MonthExpenseByTopCategoryRow
	total      store.MonthExpenseByTopCategoryRow // all expenses of the month
	categories map[int64]store.Category           // top-level expense categories
}

// loadMonthBudgets resolves effective budgets and expense usage for month.
func (s *Server) loadMonthBudgets(ctx context.Context, month string) (*monthBudgets, error) {
	rows, err := s.Q.ListBudgetsUpTo(ctx, month)
	if err != nil {
		return nil, err
	}
	mb := &monthBudgets{
		byCategory: map[int64]int64{},
		spent:      map[int64]store.MonthExpenseByTopCategoryRow{},
		categories: map[int64]store.Category{},
	}
	// Rows are oldest first, so later rows override earlier ones.
	for _, b := range rows {
		switch {
		case b.CategoryID == nil:
			mb.overall = b.AmountCents
		case b.AmountCents == nil:
			delete(mb.byCategory, *b.CategoryID)
		default:
			mb.byCategory[*b.CategoryID] = *b.AmountCents
		}
	}

	from, to, err := clock.MonthRange(month)
	if err != nil {
		return nil, err
	}
	usage, err := s.Q.MonthExpenseByTopCategory(ctx, store.MonthExpenseByTopCategoryParams{FromDate: from, ToDate: to})
	if err != nil {
		return nil, err
	}
	for _, u := range usage {
		mb.spent[u.TopCategoryID] = u
		mb.total.SpentCents += u.SpentCents
		mb.total.PendingCents += u.PendingCents
	}

	cats, err := s.Q.ListTopExpenseCategories(ctx)
	if err != nil {
		return nil, err
	}
	for _, c := range cats {
		mb.categories[c.ID] = c
	}
	return mb, nil
}

func newBudgetLine(categoryID *int64, name, icon, color string, budget int64, usage store.MonthExpenseByTopCategoryRow) BudgetLine {
	return BudgetLine{
		CategoryID:   categoryID,
		Name:         name,
		Icon:         icon,
		Color:        color,
		BudgetCents:  budget,
		SpentCents:   usage.SpentCents,
		PendingCents: usage.PendingCents,
		Ratio:        float64(usage.SpentCents) / float64(budget),
	}
}

func (mb *monthBudgets) overallLine() *BudgetLine {
	if mb.overall == nil {
		return nil
	}
	l := newBudgetLine(nil, overallBudgetName, overallBudgetIcon, overallBudgetColor, *mb.overall, mb.total)
	return &l
}

// categoryLine returns the line for a top-level expense category, nil if it has no budget.
func (mb *monthBudgets) categoryLine(id int64) *BudgetLine {
	budget, ok := mb.byCategory[id]
	cat, known := mb.categories[id]
	if !ok || !known {
		return nil
	}
	l := newBudgetLine(&cat.ID, cat.Name, cat.Icon, cat.Color, budget, mb.spent[id])
	return &l
}

// budgetMonthParam returns the ?month= param or the current month; ok=false if invalid.
func (s *Server) budgetMonthParam(r *http.Request) (string, bool) {
	m := r.URL.Query().Get("month")
	if m == "" {
		return clock.CurrentMonth(s.Clock), true
	}
	return m, isCanonicalMonth(m)
}

func (s *Server) handleListBudgets(w http.ResponseWriter, r *http.Request) {
	month, ok := s.budgetMonthParam(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "month must be YYYY-MM")
		return
	}
	mb, err := s.loadMonthBudgets(r.Context(), month)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	resp := budgetsResponse{Month: month, Overall: mb.overallLine(), Categories: []BudgetLine{}}
	for id := range mb.byCategory {
		if l := mb.categoryLine(id); l != nil {
			resp.Categories = append(resp.Categories, *l)
		}
	}
	slices.SortFunc(resp.Categories, func(a, b BudgetLine) int {
		return cmp.Or(cmp.Compare(b.Ratio, a.Ratio), cmp.Compare(a.Name, b.Name))
	})
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handlePutBudget(w http.ResponseWriter, r *http.Request) {
	var in budgetInput
	if err := decodeJSON(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if in.AmountCents != nil && *in.AmountCents <= 0 {
		writeError(w, http.StatusBadRequest, "amount_cents must be > 0 or null")
		return
	}
	ctx := r.Context()
	if in.CategoryID != nil {
		cat, err := s.Q.GetTemplateCategory(ctx, *in.CategoryID)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			writeError(w, http.StatusBadRequest, "category not found")
			return
		case err != nil:
			s.internalError(w, r, err)
			return
		case cat.Type != "expense" || cat.ParentID != nil:
			writeError(w, http.StatusBadRequest, "budgets are only allowed on top-level expense categories")
			return
		case cat.Archived && in.AmountCents != nil:
			writeError(w, http.StatusBadRequest, "category is archived")
			return
		}
	}

	month := clock.CurrentMonth(s.Clock)
	prev, err := clock.AddMonths(month, -1)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	if err := s.upsertBudget(ctx, in, month, prev); err != nil {
		s.internalError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// upsertBudget replaces the row effective from month. A removal only needs a
// NULL row when an earlier budget would otherwise still apply.
func (s *Server) upsertBudget(ctx context.Context, in budgetInput, month, prev string) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	q := s.Q.WithTx(tx)

	if err := q.DeleteBudgetAt(ctx, store.DeleteBudgetAtParams{CategoryID: in.CategoryID, EffectiveFrom: month}); err != nil {
		return err
	}
	insert := in.AmountCents != nil
	if !insert {
		rows, err := q.ListBudgetsUpTo(ctx, prev)
		if err != nil {
			return err
		}
		insert = effectiveAmount(rows, in.CategoryID) != nil
	}
	if insert {
		if err := q.InsertBudget(ctx, store.InsertBudgetParams{
			CategoryID:    in.CategoryID,
			EffectiveFrom: month,
			AmountCents:   in.AmountCents,
		}); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// effectiveAmount returns the effective budget of categoryID (nil = overall)
// from rows ordered oldest first.
func effectiveAmount(rows []store.Budget, categoryID *int64) *int64 {
	var amount *int64
	for _, b := range rows {
		if (b.CategoryID == nil) == (categoryID == nil) && (categoryID == nil || *b.CategoryID == *categoryID) {
			amount = b.AmountCents
		}
	}
	return amount
}

func (s *Server) handleBudgetStatus(w http.ResponseWriter, r *http.Request) {
	month, ok := s.budgetMonthParam(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "month must be YYYY-MM")
		return
	}
	categoryID, ok := queryInt64(r, "category_id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid category_id")
		return
	}
	ctx := r.Context()
	var topID *int64
	if categoryID != nil {
		cat, err := s.Q.GetTemplateCategory(ctx, *categoryID)
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "category not found")
			return
		} else if err != nil {
			s.internalError(w, r, err)
			return
		}
		topID = &cat.ID
		if cat.ParentID != nil {
			topID = cat.ParentID
		}
	}
	mb, err := s.loadMonthBudgets(ctx, month)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	resp := budgetStatusResponse{Overall: mb.overallLine()}
	if topID != nil {
		resp.Category = mb.categoryLine(*topID)
	}
	writeJSON(w, http.StatusOK, resp)
}
