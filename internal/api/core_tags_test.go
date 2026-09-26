package api

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestNormalizeTags(t *testing.T) {
	tests := []struct {
		in      []string
		want    string
		wantErr bool
	}{
		{nil, "", false},
		{[]string{"", "  "}, "", false},
		{[]string{" B ", "a", "b", "A"}, "a,b", false},
		{[]string{"Férias"}, "férias", false},
		{[]string{strings.Repeat("é", 40)}, strings.Repeat("é", 40), false},
		{[]string{strings.Repeat("x", 41)}, "", true},
	}
	for _, tt := range tests {
		got, err := coreNormalizeTags(tt.in)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q: err = %v", tt.in, err)
			continue
		}
		if !tt.wantErr && strings.Join(got, ",") != tt.want {
			t.Errorf("%q: got %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestListTags(t *testing.T) {
	s, h := newTestServer(t)
	coreCreateEntry(t, h, coreEntryBody(coreWith("tags", []string{"trip", "food"})))
	coreCreateEntry(t, h, coreEntryBody(coreWith("tags", []string{"trip", "tripod"})))
	coreCreateEntry(t, h, coreEntryBody(coreWith("tags", []string{"trip", "100%"})))
	coreCreateEntry(t, h, coreEntryBody(coreWith("tags", []string{"1000"})))
	// Unused tag must not be listed.
	coreExec(t, s, `INSERT INTO tags (name) VALUES ('unused')`)

	tests := []struct {
		q    string
		want string
	}{
		{"", "trip:3,100%:1,1000:1,food:1,tripod:1"},
		{"tr", "trip:3,tripod:1"},
		{" TRI ", "trip:3,tripod:1"},
		{"rip", ""},
		{"unused", ""},
		{"100%", "100%:1"},
		{"10_", ""},
	}
	for _, tt := range tests {
		t.Run(tt.q, func(t *testing.T) {
			rec := do(t, h, "GET", "/api/tags?q="+url.QueryEscape(tt.q), nil)
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d", rec.Code)
			}
			tags := decode[[]TagCount](t, rec)
			parts := make([]string, 0, len(tags))
			for _, tg := range tags {
				parts = append(parts, fmt.Sprintf("%s:%d", tg.Name, tg.Count))
			}
			if got := strings.Join(parts, ","); got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}

	t.Run("empty list is []", func(t *testing.T) {
		rec := do(t, h, "GET", "/api/tags?q=zzz", nil)
		if strings.TrimSpace(rec.Body.String()) != "[]" {
			t.Fatalf("body = %s", rec.Body)
		}
	})

	t.Run("max 20", func(t *testing.T) {
		names := make([]string, 25)
		for i := range names {
			names[i] = fmt.Sprintf("bulk%02d", i)
		}
		coreCreateEntry(t, h, coreEntryBody(coreWith("tags", names)))
		if tags := decode[[]TagCount](t, do(t, h, "GET", "/api/tags?q=bulk", nil)); len(tags) != 20 {
			t.Fatalf("len = %d", len(tags))
		}
	})

	t.Run("update replaces entry tags", func(t *testing.T) {
		e := coreCreateEntry(t, h, coreEntryBody(coreWith("tags", []string{"solo"})))
		rec := do(t, h, "PUT", fmt.Sprintf("/api/entries/%d", e.ID), coreEntryBody(coreWith("tags", []string{"other"})))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d", rec.Code)
		}
		if tags := decode[[]TagCount](t, do(t, h, "GET", "/api/tags?q=solo", nil)); len(tags) != 0 {
			t.Fatalf("solo still listed: %+v", tags)
		}
	})
}
