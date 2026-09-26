package api

import (
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/melojms/mealheiro/internal/store"
)

func (s *Server) handleListPeople(w http.ResponseWriter, r *http.Request) {
	people, err := s.Q.ListPeople(r.Context())
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	out := make([]Person, 0, len(people))
	for _, p := range people {
		out = append(out, personFromStore(p))
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleRenamePerson(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r)
	if !ok {
		writeError(w, http.StatusNotFound, "person not found")
		return
	}
	var body struct {
		Name *string `json:"name"`
	}
	if err := decodeJSON(w, r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body: "+err.Error())
		return
	}
	if body.Name == nil {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	name := strings.TrimSpace(*body.Name)
	if name == "" || utf8.RuneCountInString(name) > coreMaxNameLen {
		writeError(w, http.StatusBadRequest, "name must be 1-40 characters")
		return
	}
	p, err := s.Q.RenamePerson(r.Context(), store.RenamePersonParams{Name: name, ID: id})
	if coreIsNoRows(err) {
		writeError(w, http.StatusNotFound, "person not found")
		return
	}
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, personFromStore(p))
}
