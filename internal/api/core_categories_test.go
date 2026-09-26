package api

import (
	"net/http"
	"testing"
)

// coreExec runs raw SQL against the test server's DB.
func coreExec(t *testing.T, s *Server, query string, args ...any) {
	t.Helper()
	if _, err := s.DB.ExecContext(t.Context(), query, args...); err != nil {
		t.Fatalf("exec %q: %v", query, err)
	}
}

func coreFindCategory(cats []Category, id int64) (Category, bool) {
	for _, c := range cats {
		if c.ID == id {
			return c, true
		}
	}
	return Category{}, false
}

func TestListCategories(t *testing.T) {
	s, h := newTestServer(t)
	// Groceries (6): 3 recent; House (5): 1 on self + 2 on Mortgage (51); one old Water entry.
	coreExec(t, s, `INSERT INTO entries (type, date, amount_cents, category_id, payer_id) VALUES
		('expense','2026-03-15',100,6,1), ('expense','2026-01-01',100,6,1), ('expense','2025-12-16',100,6,1),
		('expense','2026-03-01',100,5,1), ('expense','2026-02-01',100,51,3), ('expense','2026-03-01',100,51,3),
		('expense','2025-12-15',100,1,1), ('expense','2026-03-16',100,1,1)`)
	coreExec(t, s, `UPDATE categories SET archived = TRUE WHERE id = 19`)

	rec := do(t, h, "GET", "/api/categories?type=expense", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	cats := decode[[]Category](t, rec)
	if len(cats) != 25 { // 19 top + 7 sub - archived Other
		t.Fatalf("len = %d", len(cats))
	}
	for _, c := range cats {
		if c.Type != "expense" || c.Archived {
			t.Fatalf("unexpected category %+v", c)
		}
	}
	// House: self + children = 3; Groceries = 3 (tie broken by name) ; Mortgage = 2.
	wantOrder := []int64{6, 5, 51}
	for i, id := range wantOrder {
		if cats[i].ID != id {
			t.Fatalf("order[%d] = %d (%s), want %d", i, cats[i].ID, cats[i].Name, id)
		}
	}
	house, _ := coreFindCategory(cats, 5)
	if house.UsageCount != 3 || house.EntryCount != 1 {
		t.Errorf("house = %+v", house)
	}
	mortgage, _ := coreFindCategory(cats, 51)
	if mortgage.UsageCount != 2 || mortgage.EntryCount != 2 || mortgage.ParentID == nil || *mortgage.ParentID != 5 {
		t.Errorf("mortgage = %+v", mortgage)
	}
	// Water entries are just outside the 90-day window (and one in the future).
	water, _ := coreFindCategory(cats, 1)
	if water.UsageCount != 0 || water.EntryCount != 2 {
		t.Errorf("water = %+v", water)
	}
	// Remaining zero-usage categories are sorted by name.
	for i := 4; i < len(cats); i++ {
		if cats[i].UsageCount == 0 && cats[i-1].UsageCount == 0 && cats[i-1].Name > cats[i].Name {
			t.Errorf("not name-sorted: %q before %q", cats[i-1].Name, cats[i].Name)
		}
	}

	all := decode[[]Category](t, do(t, h, "GET", "/api/categories?include_archived=true", nil))
	if len(all) != 33 {
		t.Fatalf("all len = %d", len(all))
	}
	if other, ok := coreFindCategory(all, 19); !ok || !other.Archived {
		t.Errorf("other = %+v", other)
	}

	for _, bad := range []string{"?type=foo", "?include_archived=maybe"} {
		if rec := do(t, h, "GET", "/api/categories"+bad, nil); rec.Code != http.StatusBadRequest {
			t.Errorf("%s: status = %d", bad, rec.Code)
		}
	}
}

func TestCreateCategory(t *testing.T) {
	tests := []struct {
		name      string
		setup     string
		body      map[string]any
		wantCode  int
		wantIcon  string
		wantColor string
	}{
		{"top-level defaults (last unused palette color)", "", map[string]any{"type": "expense", "name": "Pets"}, 201, "circle", "#78716c"},
		{"first unused palette color (income)", "", map[string]any{"type": "income", "name": "Gifts"}, 201, "circle", "#0ea5e9"},
		{"explicit icon/color normalized", "", map[string]any{"type": "expense", "name": "Pets", "icon": "dog", "color": "#AABBCC"}, 201, "dog", "#aabbcc"},
		{"child inherits parent", "", map[string]any{"type": "expense", "name": "Bus", "parent_id": 9}, 201, "bus", "#14b8a6"},
		{"child explicit", "", map[string]any{"type": "expense", "name": "Bike", "parent_id": 9, "icon": "bike"}, 201, "bike", "#14b8a6"},
		{"null parent", "", map[string]any{"type": "investment", "name": "Gold", "parent_id": nil}, 201, "circle", "#eab308"},
		{"all palette used cycles", "INSERT INTO categories (type, name, icon, color) VALUES ('expense','Pets','dog','#78716c')", map[string]any{"type": "expense", "name": "Kids"}, 201, "circle", "#0ea5e9"},
		{"duplicate top-level (case-insensitive)", "", map[string]any{"type": "expense", "name": "groceries"}, 409, "", ""},
		{"duplicate child", "", map[string]any{"type": "expense", "name": "Rent", "parent_id": 5}, 409, "", ""},
		{"same name other type ok", "", map[string]any{"type": "income", "name": "Groceries"}, 201, "circle", "#0ea5e9"},
		{"parent is child", "", map[string]any{"type": "expense", "name": "X", "parent_id": 50}, 400, "", ""},
		{"parent archived", "UPDATE categories SET archived = TRUE WHERE id = 9", map[string]any{"type": "expense", "name": "X", "parent_id": 9}, 400, "", ""},
		{"parent other type", "", map[string]any{"type": "income", "name": "X", "parent_id": 9}, 400, "", ""},
		{"parent missing", "", map[string]any{"type": "expense", "name": "X", "parent_id": 999}, 400, "", ""},
		{"bad type", "", map[string]any{"type": "bogus", "name": "X"}, 400, "", ""},
		{"empty name", "", map[string]any{"type": "expense", "name": " "}, 400, "", ""},
		{"bad icon", "", map[string]any{"type": "expense", "name": "X", "icon": "Not An Icon"}, 400, "", ""},
		{"bad color", "", map[string]any{"type": "expense", "name": "X", "color": "red"}, 400, "", ""},
		{"unknown field", "", map[string]any{"type": "expense", "name": "X", "archived": true}, 400, "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, h := newTestServer(t)
			if tt.setup != "" {
				coreExec(t, s, tt.setup)
			}
			rec := do(t, h, "POST", "/api/categories", tt.body)
			if rec.Code != tt.wantCode {
				t.Fatalf("status = %d, want %d (%s)", rec.Code, tt.wantCode, rec.Body)
			}
			if tt.wantCode != http.StatusCreated {
				return
			}
			c := decode[Category](t, rec)
			if c.ID == 0 || c.Icon != tt.wantIcon || c.Color != tt.wantColor || c.Archived || c.UsageCount != 0 || c.EntryCount != 0 {
				t.Fatalf("category = %+v", c)
			}
		})
	}
}

func TestPickColorCyclesWhenFull(t *testing.T) {
	if got := corePickColor(palette); got != palette[0] {
		t.Fatalf("got %s", got)
	}
	if got := corePickColor([]string{"#0EA5E9"}); got != palette[1] {
		t.Fatalf("case-insensitive: got %s", got)
	}
}

func TestUpdateCategory(t *testing.T) {
	type state struct {
		id       int64
		archived bool
	}
	tests := []struct {
		name     string
		setup    string
		id       string
		body     map[string]any
		wantCode int
		check    func(t *testing.T, c Category)
		want     []state
	}{
		{
			name: "rename + icon + color", id: "6",
			body:     map[string]any{"name": " Supermarket ", "icon": "store", "color": "#123ABC"},
			wantCode: 200,
			check: func(t *testing.T, c Category) {
				if c.Name != "Supermarket" || c.Icon != "store" || c.Color != "#123abc" || c.Archived {
					t.Fatalf("got %+v", c)
				}
			},
		},
		{
			name: "empty patch is no-op", id: "6", body: map[string]any{}, wantCode: 200,
			check: func(t *testing.T, c Category) {
				if c.Name != "Groceries" {
					t.Fatalf("got %+v", c)
				}
			},
		},
		{
			name: "archive parent archives children", id: "5",
			body: map[string]any{"archived": true}, wantCode: 200,
			want: []state{{5, true}, {50, true}, {54, true}, {9, false}, {55, false}},
		},
		{
			name: "archive child leaves parent", id: "50",
			body: map[string]any{"archived": true}, wantCode: 200,
			want: []state{{5, false}, {50, true}, {51, false}},
		},
		{
			name:  "unarchive child unarchives parent only",
			setup: "UPDATE categories SET archived = TRUE WHERE id = 5 OR parent_id = 5", id: "51",
			body: map[string]any{"archived": false}, wantCode: 200,
			want: []state{{5, false}, {51, false}, {50, true}},
		},
		{
			name:  "unarchive parent leaves children archived",
			setup: "UPDATE categories SET archived = TRUE WHERE id = 5 OR parent_id = 5", id: "5",
			body: map[string]any{"archived": false}, wantCode: 200,
			want: []state{{5, false}, {50, true}},
		},
		{name: "rename to sibling name", id: "50", body: map[string]any{"name": "mortgage"}, wantCode: 409},
		{name: "rename to top-level name", id: "6", body: map[string]any{"name": "Water"}, wantCode: 409},
		{name: "rename case only ok", id: "6", body: map[string]any{"name": "GROCERIES"}, wantCode: 200},
		{name: "empty name", id: "6", body: map[string]any{"name": ""}, wantCode: 400},
		{name: "bad color", id: "6", body: map[string]any{"color": "#12345"}, wantCode: 400},
		{name: "bad icon", id: "6", body: map[string]any{"icon": "-x"}, wantCode: 400},
		{name: "unknown field", id: "6", body: map[string]any{"parent_id": 5}, wantCode: 400},
		{name: "not found", id: "999", body: map[string]any{"name": "x"}, wantCode: 404},
		{name: "bad id", id: "x", body: map[string]any{"name": "x"}, wantCode: 404},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, h := newTestServer(t)
			if tt.setup != "" {
				coreExec(t, s, tt.setup)
			}
			rec := do(t, h, "PATCH", "/api/categories/"+tt.id, tt.body)
			if rec.Code != tt.wantCode {
				t.Fatalf("status = %d, want %d (%s)", rec.Code, tt.wantCode, rec.Body)
			}
			if tt.check != nil {
				tt.check(t, decode[Category](t, rec))
			}
			if len(tt.want) > 0 {
				all := decode[[]Category](t, do(t, h, "GET", "/api/categories?include_archived=true", nil))
				for _, w := range tt.want {
					if c, _ := coreFindCategory(all, w.id); c.Archived != w.archived {
						t.Errorf("category %d archived = %v, want %v", w.id, c.Archived, w.archived)
					}
				}
			}
		})
	}
}

func TestDeleteCategory(t *testing.T) {
	tests := []struct {
		name     string
		setup    string
		id       string
		wantCode int
	}{
		{"unused leaf", "", "56", 204},
		{"unused top-level", "", "19", 204},
		{"has children", "", "5", 409},
		{"has entry", "INSERT INTO entries (type, date, amount_cents, category_id, payer_id) VALUES ('expense','2020-01-01',1,19,1)", "19", 409},
		{"has template", "INSERT INTO templates (type, category_id, payer_id, amount_cents, start_month) VALUES ('expense',19,1,1,'2026-01')", "19", 409},
		{"has budget", "INSERT INTO budgets (category_id, effective_from, amount_cents) VALUES (19,'2026-01',100)", "19", 409},
		{"not found", "", "999", 404},
		{"bad id", "", "0", 404},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, h := newTestServer(t)
			if tt.setup != "" {
				coreExec(t, s, tt.setup)
			}
			rec := do(t, h, "DELETE", "/api/categories/"+tt.id, nil)
			if rec.Code != tt.wantCode {
				t.Fatalf("status = %d, want %d (%s)", rec.Code, tt.wantCode, rec.Body)
			}
			if tt.wantCode == http.StatusConflict {
				if e := decode[errorBody](t, rec); e.Error == "" {
					t.Fatal("missing error")
				}
			}
			if tt.wantCode == http.StatusNoContent {
				all := decode[[]Category](t, do(t, h, "GET", "/api/categories?include_archived=true", nil))
				if _, ok := coreFindCategory(all, 19); ok && tt.id == "19" {
					t.Fatal("category still present")
				}
			}
		})
	}
}
