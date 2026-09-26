package api

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/melojms/mealheiro/internal/clock"
	"github.com/melojms/mealheiro/internal/store"
)

const (
	coreMaxNoteLen        = 500
	coreDefaultEntryLimit = 50
	coreMaxEntryLimit     = 500
)

// coreEntryInput is the EntryInput request body.
type coreEntryInput struct {
	Type        string   `json:"type"`
	Date        string   `json:"date"`
	AmountCents int64    `json:"amount_cents"`
	CategoryID  int64    `json:"category_id"`
	PayerID     int64    `json:"payer_id"`
	Note        string   `json:"note"`
	Tags        []string `json:"tags"`
}

// validate checks in against the DB and normalizes note and tags. An archived
// category is accepted only when it equals keepCategoryID (unchanged on update).
func (in *coreEntryInput) validate(ctx context.Context, q *store.Queries, keepCategoryID int64) error {
	if !coreValidEntryType(in.Type) {
		return coreBadRequest{"invalid type"}
	}
	if _, err := clock.ParseDate(in.Date); err != nil {
		return coreBadRequest{err.Error()}
	}
	if in.AmountCents <= 0 {
		return coreBadRequest{"amount_cents must be > 0"}
	}
	in.Note = strings.TrimSpace(in.Note)
	if utf8.RuneCountInString(in.Note) > coreMaxNoteLen {
		return coreBadRequest{"note is too long"}
	}
	tags, err := coreNormalizeTags(in.Tags)
	if err != nil {
		return coreBadRequest{err.Error()}
	}
	in.Tags = tags

	cat, err := q.GetCategory(ctx, in.CategoryID)
	if coreIsNoRows(err) {
		return coreBadRequest{"category not found"}
	}
	if err != nil {
		return err
	}
	if cat.Archived && cat.ID != keepCategoryID {
		return coreBadRequest{"category is archived"}
	}
	if cat.Type != in.Type {
		return coreBadRequest{"category type does not match entry type"}
	}
	if _, err := q.GetPerson(ctx, in.PayerID); coreIsNoRows(err) {
		return coreBadRequest{"payer not found"}
	} else if err != nil {
		return err
	}
	return nil
}

// coreEntryWriteFailed maps an entry write error to a response; false if err is nil.
func (s *Server) coreEntryWriteFailed(w http.ResponseWriter, r *http.Request, err error) bool {
	var bad coreBadRequest
	switch {
	case err == nil:
		return false
	case errors.As(err, &bad):
		writeError(w, http.StatusBadRequest, bad.msg)
	case coreIsNoRows(err):
		writeError(w, http.StatusNotFound, "entry not found")
	default:
		s.internalError(w, r, err)
	}
	return true
}

// coreWriteEntry responds with the entry_view row of id.
func (s *Server) coreWriteEntry(w http.ResponseWriter, r *http.Request, status int, id int64) {
	v, err := s.Q.GetEntryView(r.Context(), id)
	if coreIsNoRows(err) {
		writeError(w, http.StatusNotFound, "entry not found")
		return
	}
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	writeJSON(w, status, entryFromView(v))
}

func (s *Server) handleGetEntry(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r)
	if !ok {
		writeError(w, http.StatusNotFound, "entry not found")
		return
	}
	s.coreWriteEntry(w, r, http.StatusOK, id)
}

func (s *Server) handleCreateEntry(w http.ResponseWriter, r *http.Request) {
	var in coreEntryInput
	if err := decodeJSON(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body: "+err.Error())
		return
	}
	ctx := r.Context()
	var id int64
	err := s.coreInTx(ctx, func(q *store.Queries) error {
		if err := in.validate(ctx, q, 0); err != nil {
			return err
		}
		var err error
		id, err = q.CreateEntry(ctx, store.CreateEntryParams{
			Type: in.Type, Date: in.Date, AmountCents: in.AmountCents,
			CategoryID: in.CategoryID, PayerID: in.PayerID, Note: in.Note, Now: s.coreTimestamp(),
		})
		if err != nil {
			return err
		}
		return coreReplaceEntryTags(ctx, q, id, in.Tags)
	})
	if s.coreEntryWriteFailed(w, r, err) {
		return
	}
	s.coreWriteEntry(w, r, http.StatusCreated, id)
}

func (s *Server) handleUpdateEntry(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r)
	if !ok {
		writeError(w, http.StatusNotFound, "entry not found")
		return
	}
	var in coreEntryInput
	if err := decodeJSON(w, r, &in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body: "+err.Error())
		return
	}
	ctx := r.Context()
	err := s.coreInTx(ctx, func(q *store.Queries) error {
		cur, err := q.GetEntry(ctx, id)
		if err != nil {
			return err
		}
		if err := in.validate(ctx, q, cur.CategoryID); err != nil {
			return err
		}
		err = q.UpdateEntry(ctx, store.UpdateEntryParams{
			Type: in.Type, Date: in.Date, AmountCents: in.AmountCents,
			CategoryID: in.CategoryID, PayerID: in.PayerID, Note: in.Note,
			UpdatedAt: s.coreTimestamp(), ID: id,
		})
		if err != nil {
			return err
		}
		return coreReplaceEntryTags(ctx, q, id, in.Tags)
	})
	if s.coreEntryWriteFailed(w, r, err) {
		return
	}
	s.coreWriteEntry(w, r, http.StatusOK, id)
}

func (s *Server) handleConfirmEntry(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r)
	if !ok {
		writeError(w, http.StatusNotFound, "entry not found")
		return
	}
	var body struct {
		AmountCents *int64 `json:"amount_cents"`
	}
	// The body is optional.
	if r.ContentLength != 0 {
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, http.StatusBadRequest, "invalid body: "+err.Error())
			return
		}
	}
	if body.AmountCents != nil && *body.AmountCents <= 0 {
		writeError(w, http.StatusBadRequest, "amount_cents must be > 0")
		return
	}
	ctx := r.Context()
	err := s.coreInTx(ctx, func(q *store.Queries) error {
		cur, err := q.GetEntry(ctx, id)
		if err != nil {
			return err
		}
		if cur.Status != "pending" {
			return coreBadRequest{"entry is not pending"}
		}
		amount := cur.AmountCents
		if body.AmountCents != nil {
			amount = *body.AmountCents
		}
		return q.ConfirmEntry(ctx, store.ConfirmEntryParams{AmountCents: amount, UpdatedAt: s.coreTimestamp(), ID: id})
	})
	if s.coreEntryWriteFailed(w, r, err) {
		return
	}
	s.coreWriteEntry(w, r, http.StatusOK, id)
}

func (s *Server) handleDeleteEntry(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r)
	if !ok {
		writeError(w, http.StatusNotFound, "entry not found")
		return
	}
	n, err := s.Q.DeleteEntry(r.Context(), id)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	if n == 0 {
		writeError(w, http.StatusNotFound, "entry not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// EntryList is the GET /api/entries response.
type EntryList struct {
	Entries    []Entry          `json:"entries"`
	TotalCount int64            `json:"total_count"`
	Totals     EntryTotals      `json:"totals"`
	Breakdown  []EntryBreakdown `json:"breakdown"`
}

type EntryTotals struct {
	ExpenseCents    int64 `json:"expense_cents"`
	IncomeCents     int64 `json:"income_cents"`
	InvestmentCents int64 `json:"investment_cents"`
}

type EntryBreakdown struct {
	CategoryID  int64  `json:"category_id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Color       string `json:"color"`
	Icon        string `json:"icon"`
	AmountCents int64  `json:"amount_cents"`
}

// coreEntryFilter is a parsed GET /api/entries query rendered as a WHERE clause
// over entry_view aliased as v.
type coreEntryFilter struct {
	where  []string
	args   []any
	limit  int64
	offset int64
}

func (f *coreEntryFilter) add(cond string, args ...any) {
	f.where = append(f.where, cond)
	f.args = append(f.args, args...)
}

func (f *coreEntryFilter) clause() string {
	if len(f.where) == 0 {
		return ""
	}
	return " WHERE " + strings.Join(f.where, " AND ")
}

// coreParseEntryFilter validates the query string; the error message is user-facing.
func coreParseEntryFilter(r *http.Request) (*coreEntryFilter, error) {
	qs := r.URL.Query()
	f := &coreEntryFilter{limit: coreDefaultEntryLimit}

	for _, key := range []string{"from", "to"} {
		d := qs.Get(key)
		if d == "" {
			continue
		}
		if _, err := clock.ParseDate(d); err != nil {
			return nil, errors.New("invalid " + key + ": " + err.Error())
		}
		if key == "from" {
			f.add("v.date >= ?", d)
		} else {
			f.add("v.date <= ?", d)
		}
	}
	if t := qs.Get("type"); t != "" {
		if !coreValidEntryType(t) {
			return nil, errors.New("invalid type")
		}
		f.add("v.type = ?", t)
	}
	if st := qs.Get("status"); st != "" {
		if st != "confirmed" && st != "pending" {
			return nil, errors.New("invalid status")
		}
		f.add("v.status = ?", st)
	}

	ints := map[string]*int64{}
	for _, key := range []string{"category_id", "payer_id", "min_cents", "max_cents", "limit", "offset"} {
		v, ok := queryInt64(r, key)
		if !ok || (v != nil && *v < 0) {
			return nil, errors.New("invalid " + key)
		}
		ints[key] = v
	}
	if v := ints["category_id"]; v != nil {
		f.add("(v.category_id = ? OR v.parent_category_id = ?)", *v, *v)
	}
	if v := ints["payer_id"]; v != nil {
		f.add("v.payer_id = ?", *v)
	}
	if v := ints["min_cents"]; v != nil {
		f.add("v.amount_cents >= ?", *v)
	}
	if v := ints["max_cents"]; v != nil {
		f.add("v.amount_cents <= ?", *v)
	}
	if v := ints["limit"]; v != nil {
		if *v == 0 {
			return nil, errors.New("invalid limit")
		}
		f.limit = min(*v, coreMaxEntryLimit)
	}
	if v := ints["offset"]; v != nil {
		f.offset = *v
	}

	if tag := strings.ToLower(strings.TrimSpace(qs.Get("tag"))); tag != "" {
		f.add(`EXISTS (SELECT 1 FROM entry_tags et JOIN tags t ON t.id = et.tag_id
			WHERE et.entry_id = v.id AND t.name = ?)`, tag)
	}
	if q := strings.TrimSpace(qs.Get("q")); q != "" {
		p := "%" + coreLikeEscape(strings.ToLower(q)) + "%"
		f.add(`(lower(v.note) LIKE ? ESCAPE '\' OR lower(v.category_name) LIKE ? ESCAPE '\'
			OR lower(v.parent_category_name) LIKE ? ESCAPE '\'
			OR EXISTS (SELECT 1 FROM entry_tags et JOIN tags t ON t.id = et.tag_id
				WHERE et.entry_id = v.id AND t.name LIKE ? ESCAPE '\'))`, p, p, p, p)
	}
	return f, nil
}

const coreEntryViewColumns = `v.id, v.type, v.date, v.amount_cents, v.category_id, v.category_name,
	v.parent_category_id, v.parent_category_name, v.top_category_id, v.payer_id, v.payer_name,
	v.note, v.status, v.template_id, v.template_month, v.tags_csv, v.created_at, v.updated_at`

func (s *Server) handleListEntries(w http.ResponseWriter, r *http.Request) {
	f, err := coreParseEntryFilter(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	list, err := coreQueryEntryList(r.Context(), s.DB, f)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func coreQueryEntryList(ctx context.Context, db store.DBTX, f *coreEntryFilter) (EntryList, error) {
	list := EntryList{Entries: []Entry{}, Breakdown: []EntryBreakdown{}}
	where := f.clause()

	rows, err := db.QueryContext(ctx, "SELECT "+coreEntryViewColumns+" FROM entry_view v"+where+
		" ORDER BY v.date DESC, v.id DESC LIMIT ? OFFSET ?", slices.Concat(f.args, []any{f.limit, f.offset})...)
	if err != nil {
		return list, err
	}
	defer rows.Close()
	for rows.Next() {
		var v store.EntryView
		if err := rows.Scan(&v.ID, &v.Type, &v.Date, &v.AmountCents, &v.CategoryID, &v.CategoryName,
			&v.ParentCategoryID, &v.ParentCategoryName, &v.TopCategoryID, &v.PayerID, &v.PayerName,
			&v.Note, &v.Status, &v.TemplateID, &v.TemplateMonth, &v.TagsCsv, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return list, err
		}
		list.Entries = append(list.Entries, entryFromView(v))
	}
	if err := rows.Err(); err != nil {
		return list, err
	}
	rows.Close()

	err = db.QueryRowContext(ctx, `SELECT count(*),
		coalesce(sum(CASE WHEN v.type = 'expense' THEN v.amount_cents END), 0),
		coalesce(sum(CASE WHEN v.type = 'income' THEN v.amount_cents END), 0),
		coalesce(sum(CASE WHEN v.type = 'investment' THEN v.amount_cents END), 0)
		FROM entry_view v`+where, f.args...).
		Scan(&list.TotalCount, &list.Totals.ExpenseCents, &list.Totals.IncomeCents, &list.Totals.InvestmentCents)
	if err != nil {
		return list, err
	}

	rows, err = db.QueryContext(ctx, `SELECT tc.id, tc.name, tc.type, tc.color, tc.icon, sum(v.amount_cents) AS total
		FROM entry_view v JOIN categories tc ON tc.id = v.top_category_id`+where+`
		GROUP BY tc.id ORDER BY total DESC, tc.name`, f.args...)
	if err != nil {
		return list, err
	}
	defer rows.Close()
	for rows.Next() {
		var b EntryBreakdown
		if err := rows.Scan(&b.CategoryID, &b.Name, &b.Type, &b.Color, &b.Icon, &b.AmountCents); err != nil {
			return list, err
		}
		list.Breakdown = append(list.Breakdown, b)
	}
	return list, rows.Err()
}
