package api

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/melojms/mm-budget/internal/clock"
	"github.com/melojms/mm-budget/internal/store"
)

// testNow is the frozen "now" for API tests: 2026-03-15 12:00 Europe/Lisbon.
var testNow = time.Date(2026, 3, 15, 12, 0, 0, 0, mustLoc("Europe/Lisbon"))

func mustLoc(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic(err)
	}
	return loc
}

// newTestServer returns a Server backed by a fresh migrated+seeded SQLite DB.
func newTestServer(t *testing.T) (*Server, http.Handler) {
	t.Helper()
	db := store.OpenTest(t)
	s := New(db, clock.Fixed{T: testNow}, slog.New(slog.NewTextHandler(io.Discard, nil)), t.TempDir(), nil)
	return s, s.Routes()
}

// do performs a request with an optional JSON body and returns the recorder.
func do(t *testing.T, h http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var rd io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		rd = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, path, rd)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// decode unmarshals a recorder body into T, failing the test on error.
func decode[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(rec.Body.Bytes(), &v); err != nil {
		t.Fatalf("decode %q: %v", rec.Body.String(), err)
	}
	return v
}

func TestHealthAndMeta(t *testing.T) {
	_, h := newTestServer(t)
	if rec := do(t, h, "GET", "/healthz", nil); rec.Code != http.StatusOK {
		t.Fatalf("healthz = %d", rec.Code)
	}
	rec := do(t, h, "GET", "/api/meta", nil)
	meta := decode[map[string]string](t, rec)
	if meta["today"] != "2026-03-15" || meta["month"] != "2026-03" {
		t.Fatalf("meta = %v", meta)
	}
}

func TestEntryViewMapping(t *testing.T) {
	s, _ := newTestServer(t)
	ctx := t.Context()
	res, err := s.DB.ExecContext(ctx, `INSERT INTO entries (type, date, amount_cents, category_id, payer_id, note) VALUES ('expense','2026-03-01',1250,51,3,'x')`)
	if err != nil {
		t.Fatal(err)
	}
	id, _ := res.LastInsertId()
	if _, err := s.DB.ExecContext(ctx, `INSERT INTO tags (id, name) VALUES (1,'zeta'),(2,'alpha'); INSERT INTO entry_tags VALUES (?,1),(?,2)`, id, id); err != nil {
		t.Fatal(err)
	}
	v, err := s.Q.GetEntryView(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	e := entryFromView(v)
	if e.ParentCategoryName == nil || *e.ParentCategoryName != "House" || e.CategoryName != "Mortgage" || e.PayerName != "Joint" {
		t.Fatalf("unexpected mapping: %+v", e)
	}
	if len(e.Tags) != 2 || e.Tags[0] != "alpha" || e.Tags[1] != "zeta" || e.Recurring {
		t.Fatalf("tags/recurring: %+v", e)
	}
}
