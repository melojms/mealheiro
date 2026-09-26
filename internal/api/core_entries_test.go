package api

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func coreEntryBody(mods ...func(map[string]any)) map[string]any {
	b := map[string]any{"type": "expense", "date": "2026-03-10", "amount_cents": 1250, "category_id": 6, "payer_id": 1}
	for _, m := range mods {
		m(b)
	}
	return b
}

func coreWith(k string, v any) func(map[string]any) {
	return func(b map[string]any) { b[k] = v }
}

func coreWithout(k string) func(map[string]any) {
	return func(b map[string]any) { delete(b, k) }
}

// coreCreateEntry creates an entry through the API and returns it.
func coreCreateEntry(t *testing.T, h http.Handler, body map[string]any) Entry {
	t.Helper()
	rec := do(t, h, "POST", "/api/entries", body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create entry: %d %s", rec.Code, rec.Body)
	}
	return decode[Entry](t, rec)
}

func TestCreateEntry(t *testing.T) {
	tests := []struct {
		name     string
		setup    string
		body     map[string]any
		wantCode int
		check    func(t *testing.T, e Entry)
	}{
		{
			name: "minimal", body: coreEntryBody(), wantCode: 201,
			check: func(t *testing.T, e Entry) {
				if e.ID == 0 || e.Type != "expense" || e.Date != "2026-03-10" || e.AmountCents != 1250 ||
					e.CategoryName != "Groceries" || e.ParentCategoryID != nil || e.ParentCategoryName != nil ||
					e.PayerName != "Person A" || e.Note != "" || len(e.Tags) != 0 || e.Status != "confirmed" ||
					e.TemplateID != nil || e.Recurring {
					t.Fatalf("got %+v", e)
				}
				if e.CreatedAt != "2026-03-15T12:00:00Z" || e.UpdatedAt != e.CreatedAt {
					t.Fatalf("timestamps %s %s", e.CreatedAt, e.UpdatedAt)
				}
			},
		},
		{
			name: "subcategory, note, tags normalized",
			body: coreEntryBody(coreWith("category_id", 51), coreWith("payer_id", 3), coreWith("note", "  march  "),
				coreWith("tags", []string{" Trip ", "trip", "", "ALPHA", "  "})),
			wantCode: 201,
			check: func(t *testing.T, e Entry) {
				if e.ParentCategoryID == nil || *e.ParentCategoryID != 5 || *e.ParentCategoryName != "House" ||
					e.Note != "march" || strings.Join(e.Tags, ",") != "alpha,trip" {
					t.Fatalf("got %+v", e)
				}
			},
		},
		{name: "income", body: coreEntryBody(coreWith("type", "income"), coreWith("category_id", 100)), wantCode: 201},
		{name: "tag 40 chars", body: coreEntryBody(coreWith("tags", []string{strings.Repeat("a", 40)})), wantCode: 201},
		{name: "tag 41 chars", body: coreEntryBody(coreWith("tags", []string{strings.Repeat("a", 41)})), wantCode: 400},
		{name: "zero amount", body: coreEntryBody(coreWith("amount_cents", 0)), wantCode: 400},
		{name: "negative amount", body: coreEntryBody(coreWith("amount_cents", -5)), wantCode: 400},
		{name: "fractional amount", body: coreEntryBody(coreWith("amount_cents", 12.5)), wantCode: 400},
		{name: "bad date", body: coreEntryBody(coreWith("date", "2026-02-30")), wantCode: 400},
		{name: "missing date", body: coreEntryBody(coreWithout("date")), wantCode: 400},
		{name: "bad type", body: coreEntryBody(coreWith("type", "transfer")), wantCode: 400},
		{name: "type mismatch", body: coreEntryBody(coreWith("category_id", 100)), wantCode: 400},
		{name: "unknown category", body: coreEntryBody(coreWith("category_id", 999)), wantCode: 400},
		{name: "archived category", setup: "UPDATE categories SET archived = TRUE WHERE id = 6", body: coreEntryBody(), wantCode: 400},
		{name: "unknown payer", body: coreEntryBody(coreWith("payer_id", 9)), wantCode: 400},
		{name: "note too long", body: coreEntryBody(coreWith("note", strings.Repeat("x", 501))), wantCode: 400},
		{name: "unknown field", body: coreEntryBody(coreWith("status", "pending")), wantCode: 400},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, h := newTestServer(t)
			if tt.setup != "" {
				coreExec(t, s, tt.setup)
			}
			rec := do(t, h, "POST", "/api/entries", tt.body)
			if rec.Code != tt.wantCode {
				t.Fatalf("status = %d, want %d (%s)", rec.Code, tt.wantCode, rec.Body)
			}
			if tt.wantCode != http.StatusCreated {
				if e := decode[errorBody](t, rec); e.Error == "" {
					t.Fatal("missing error message")
				}
				return
			}
			e := decode[Entry](t, rec)
			if tt.check != nil {
				tt.check(t, e)
			}
			got := decode[Entry](t, do(t, h, "GET", fmt.Sprintf("/api/entries/%d", e.ID), nil))
			if got.ID != e.ID || strings.Join(got.Tags, ",") != strings.Join(e.Tags, ",") {
				t.Fatalf("GET = %+v, want %+v", got, e)
			}
		})
	}
}

func TestGetEntryNotFound(t *testing.T) {
	_, h := newTestServer(t)
	for _, p := range []string{"/api/entries/999", "/api/entries/abc", "/api/entries/-1"} {
		if rec := do(t, h, "GET", p, nil); rec.Code != http.StatusNotFound {
			t.Errorf("%s: %d", p, rec.Code)
		}
	}
}

func TestUpdateEntry(t *testing.T) {
	tests := []struct {
		name     string
		setup    string // runs after the entry (id 1, Groceries, tags a,b) is created
		body     map[string]any
		wantCode int
		check    func(t *testing.T, e Entry)
	}{
		{
			name: "full update replaces tags",
			body: coreEntryBody(coreWith("type", "investment"), coreWith("category_id", 202), coreWith("date", "2026-01-31"),
				coreWith("amount_cents", 99), coreWith("payer_id", 2), coreWith("note", "btc"), coreWith("tags", []string{"B", "c"})),
			wantCode: 200,
			check: func(t *testing.T, e Entry) {
				if e.Type != "investment" || e.CategoryName != "BTC" || e.Date != "2026-01-31" || e.AmountCents != 99 ||
					e.PayerID != 2 || e.Note != "btc" || strings.Join(e.Tags, ",") != "b,c" {
					t.Fatalf("got %+v", e)
				}
			},
		},
		{
			name: "omitted tags clears them", body: coreEntryBody(), wantCode: 200,
			check: func(t *testing.T, e Entry) {
				if len(e.Tags) != 0 || e.Tags == nil {
					t.Fatalf("tags = %#v", e.Tags)
				}
			},
		},
		{
			name: "archived category allowed if unchanged", setup: "UPDATE categories SET archived = TRUE WHERE id = 6",
			body: coreEntryBody(coreWith("amount_cents", 5)), wantCode: 200,
		},
		{
			name: "archived category rejected if changed", setup: "UPDATE categories SET archived = TRUE WHERE id = 12",
			body: coreEntryBody(coreWith("category_id", 12)), wantCode: 400,
		},
		{
			name: "status unchanged (pending stays pending)", setup: "UPDATE entries SET status = 'pending' WHERE id = 1",
			body: coreEntryBody(), wantCode: 200,
			check: func(t *testing.T, e Entry) {
				if e.Status != "pending" {
					t.Fatalf("status = %s", e.Status)
				}
			},
		},
		{
			name: "recurring flag kept", setup: "UPDATE entries SET template_month = '2026-03' WHERE id = 1",
			body: coreEntryBody(), wantCode: 200,
			check: func(t *testing.T, e Entry) {
				if !e.Recurring {
					t.Fatal("recurring lost")
				}
			},
		},
		{name: "invalid", body: coreEntryBody(coreWith("amount_cents", 0)), wantCode: 400},
		{name: "type mismatch", body: coreEntryBody(coreWith("type", "income")), wantCode: 400},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, h := newTestServer(t)
			orig := coreCreateEntry(t, h, coreEntryBody(coreWith("tags", []string{"a", "b"})))
			coreExec(t, s, `UPDATE entries SET created_at = '2026-01-01T00:00:00Z', updated_at = '2026-01-01T00:00:00Z'`)
			if tt.setup != "" {
				coreExec(t, s, tt.setup)
			}
			rec := do(t, h, "PUT", fmt.Sprintf("/api/entries/%d", orig.ID), tt.body)
			if rec.Code != tt.wantCode {
				t.Fatalf("status = %d, want %d (%s)", rec.Code, tt.wantCode, rec.Body)
			}
			if tt.wantCode != http.StatusOK {
				// Nothing changed.
				got := decode[Entry](t, do(t, h, "GET", fmt.Sprintf("/api/entries/%d", orig.ID), nil))
				if got.AmountCents != orig.AmountCents || strings.Join(got.Tags, ",") != "a,b" {
					t.Fatalf("entry changed on failed update: %+v", got)
				}
				return
			}
			e := decode[Entry](t, rec)
			if e.UpdatedAt != "2026-03-15T12:00:00Z" || e.CreatedAt != "2026-01-01T00:00:00Z" {
				t.Fatalf("timestamps: created %s updated %s", e.CreatedAt, e.UpdatedAt)
			}
			if tt.check != nil {
				tt.check(t, e)
			}
		})
	}

	t.Run("not found", func(t *testing.T) {
		_, h := newTestServer(t)
		if rec := do(t, h, "PUT", "/api/entries/42", coreEntryBody()); rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d", rec.Code)
		}
	})
}

func TestConfirmEntry(t *testing.T) {
	tests := []struct {
		name       string
		pending    bool
		body       any
		wantCode   int
		wantAmount int64
	}{
		{"keeps amount (empty object)", true, map[string]any{}, 200, 1250},
		{"keeps amount (no body)", true, nil, 200, 1250},
		{"updates amount", true, map[string]any{"amount_cents": 4321}, 200, 4321},
		{"null amount keeps", true, map[string]any{"amount_cents": nil}, 200, 1250},
		{"zero amount", true, map[string]any{"amount_cents": 0}, 400, 0},
		{"unknown field", true, map[string]any{"amount": 1}, 400, 0},
		{"not pending", false, map[string]any{}, 400, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, h := newTestServer(t)
			e := coreCreateEntry(t, h, coreEntryBody())
			if tt.pending {
				coreExec(t, s, `UPDATE entries SET status = 'pending', updated_at = '2026-01-01T00:00:00Z'`)
			}
			rec := do(t, h, "POST", fmt.Sprintf("/api/entries/%d/confirm", e.ID), tt.body)
			if rec.Code != tt.wantCode {
				t.Fatalf("status = %d, want %d (%s)", rec.Code, tt.wantCode, rec.Body)
			}
			if tt.wantCode != http.StatusOK {
				return
			}
			got := decode[Entry](t, rec)
			if got.Status != "confirmed" || got.AmountCents != tt.wantAmount || got.UpdatedAt != "2026-03-15T12:00:00Z" {
				t.Fatalf("got %+v", got)
			}
		})
	}
	t.Run("not found", func(t *testing.T) {
		_, h := newTestServer(t)
		if rec := do(t, h, "POST", "/api/entries/7/confirm", map[string]any{}); rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d", rec.Code)
		}
	})
}

func TestDeleteEntry(t *testing.T) {
	_, h := newTestServer(t)
	e := coreCreateEntry(t, h, coreEntryBody(coreWith("tags", []string{"x"})))
	if rec := do(t, h, "DELETE", fmt.Sprintf("/api/entries/%d", e.ID), nil); rec.Code != http.StatusNoContent {
		t.Fatalf("delete = %d", rec.Code)
	}
	if rec := do(t, h, "GET", fmt.Sprintf("/api/entries/%d", e.ID), nil); rec.Code != http.StatusNotFound {
		t.Fatalf("get after delete = %d", rec.Code)
	}
	if rec := do(t, h, "DELETE", fmt.Sprintf("/api/entries/%d", e.ID), nil); rec.Code != http.StatusNotFound {
		t.Fatalf("second delete = %d", rec.Code)
	}
	if tags := decode[[]TagCount](t, do(t, h, "GET", "/api/tags", nil)); len(tags) != 0 {
		t.Fatalf("tags after delete = %+v", tags)
	}
}

// coreSeedEntries creates a fixed set of entries used by the list tests:
//
//	1 expense 2026-01-10  1000 Groceries   A  "Pingo Doce"   [food]
//	2 expense 2026-02-01  5000 Mortgage    J  ""             [house]
//	3 expense 2026-02-15   300 House       B  "100%_off"     []
//	4 income  2026-02-28 20000 Salary      A  ""             [food, work]
//	5 expense 2026-03-01  2500 Taxi/Uber   B  "airport"      [trip]   (pending)
//	6 investment 2026-03-05 7000 BTC       J  ""             []
//	7 expense 2026-03-05   700 Groceries   A  ""             [trip]
func coreSeedEntries(t *testing.T, s *Server, h http.Handler) {
	t.Helper()
	rows := []map[string]any{
		coreEntryBody(coreWith("date", "2026-01-10"), coreWith("amount_cents", 1000), coreWith("note", "Pingo Doce"), coreWith("tags", []string{"food"})),
		coreEntryBody(coreWith("date", "2026-02-01"), coreWith("amount_cents", 5000), coreWith("category_id", 51), coreWith("payer_id", 3), coreWith("tags", []string{"house"})),
		coreEntryBody(coreWith("date", "2026-02-15"), coreWith("amount_cents", 300), coreWith("category_id", 5), coreWith("payer_id", 2), coreWith("note", "100%_off")),
		coreEntryBody(coreWith("type", "income"), coreWith("date", "2026-02-28"), coreWith("amount_cents", 20000), coreWith("category_id", 100), coreWith("tags", []string{"food", "work"})),
		coreEntryBody(coreWith("date", "2026-03-01"), coreWith("amount_cents", 2500), coreWith("category_id", 56), coreWith("payer_id", 2), coreWith("note", "airport"), coreWith("tags", []string{"trip"})),
		coreEntryBody(coreWith("type", "investment"), coreWith("date", "2026-03-05"), coreWith("amount_cents", 7000), coreWith("category_id", 202), coreWith("payer_id", 3)),
		coreEntryBody(coreWith("date", "2026-03-05"), coreWith("amount_cents", 700), coreWith("tags", []string{"trip"})),
	}
	for _, b := range rows {
		coreCreateEntry(t, h, b)
	}
	coreExec(t, s, `UPDATE entries SET status = 'pending' WHERE id = 5`)
}

func TestListEntries(t *testing.T) {
	s, h := newTestServer(t)
	coreSeedEntries(t, s, h)

	type bd struct {
		id     int64
		amount int64
	}
	tests := []struct {
		name          string
		query         string
		wantIDs       []int64 // page, in order
		wantTotal     int64
		wantTotals    EntryTotals
		wantBreakdown []bd
	}{
		{
			name: "all", query: "", wantIDs: []int64{7, 6, 5, 4, 3, 2, 1}, wantTotal: 7,
			wantTotals:    EntryTotals{9500, 20000, 7000},
			wantBreakdown: []bd{{100, 20000}, {202, 7000}, {5, 5300}, {9, 2500}, {6, 1700}},
		},
		{
			name: "limit/offset keep totals", query: "limit=2&offset=1", wantIDs: []int64{6, 5}, wantTotal: 7,
			wantTotals:    EntryTotals{9500, 20000, 7000},
			wantBreakdown: []bd{{100, 20000}, {202, 7000}, {5, 5300}, {9, 2500}, {6, 1700}},
		},
		{name: "limit clamped", query: "limit=10000", wantIDs: []int64{7, 6, 5, 4, 3, 2, 1}, wantTotal: 7, wantTotals: EntryTotals{9500, 20000, 7000}, wantBreakdown: []bd{{100, 20000}, {202, 7000}, {5, 5300}, {9, 2500}, {6, 1700}}},
		{name: "offset past end", query: "offset=50", wantIDs: []int64{}, wantTotal: 7, wantTotals: EntryTotals{9500, 20000, 7000}, wantBreakdown: []bd{{100, 20000}, {202, 7000}, {5, 5300}, {9, 2500}, {6, 1700}}},
		{
			name: "date range inclusive", query: "from=2026-02-01&to=2026-02-28", wantIDs: []int64{4, 3, 2}, wantTotal: 3,
			wantTotals: EntryTotals{5300, 20000, 0}, wantBreakdown: []bd{{100, 20000}, {5, 5300}},
		},
		{
			name: "type", query: "type=expense", wantIDs: []int64{7, 5, 3, 2, 1}, wantTotal: 5,
			wantTotals: EntryTotals{9500, 0, 0}, wantBreakdown: []bd{{5, 5300}, {9, 2500}, {6, 1700}},
		},
		{
			name: "category parent matches children", query: "category_id=5", wantIDs: []int64{3, 2}, wantTotal: 2,
			wantTotals: EntryTotals{5300, 0, 0}, wantBreakdown: []bd{{5, 5300}},
		},
		{
			name: "category leaf", query: "category_id=51", wantIDs: []int64{2}, wantTotal: 1,
			wantTotals: EntryTotals{5000, 0, 0}, wantBreakdown: []bd{{5, 5000}},
		},
		{
			name: "payer", query: "payer_id=3", wantIDs: []int64{6, 2}, wantTotal: 2,
			wantTotals: EntryTotals{5000, 0, 7000}, wantBreakdown: []bd{{202, 7000}, {5, 5000}},
		},
		{
			name: "tag (case-insensitive input)", query: "tag=%20TRIP%20", wantIDs: []int64{7, 5}, wantTotal: 2,
			wantTotals: EntryTotals{3200, 0, 0}, wantBreakdown: []bd{{9, 2500}, {6, 700}},
		},
		{name: "tag exact only", query: "tag=tri", wantIDs: []int64{}, wantTotal: 0, wantBreakdown: []bd{}},
		{
			name: "q note case-insensitive", query: "q=pingo", wantIDs: []int64{1}, wantTotal: 1,
			wantTotals: EntryTotals{1000, 0, 0}, wantBreakdown: []bd{{6, 1000}},
		},
		{
			name: "q category name", query: "q=GROCER", wantIDs: []int64{7, 1}, wantTotal: 2,
			wantTotals: EntryTotals{1700, 0, 0}, wantBreakdown: []bd{{6, 1700}},
		},
		{
			name: "q parent category name", query: "q=transport", wantIDs: []int64{5}, wantTotal: 1,
			wantTotals: EntryTotals{2500, 0, 0}, wantBreakdown: []bd{{9, 2500}},
		},
		{
			name: "q tag substring", query: "q=ork", wantIDs: []int64{4}, wantTotal: 1,
			wantTotals: EntryTotals{0, 20000, 0}, wantBreakdown: []bd{{100, 20000}},
		},
		{
			name: "q escapes % and _", query: "q=" + url.QueryEscape("0%_"), wantIDs: []int64{3}, wantTotal: 1,
			wantTotals: EntryTotals{300, 0, 0}, wantBreakdown: []bd{{5, 300}},
		},
		{name: "q literal % not wildcard", query: "q=" + url.QueryEscape("%"), wantIDs: []int64{3}, wantTotal: 1, wantTotals: EntryTotals{300, 0, 0}, wantBreakdown: []bd{{5, 300}}},
		{name: "q literal _ not wildcard", query: "q=o_", wantIDs: []int64{}, wantTotal: 0, wantBreakdown: []bd{}},
		{
			name: "amount range", query: "min_cents=700&max_cents=2500", wantIDs: []int64{7, 5, 1}, wantTotal: 3,
			wantTotals: EntryTotals{4200, 0, 0}, wantBreakdown: []bd{{9, 2500}, {6, 1700}},
		},
		{
			name: "status pending", query: "status=pending", wantIDs: []int64{5}, wantTotal: 1,
			wantTotals: EntryTotals{2500, 0, 0}, wantBreakdown: []bd{{9, 2500}},
		},
		{
			name: "combined", query: "type=expense&from=2026-02-01&payer_id=2&q=air", wantIDs: []int64{5}, wantTotal: 1,
			wantTotals: EntryTotals{2500, 0, 0}, wantBreakdown: []bd{{9, 2500}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := do(t, h, "GET", "/api/entries?"+tt.query, nil)
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d (%s)", rec.Code, rec.Body)
			}
			if strings.Contains(rec.Body.String(), "null,") && strings.Contains(rec.Body.String(), `"entries":null`) {
				t.Fatal("entries is null")
			}
			list := decode[EntryList](t, rec)
			if list.Entries == nil || list.Breakdown == nil {
				t.Fatalf("nil slices: %s", rec.Body)
			}
			gotIDs := make([]int64, 0, len(list.Entries))
			for _, e := range list.Entries {
				gotIDs = append(gotIDs, e.ID)
			}
			if fmt.Sprint(gotIDs) != fmt.Sprint(tt.wantIDs) {
				t.Errorf("ids = %v, want %v", gotIDs, tt.wantIDs)
			}
			if list.TotalCount != tt.wantTotal {
				t.Errorf("total_count = %d, want %d", list.TotalCount, tt.wantTotal)
			}
			if list.Totals != tt.wantTotals {
				t.Errorf("totals = %+v, want %+v", list.Totals, tt.wantTotals)
			}
			gotBD := make([]bd, 0, len(list.Breakdown))
			for _, b := range list.Breakdown {
				gotBD = append(gotBD, bd{b.CategoryID, b.AmountCents})
			}
			if fmt.Sprint(gotBD) != fmt.Sprint(tt.wantBreakdown) {
				t.Errorf("breakdown = %v, want %v", gotBD, tt.wantBreakdown)
			}
		})
	}

	t.Run("breakdown carries top-level category metadata", func(t *testing.T) {
		list := decode[EntryList](t, do(t, h, "GET", "/api/entries?category_id=56", nil))
		b := list.Breakdown[0]
		if b.CategoryID != 9 || b.Name != "Transport" || b.Type != "expense" || b.Color != "#14b8a6" || b.Icon != "bus" {
			t.Fatalf("breakdown = %+v", b)
		}
		e := list.Entries[0]
		if e.Status != "pending" || e.CategoryName != "Taxi/Uber" || *e.ParentCategoryName != "Transport" || e.Tags[0] != "trip" {
			t.Fatalf("entry = %+v", e)
		}
	})

	t.Run("default limit is 50", func(t *testing.T) {
		for range 50 {
			coreCreateEntry(t, h, coreEntryBody(coreWith("date", "2025-01-01"), coreWith("amount_cents", 1)))
		}
		list := decode[EntryList](t, do(t, h, "GET", "/api/entries", nil))
		if len(list.Entries) != 50 || list.TotalCount != 57 || list.Totals.ExpenseCents != 9550 {
			t.Fatalf("len=%d total=%d totals=%+v", len(list.Entries), list.TotalCount, list.Totals)
		}
	})
}

func TestListEntriesValidation(t *testing.T) {
	_, h := newTestServer(t)
	for _, q := range []string{
		"from=2026-13-01", "to=yesterday", "type=bogus", "status=done", "category_id=x", "payer_id=-1",
		"min_cents=1.5", "max_cents=-3", "limit=0", "limit=abc", "offset=-1",
	} {
		rec := do(t, h, "GET", "/api/entries?"+q, nil)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("%s: status = %d", q, rec.Code)
			continue
		}
		if e := decode[errorBody](t, rec); e.Error == "" {
			t.Errorf("%s: missing error", q)
		}
	}
}

func TestListEntriesEmpty(t *testing.T) {
	_, h := newTestServer(t)
	rec := do(t, h, "GET", "/api/entries", nil)
	body := rec.Body.String()
	if !strings.Contains(body, `"entries":[]`) || !strings.Contains(body, `"breakdown":[]`) {
		t.Fatalf("body = %s", body)
	}
}
