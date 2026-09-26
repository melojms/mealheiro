package api

import "net/http"

// routesOps registers this domain's endpoints. See docs/API.md.
func (s *Server) routesOps(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/templates", s.handleListTemplates)
	mux.HandleFunc("POST /api/templates", s.handleCreateTemplate)
	mux.HandleFunc("PUT /api/templates/{id}", s.handleUpdateTemplate)
	mux.HandleFunc("DELETE /api/templates/{id}", s.handleDeleteTemplate)
	mux.HandleFunc("POST /api/recurring/run", s.handleRunRecurring)
	mux.HandleFunc("GET /api/pending", s.handleListPending)

	mux.HandleFunc("GET /api/budgets", s.handleListBudgets)
	mux.HandleFunc("PUT /api/budgets", s.handlePutBudget)
	mux.HandleFunc("GET /api/budgets/status", s.handleBudgetStatus)

	mux.HandleFunc("GET /api/export.csv", s.handleExportCSV)
	mux.HandleFunc("GET /api/backup", s.handleBackup)
}
