package api

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/melojms/mm-budget/internal/clock"
	"github.com/melojms/mm-budget/internal/recurring"
	"github.com/melojms/mm-budget/internal/store"
)

// Template is the JSON shape of a recurring template.
type Template struct {
	ID                 int64   `json:"id"`
	Type               string  `json:"type"`
	CategoryID         int64   `json:"category_id"`
	CategoryName       string  `json:"category_name"`
	ParentCategoryName *string `json:"parent_category_name"`
	PayerID            int64   `json:"payer_id"`
	PayerName          string  `json:"payer_name"`
	AmountCents        int64   `json:"amount_cents"`
	Variable           bool    `json:"variable"`
	Note               string  `json:"note"`
	StartMonth         string  `json:"start_month"`
	EndMonth           *string `json:"end_month"`
	Active             bool    `json:"active"`
	LastGeneratedMonth *string `json:"last_generated_month"`
}

type templateInput struct {
	Type        string  `json:"type"`
	CategoryID  int64   `json:"category_id"`
	PayerID     int64   `json:"payer_id"`
	AmountCents int64   `json:"amount_cents"`
	Variable    bool    `json:"variable"`
	Note        string  `json:"note"`
	StartMonth  string  `json:"start_month"`
	EndMonth    *string `json:"end_month"`
	Active      *bool   `json:"active"`
}

// PendingEntry is an Entry plus the "from a past month" flag.
type PendingEntry struct {
	Entry
	Stale bool `json:"stale"`
}

func templateFromRow(r store.ListTemplateViewsRow) Template {
	t := Template{
		ID:           r.ID,
		Type:         r.Type,
		CategoryID:   r.CategoryID,
		CategoryName: r.CategoryName,
		PayerID:      r.PayerID,
		PayerName:    r.PayerName,
		AmountCents:  r.AmountCents,
		Variable:     r.Variable,
		Note:         r.Note,
		StartMonth:   r.StartMonth,
		EndMonth:     r.EndMonth,
		Active:       r.Active,
	}
	if r.ParentCategoryName != "" {
		t.ParentCategoryName = &r.ParentCategoryName
	}
	if r.LastGeneratedMonth != "" {
		t.LastGeneratedMonth = &r.LastGeneratedMonth
	}
	return t
}

func knownEntryType(t string) bool {
	return t == "expense" || t == "income" || t == "investment"
}

// isCanonicalMonth reports whether s is a canonical YYYY-MM month.
func isCanonicalMonth(s string) bool {
	t, err := clock.ParseMonth(s)
	return err == nil && t.Format(clock.MonthLayout) == s
}

// validate normalizes in and checks it against the DB. prevCategoryID is the
// template's current category (0 on create): an archived category is only
// accepted when unchanged, so paused templates stay editable.
func (s *Server) validateTemplate(ctx context.Context, in *templateInput, prevCategoryID int64) (msg string, err error) {
	in.Note = strings.TrimSpace(in.Note)
	if in.EndMonth != nil && *in.EndMonth == "" {
		in.EndMonth = nil
	}
	switch {
	case !knownEntryType(in.Type):
		return "type must be expense, income or investment", nil
	case in.AmountCents <= 0:
		return "amount_cents must be > 0", nil
	case !isCanonicalMonth(in.StartMonth):
		return "start_month must be YYYY-MM", nil
	case in.EndMonth != nil && !isCanonicalMonth(*in.EndMonth):
		return "end_month must be YYYY-MM", nil
	case in.EndMonth != nil && *in.EndMonth < in.StartMonth:
		return "end_month must be >= start_month", nil
	}

	cat, err := s.Q.GetTemplateCategory(ctx, in.CategoryID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return "category not found", nil
	case err != nil:
		return "", err
	case cat.Type != in.Type:
		return "category type does not match template type", nil
	case cat.Archived && cat.ID != prevCategoryID:
		return "category is archived", nil
	}

	if _, err := s.Q.GetPerson(ctx, in.PayerID); errors.Is(err, sql.ErrNoRows) {
		return "payer not found", nil
	} else if err != nil {
		return "", err
	}
	return "", nil
}

// findTemplate returns the JSON view of template id.
func (s *Server) findTemplate(ctx context.Context, id int64) (Template, error) {
	rows, err := s.Q.ListTemplateViews(ctx)
	if err != nil {
		return Template{}, err
	}
	for _, r := range rows {
		if r.ID == id {
			return templateFromRow(r), nil
		}
	}
	return Template{}, sql.ErrNoRows
}

// generate runs recurring generation after a template change. Failures are only
// logged: the template is saved and the periodic job will retry.
func (s *Server) generate(ctx context.Context) {
	if _, err := recurring.Generate(ctx, s.DB, s.Clock); err != nil {
		s.Log.Error("recurring generation failed", "err", err)
	}
}

func (s *Server) handleListTemplates(w http.ResponseWriter, r *http.Request) {
	rows, err := s.Q.ListTemplateViews(r.Context())
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	out := make([]Template, 0, len(rows))
	for _, row := range rows {
		out = append(out, templateFromRow(row))
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleCreateTemplate(w http.ResponseWriter, r *http.Request) {
	var in templateInput
	if err := decodeJSON(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	ctx := r.Context()
	if msg, err := s.validateTemplate(ctx, &in, 0); err != nil {
		s.internalError(w, r, err)
		return
	} else if msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}
	id, err := s.Q.CreateTemplate(ctx, store.CreateTemplateParams{
		Type:        in.Type,
		CategoryID:  in.CategoryID,
		PayerID:     in.PayerID,
		AmountCents: in.AmountCents,
		Variable:    in.Variable,
		Note:        in.Note,
		StartMonth:  in.StartMonth,
		EndMonth:    in.EndMonth,
		Active:      in.Active == nil || *in.Active,
	})
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	s.generate(ctx)
	t, err := s.findTemplate(ctx, id)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, t)
}

func (s *Server) handleUpdateTemplate(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var in templateInput
	if err := decodeJSON(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	ctx := r.Context()
	prev, err := s.Q.GetTemplate(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "template not found")
		return
	} else if err != nil {
		s.internalError(w, r, err)
		return
	}
	if msg, err := s.validateTemplate(ctx, &in, prev.CategoryID); err != nil {
		s.internalError(w, r, err)
		return
	} else if msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}
	if _, err := s.Q.UpdateTemplate(ctx, store.UpdateTemplateParams{
		Type:        in.Type,
		CategoryID:  in.CategoryID,
		PayerID:     in.PayerID,
		AmountCents: in.AmountCents,
		Variable:    in.Variable,
		Note:        in.Note,
		StartMonth:  in.StartMonth,
		EndMonth:    in.EndMonth,
		Active:      in.Active == nil || *in.Active,
		ID:          id,
	}); err != nil {
		s.internalError(w, r, err)
		return
	}
	s.generate(ctx)
	t, err := s.findTemplate(ctx, id)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func (s *Server) handleDeleteTemplate(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	n, err := s.Q.DeleteTemplate(r.Context(), id)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	if n == 0 {
		writeError(w, http.StatusNotFound, "template not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleRunRecurring(w http.ResponseWriter, r *http.Request) {
	n, err := recurring.Generate(r.Context(), s.DB, s.Clock)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"created": n})
}

func (s *Server) handleListPending(w http.ResponseWriter, r *http.Request) {
	views, err := s.Q.ListPendingEntryViews(r.Context())
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	current := clock.CurrentMonth(s.Clock)
	out := make([]PendingEntry, 0, len(views))
	for _, v := range views {
		out = append(out, PendingEntry{Entry: entryFromView(v), Stale: clock.MonthOf(v.Date) < current})
	}
	writeJSON(w, http.StatusOK, out)
}
