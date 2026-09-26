package api

import (
	"net/http"
	"strings"
	"testing"
)

func TestListPeople(t *testing.T) {
	_, h := newTestServer(t)
	rec := do(t, h, "GET", "/api/people", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	people := decode[[]Person](t, rec)
	want := []Person{{1, "Person A", "person"}, {2, "Person B", "person"}, {3, "Joint", "joint"}}
	if len(people) != len(want) {
		t.Fatalf("people = %+v", people)
	}
	for i := range want {
		if people[i] != want[i] {
			t.Errorf("people[%d] = %+v, want %+v", i, people[i], want[i])
		}
	}
}

func TestRenamePerson(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		body     any
		wantCode int
		wantName string
	}{
		{"ok trims", "/api/people/1", map[string]any{"name": "  Ana  "}, 200, "Ana"},
		{"unicode 40 chars", "/api/people/2", map[string]any{"name": strings.Repeat("é", 40)}, 200, strings.Repeat("é", 40)},
		{"too long", "/api/people/1", map[string]any{"name": strings.Repeat("a", 41)}, 400, ""},
		{"empty", "/api/people/1", map[string]any{"name": "   "}, 400, ""},
		{"missing name", "/api/people/1", map[string]any{}, 400, ""},
		{"unknown field", "/api/people/1", map[string]any{"name": "x", "kind": "joint"}, 400, ""},
		{"not found", "/api/people/99", map[string]any{"name": "x"}, 404, ""},
		{"bad id", "/api/people/abc", map[string]any{"name": "x"}, 404, ""},
		{"no body", "/api/people/1", nil, 400, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, h := newTestServer(t)
			rec := do(t, h, "PATCH", tt.path, tt.body)
			if rec.Code != tt.wantCode {
				t.Fatalf("status = %d, want %d (%s)", rec.Code, tt.wantCode, rec.Body)
			}
			if tt.wantCode != 200 {
				if e := decode[errorBody](t, rec); e.Error == "" {
					t.Fatal("missing error message")
				}
				return
			}
			if p := decode[Person](t, rec); p.Name != tt.wantName {
				t.Fatalf("name = %q, want %q", p.Name, tt.wantName)
			}
		})
	}
}
