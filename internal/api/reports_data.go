package api

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/melojms/mm-budget/internal/clock"
	"github.com/melojms/mm-budget/internal/reports"
	"github.com/melojms/mm-budget/internal/store"
)

// parsePayer parses the optional payer_id filter. Unknown or malformed ids are
// a 400 (written here, ok=false).
func (s *Server) parsePayer(w http.ResponseWriter, r *http.Request) (*int64, bool) {
	id, ok := queryInt64(r, "payer_id")
	if !ok || (id != nil && *id <= 0) {
		writeError(w, http.StatusBadRequest, "invalid payer_id")
		return nil, false
	}
	if id == nil {
		return nil, true
	}
	if _, err := s.Q.GetPerson(r.Context(), *id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusBadRequest, "unknown payer_id")
		} else {
			s.internalError(w, r, err)
		}
		return nil, false
	}
	return id, true
}

// payerArg turns an optional payer id into a sqlc nullable argument.
func payerArg(p *int64) any {
	if p == nil {
		return nil
	}
	return *p
}

// reportData loads aggregated rows for months from..to (inclusive) plus all categories.
func (s *Server) reportData(r *http.Request, fromMonth, toMonth string, payer *int64) ([]reports.Row, reports.Categories, error) {
	from, _, err := clock.MonthRange(fromMonth)
	if err != nil {
		return nil, nil, err
	}
	_, to, err := clock.MonthRange(toMonth)
	if err != nil {
		return nil, nil, err
	}
	return s.reportDataDates(r, from, to, payer)
}

// reportDataDates loads aggregated rows for dates from..to (inclusive) plus all categories.
func (s *Server) reportDataDates(r *http.Request, from, to string, payer *int64) ([]reports.Row, reports.Categories, error) {
	ctx := r.Context()
	agg, err := s.Q.ReportAggregate(ctx, store.ReportAggregateParams{FromDate: from, ToDate: to, PayerID: payerArg(payer)})
	if err != nil {
		return nil, nil, err
	}
	rows := make([]reports.Row, len(agg))
	for i, a := range agg {
		rows[i] = reports.Row{
			Month:         a.Month,
			Type:          a.Type,
			CategoryID:    a.CategoryID,
			TopCategoryID: a.TopCategoryID,
			Pending:       a.Pending,
			Recurring:     a.Recurring,
			AmountCents:   a.AmountCents,
			Count:         a.EntryCount,
		}
	}

	all, err := s.Q.ReportCategories(ctx)
	if err != nil {
		return nil, nil, err
	}
	cats := make(reports.Categories, len(all))
	for _, c := range all {
		cats[c.ID] = reports.Category{ID: c.ID, ParentID: c.ParentID, Type: c.Type, Name: c.Name, Icon: c.Icon, Color: c.Color}
	}
	return rows, cats, nil
}

// insightInput gathers everything reports.BuildInsights needs for month.
func (s *Server) insightInput(r *http.Request, month string, payer *int64) (reports.InsightInput, error) {
	ctx := r.Context()
	in := reports.InsightInput{Month: month, Today: clock.Today(s.Clock)}

	from, to, err := reports.MonthWindow(month)
	if err != nil {
		return in, err
	}
	if in.Rows, in.Cats, err = s.reportData(r, from, to, payer); err != nil {
		return in, err
	}

	first, last, err := clock.MonthRange(month)
	if err != nil {
		return in, err
	}
	if month == clock.MonthOf(in.Today) {
		in.NonRecurringSoFar, err = s.Q.ReportNonRecurringExpenseSum(ctx, store.ReportNonRecurringExpenseSumParams{
			FromDate: first, ToDate: in.Today, PayerID: payerArg(payer),
		})
		if err != nil {
			return in, err
		}
	}

	top, err := s.Q.ReportTopExpenses(ctx, store.ReportTopExpensesParams{
		FromDate: first, ToDate: last, PayerID: payerArg(payer), MaxRows: reports.TopExpenseCount,
	})
	if err != nil {
		return in, err
	}
	for _, e := range top {
		in.Top = append(in.Top, reports.TopEntry{
			ID: e.ID, Date: e.Date, AmountCents: e.AmountCents,
			CategoryID: e.CategoryID, CategoryName: e.CategoryName, Note: e.Note,
		})
	}

	tpls, err := s.Q.ReportTemplatesStarting(ctx, store.ReportTemplatesStartingParams{Month: month, PayerID: payerArg(payer)})
	if err != nil {
		return in, err
	}
	for _, t := range tpls {
		in.NewTemplates = append(in.NewTemplates, reports.NewTemplate{
			ID: t.ID, Type: t.Type, CategoryID: t.CategoryID, CategoryName: t.CategoryName,
			AmountCents: t.AmountCents, Variable: t.Variable, Note: t.Note,
		})
	}
	return in, nil
}
