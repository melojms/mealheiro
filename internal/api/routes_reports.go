package api

import (
	"net/http"
	"slices"
	"strconv"

	"github.com/melojms/mealheiro/internal/clock"
	"github.com/melojms/mealheiro/internal/reports"
)

// Limits of the trends window.
const (
	defaultTrendMonths = 12
	maxTrendMonths     = 120
)

// routesReports registers this domain's endpoints. See docs/API.md.
func (s *Server) routesReports(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/reports/month", s.handleMonthReport)
	mux.HandleFunc("GET /api/reports/trends", s.handleTrendsReport)
	mux.HandleFunc("GET /api/reports/year", s.handleYearReport)
	mux.HandleFunc("GET /api/reports/years", s.handleReportYears)
	mux.HandleFunc("GET /api/insights", s.handleInsights)
}

func (s *Server) handleMonthReport(w http.ResponseWriter, r *http.Request) {
	month, payer, ok := s.monthAndPayer(w, r)
	if !ok {
		return
	}
	from, to, err := reports.MonthWindow(month)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	rows, cats, err := s.reportData(r, from, to, payer)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	rep, err := reports.BuildMonth(month, rows, cats)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, rep)
}

func (s *Server) handleTrendsReport(w http.ResponseWriter, r *http.Request) {
	payer, ok := s.parsePayer(w, r)
	if !ok {
		return
	}
	q := r.URL.Query()
	end := q.Get("end")
	if end == "" {
		end = clock.CurrentMonth(s.Clock)
	} else if _, err := clock.ParseMonth(end); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	n := defaultTrendMonths
	if raw := q.Get("months"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 1 || v > maxTrendMonths {
			writeError(w, http.StatusBadRequest, "months must be an integer between 1 and "+strconv.Itoa(maxTrendMonths))
			return
		}
		n = v
	}
	categoryID, ok := queryInt64(r, "category_id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid category_id")
		return
	}

	start, err := clock.AddMonths(end, -(n - 1))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	months, err := clock.MonthsBetween(start, end)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	rows, cats, err := s.reportData(r, start, end, payer)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	if categoryID != nil {
		c, found := cats[*categoryID]
		if !found || c.ParentID != nil || c.Type != reports.TypeExpense {
			writeError(w, http.StatusBadRequest, "category_id must be a top-level expense category")
			return
		}
	}
	writeJSON(w, http.StatusOK, reports.BuildTrends(months, rows, cats, categoryID))
}

func (s *Server) handleYearReport(w http.ResponseWriter, r *http.Request) {
	payer, ok := s.parsePayer(w, r)
	if !ok {
		return
	}
	year := s.Clock.Now().Year()
	if raw := r.URL.Query().Get("year"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 1000 || v > 9999 {
			writeError(w, http.StatusBadRequest, "invalid year")
			return
		}
		year = v
	}
	from, to := reports.YearWindow(year)
	rows, cats, err := s.reportDataDates(r, from, to, payer)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, reports.BuildYear(year, rows, cats))
}

func (s *Server) handleReportYears(w http.ResponseWriter, r *http.Request) {
	payer, ok := s.parsePayer(w, r)
	if !ok {
		return
	}
	years, err := s.Q.ReportEntryYears(r.Context(), payerArg(payer))
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	if cur := int64(s.Clock.Now().Year()); !slices.Contains(years, cur) {
		years = append(years, cur)
		slices.Sort(years)
		slices.Reverse(years)
	}
	writeJSON(w, http.StatusOK, years)
}

func (s *Server) handleInsights(w http.ResponseWriter, r *http.Request) {
	month, payer, ok := s.monthAndPayer(w, r)
	if !ok {
		return
	}
	in, err := s.insightInput(r, month, payer)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	insights, err := reports.BuildInsights(in)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, insights)
}

// monthAndPayer parses the optional month (default current) and payer_id params,
// writing a 400 and returning ok=false when invalid.
func (s *Server) monthAndPayer(w http.ResponseWriter, r *http.Request) (month string, payer *int64, ok bool) {
	payer, ok = s.parsePayer(w, r)
	if !ok {
		return "", nil, false
	}
	month = r.URL.Query().Get("month")
	if month == "" {
		return clock.CurrentMonth(s.Clock), payer, true
	}
	if _, err := clock.ParseMonth(month); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return "", nil, false
	}
	return month, payer, true
}
