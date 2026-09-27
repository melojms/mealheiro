package api

import (
	"net/http"
	"strconv"
	"testing"
)

// opsInsertEntry adds an entry directly (entries endpoints belong to another domain).
func opsInsertEntry(t *testing.T, s *Server, typ, date string, cents, categoryID int64, status string) int64 {
	t.Helper()
	res, err := s.DB.ExecContext(t.Context(),
		`INSERT INTO entries (type, date, amount_cents, category_id, payer_id, status) VALUES (?, ?, ?, ?, 1, ?)`,
		typ, date, cents, categoryID, status)
	if err != nil {
		t.Fatal(err)
	}
	id, _ := res.LastInsertId()
	return id
}

func templateBody() map[string]any {
	return map[string]any{
		"type": "expense", "category_id": 51, "payer_id": 3, "amount_cents": 80000,
		"variable": false, "note": " mortgage ", "start_month": "2026-01",
	}
}

func opsWith(base map[string]any, kv ...any) map[string]any {
	out := map[string]any{}
	for k, v := range base {
		out[k] = v
	}
	for i := 0; i < len(kv); i += 2 {
		if kv[i+1] == nil {
			delete(out, kv[i].(string))
			continue
		}
		out[kv[i].(string)] = kv[i+1]
	}
	return out
}

func TestCreateTemplateGenerates(t *testing.T) {
	_, h := newTestServer(t)
	rec := do(t, h, "POST", "/api/templates", templateBody())
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d %s", rec.Code, rec.Body)
	}
	tpl := decode[Template](t, rec)
	if tpl.CategoryName != "Mortgage" || tpl.ParentCategoryName == nil || *tpl.ParentCategoryName != "House" ||
		tpl.PayerName != "Joint" || tpl.Note != "mortgage" || !tpl.Active || tpl.EndMonth != nil {
		t.Fatalf("template = %+v", tpl)
	}
	if tpl.LastGeneratedMonth == nil || *tpl.LastGeneratedMonth != "2026-03" {
		t.Fatalf("last_generated_month = %v", tpl.LastGeneratedMonth)
	}

	rec = do(t, h, "POST", "/api/recurring/run", nil)
	if got := decode[map[string]int](t, rec); rec.Code != http.StatusOK || got["created"] != 0 {
		t.Fatalf("run = %d %v", rec.Code, got)
	}
}

func TestCreateTemplateValidation(t *testing.T) {
	_, h := newTestServer(t)
	tests := []struct {
		name string
		body any
		want int
	}{
		{"bad type", opsWith(templateBody(), "type", "gift"), 400},
		{"zero amount", opsWith(templateBody(), "amount_cents", 0), 400},
		{"negative amount", opsWith(templateBody(), "amount_cents", -5), 400},
		{"bad start", opsWith(templateBody(), "start_month", "2026-1"), 400},
		{"missing start", opsWith(templateBody(), "start_month", nil), 400},
		{"bad end", opsWith(templateBody(), "end_month", "2026-13"), 400},
		{"end before start", opsWith(templateBody(), "end_month", "2025-12"), 400},
		{"unknown category", opsWith(templateBody(), "category_id", 9999), 400},
		{"type mismatch", opsWith(templateBody(), "category_id", 100), 400},
		{"unknown payer", opsWith(templateBody(), "payer_id", 42), 400},
		{"unknown field", opsWith(templateBody(), "foo", 1), 400},
		{"not json", "nope", 400},
		{"empty end is null", opsWith(templateBody(), "end_month", ""), 201},
		{"end equals start", opsWith(templateBody(), "end_month", "2026-01"), 201},
		{"inactive", opsWith(templateBody(), "active", false), 201},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := do(t, h, "POST", "/api/templates", tt.body)
			if rec.Code != tt.want {
				t.Fatalf("status = %d %s", rec.Code, rec.Body)
			}
			if tt.want == 400 && decode[map[string]string](t, rec)["error"] == "" {
				t.Fatal("missing error message")
			}
		})
	}
}

func TestTemplateArchivedCategory(t *testing.T) {
	s, h := newTestServer(t)
	rec := do(t, h, "POST", "/api/templates", opsWith(templateBody(), "category_id", 7))
	if rec.Code != http.StatusCreated {
		t.Fatalf("create = %d %s", rec.Code, rec.Body)
	}
	id := decode[Template](t, rec).ID
	if _, err := s.DB.ExecContext(t.Context(), `UPDATE categories SET archived = TRUE WHERE id IN (7, 8)`); err != nil {
		t.Fatal(err)
	}
	// New templates cannot use an archived category...
	if rec := do(t, h, "POST", "/api/templates", opsWith(templateBody(), "category_id", 8)); rec.Code != 400 {
		t.Fatalf("create archived = %d", rec.Code)
	}
	// ...but an existing one can still be edited (e.g. paused) while keeping it.
	if rec := do(t, h, "PUT", "/api/templates/"+opsItoa(id), opsWith(templateBody(), "category_id", 7, "active", false)); rec.Code != 200 {
		t.Fatalf("pause = %d %s", rec.Code, rec.Body)
	}
	if rec := do(t, h, "PUT", "/api/templates/"+opsItoa(id), opsWith(templateBody(), "category_id", 8)); rec.Code != 400 {
		t.Fatalf("switch to archived = %d", rec.Code)
	}
}

func TestUpdateTemplateAffectsFutureOnly(t *testing.T) {
	s, h := newTestServer(t)
	id := decode[Template](t, do(t, h, "POST", "/api/templates", templateBody())).ID

	rec := do(t, h, "PUT", "/api/templates/"+opsItoa(id), opsWith(templateBody(), "amount_cents", 90000, "variable", true))
	if rec.Code != http.StatusOK {
		t.Fatalf("update = %d %s", rec.Code, rec.Body)
	}
	if tpl := decode[Template](t, rec); tpl.AmountCents != 90000 || !tpl.Variable {
		t.Fatalf("template = %+v", tpl)
	}
	var n int
	if err := s.DB.QueryRowContext(t.Context(), `SELECT count(*) FROM entries WHERE template_id = ? AND amount_cents = 80000 AND status = 'confirmed'`, id).Scan(&n); err != nil || n != 3 {
		t.Fatalf("existing entries changed: %d %v", n, err)
	}

	// Moving start_month earlier back-fills the missing month.
	rec = do(t, h, "PUT", "/api/templates/"+opsItoa(id), opsWith(templateBody(), "start_month", "2025-12", "variable", true))
	if rec.Code != http.StatusOK {
		t.Fatalf("update = %d", rec.Code)
	}
	var status string
	var amount int64
	if err := s.DB.QueryRowContext(t.Context(), `SELECT status, amount_cents FROM entries WHERE template_id = ? AND template_month = '2025-12'`, id).Scan(&status, &amount); err != nil {
		t.Fatal(err)
	}
	if status != "pending" || amount != 80000 {
		t.Fatalf("backfill = %s %d", status, amount)
	}
}

func TestUpdateDeleteTemplateErrors(t *testing.T) {
	_, h := newTestServer(t)
	tests := []struct {
		method, path string
		body         any
		want         int
	}{
		{"PUT", "/api/templates/999", templateBody(), 404},
		{"PUT", "/api/templates/abc", templateBody(), 400},
		{"PUT", "/api/templates/0", templateBody(), 400},
		{"DELETE", "/api/templates/999", nil, 404},
		{"DELETE", "/api/templates/x", nil, 400},
	}
	for _, tt := range tests {
		if rec := do(t, h, tt.method, tt.path, tt.body); rec.Code != tt.want {
			t.Errorf("%s %s = %d, want %d", tt.method, tt.path, rec.Code, tt.want)
		}
	}
	id := decode[Template](t, do(t, h, "POST", "/api/templates", templateBody())).ID
	if rec := do(t, h, "PUT", "/api/templates/"+opsItoa(id), opsWith(templateBody(), "amount_cents", 0)); rec.Code != 400 {
		t.Errorf("invalid update = %d", rec.Code)
	}
}

func TestDeleteTemplateKeepsEntries(t *testing.T) {
	s, h := newTestServer(t)
	id := decode[Template](t, do(t, h, "POST", "/api/templates", opsWith(templateBody(), "variable", true))).ID
	if rec := do(t, h, "DELETE", "/api/templates/"+opsItoa(id), nil); rec.Code != http.StatusNoContent {
		t.Fatalf("delete = %d", rec.Code)
	}
	pending := decode[[]PendingEntry](t, do(t, h, "GET", "/api/pending", nil))
	if len(pending) != 3 {
		t.Fatalf("pending = %d", len(pending))
	}
	for _, p := range pending {
		if p.TemplateID != nil || !p.Recurring {
			t.Fatalf("entry = %+v", p)
		}
	}
	var runs int
	_ = s.DB.QueryRowContext(t.Context(), `SELECT count(*) FROM template_runs`).Scan(&runs)
	if runs != 0 {
		t.Fatalf("runs = %d", runs)
	}
	if list := decode[[]Template](t, do(t, h, "GET", "/api/templates", nil)); len(list) != 0 {
		t.Fatalf("templates = %+v", list)
	}
}

func TestListTemplatesOrder(t *testing.T) {
	_, h := newTestServer(t)
	for _, body := range []map[string]any{
		opsWith(templateBody(), "category_id", 7, "active", false),          // Subscriptions, inactive
		opsWith(templateBody(), "type", "income", "category_id", 100),       // Salary
		opsWith(templateBody(), "category_id", 2, "start_month", "2026-04"), // Electricity
		opsWith(templateBody(), "category_id", 51),                          // Mortgage
		opsWith(templateBody(), "type", "investment", "category_id", 200),   // ETFs/Stocks
	} {
		if rec := do(t, h, "POST", "/api/templates", body); rec.Code != http.StatusCreated {
			t.Fatalf("create = %d %s", rec.Code, rec.Body)
		}
	}
	rec := do(t, h, "GET", "/api/templates", nil)
	list := decode[[]Template](t, rec)
	var names []string
	for _, tpl := range list {
		names = append(names, tpl.CategoryName)
	}
	want := []string{"Electricity", "Mortgage", "Salary", "ETFs/Stocks", "Subscriptions"}
	if len(names) != len(want) {
		t.Fatalf("names = %v", names)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Fatalf("names = %v, want %v", names, want)
		}
	}
	if list[0].LastGeneratedMonth != nil || list[4].LastGeneratedMonth != nil || list[0].ParentCategoryName != nil {
		t.Fatalf("future/inactive/top-level fields: %+v %+v", list[0], list[4])
	}
}

func TestListTemplatesEmpty(t *testing.T) {
	_, h := newTestServer(t)
	rec := do(t, h, "GET", "/api/templates", nil)
	if rec.Code != 200 || rec.Body.String() != "[]\n" {
		t.Fatalf("got %d %q", rec.Code, rec.Body)
	}
	if rec := do(t, h, "GET", "/api/pending", nil); rec.Body.String() != "[]\n" {
		t.Fatalf("pending = %q", rec.Body)
	}
}

func TestPending(t *testing.T) {
	s, h := newTestServer(t)
	opsInsertEntry(t, s, "expense", "2026-03-10", 100, 6, "pending")
	opsInsertEntry(t, s, "expense", "2026-02-28", 200, 6, "pending")
	opsInsertEntry(t, s, "expense", "2026-01-05", 300, 6, "confirmed")
	opsInsertEntry(t, s, "income", "2026-04-01", 400, 100, "pending")

	list := decode[[]PendingEntry](t, do(t, h, "GET", "/api/pending", nil))
	type got struct {
		date  string
		stale bool
	}
	want := []got{{"2026-02-28", true}, {"2026-03-10", false}, {"2026-04-01", false}}
	if len(list) != len(want) {
		t.Fatalf("list = %+v", list)
	}
	for i, p := range list {
		if (got{p.Date, p.Stale}) != want[i] || p.Status != "pending" {
			t.Errorf("item %d = %+v", i, p)
		}
	}
	// JSON carries the Entry fields flattened plus stale.
	raw := decode[[]map[string]any](t, do(t, h, "GET", "/api/pending", nil))
	if _, ok := raw[0]["category_name"]; !ok || raw[0]["stale"] != true {
		t.Fatalf("raw = %v", raw[0])
	}
}

func opsItoa(n int64) string { return strconv.FormatInt(n, 10) }

func TestTemplatePersonalFlag(t *testing.T) {
	tests := []struct {
		name string
		body map[string]any
		want bool
	}{
		{name: "defaults to shared", body: opsWith(templateBody(), "payer_id", 1), want: false},
		{name: "personal expense kept", body: opsWith(templateBody(), "payer_id", 1, "personal", true), want: true},
		{name: "cleared when joint pays", body: opsWith(templateBody(), "personal", true), want: false},
		{name: "cleared on income", body: opsWith(templateBody(), "payer_id", 1, "personal", true, "type", "income", "category_id", 100), want: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, h := newTestServer(t)
			rec := do(t, h, "POST", "/api/templates", tc.body)
			if rec.Code != http.StatusCreated {
				t.Fatalf("create = %d %s", rec.Code, rec.Body)
			}
			tpl := decode[Template](t, rec)
			if tpl.Personal != tc.want {
				t.Fatalf("create personal = %v, want %v", tpl.Personal, tc.want)
			}
			rec = do(t, h, "PUT", "/api/templates/"+opsItoa(tpl.ID), tc.body)
			if rec.Code != http.StatusOK {
				t.Fatalf("update = %d %s", rec.Code, rec.Body)
			}
			if got := decode[Template](t, rec).Personal; got != tc.want {
				t.Fatalf("update personal = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestUpdateTemplatePersonalKeepsGeneratedEntries(t *testing.T) {
	s, h := newTestServer(t)
	body := opsWith(templateBody(), "payer_id", 1, "personal", true)
	id := decode[Template](t, do(t, h, "POST", "/api/templates", body)).ID

	if rec := do(t, h, "PUT", "/api/templates/"+opsItoa(id), opsWith(body, "personal", false)); rec.Code != http.StatusOK {
		t.Fatalf("update = %d %s", rec.Code, rec.Body)
	}
	var n int
	if err := s.DB.QueryRowContext(t.Context(), `SELECT count(*) FROM entries WHERE template_id = ? AND personal`, id).Scan(&n); err != nil || n != 3 {
		t.Fatalf("personal generated entries = %d %v, want 3", n, err)
	}
}
