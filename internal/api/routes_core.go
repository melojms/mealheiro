package api

import "net/http"

// routesCore registers this domain's endpoints. See docs/API.md.
func (s *Server) routesCore(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/people", s.handleListPeople)
	mux.HandleFunc("PATCH /api/people/{id}", s.handleRenamePerson)

	mux.HandleFunc("GET /api/categories", s.handleListCategories)
	mux.HandleFunc("POST /api/categories", s.handleCreateCategory)
	mux.HandleFunc("PATCH /api/categories/{id}", s.handleUpdateCategory)
	mux.HandleFunc("DELETE /api/categories/{id}", s.handleDeleteCategory)

	mux.HandleFunc("GET /api/entries", s.handleListEntries)
	mux.HandleFunc("GET /api/entries/{id}", s.handleGetEntry)
	mux.HandleFunc("POST /api/entries", s.handleCreateEntry)
	mux.HandleFunc("PUT /api/entries/{id}", s.handleUpdateEntry)
	mux.HandleFunc("POST /api/entries/{id}/confirm", s.handleConfirmEntry)
	mux.HandleFunc("DELETE /api/entries/{id}", s.handleDeleteEntry)

	mux.HandleFunc("GET /api/tags", s.handleListTags)
}
