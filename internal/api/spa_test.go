package api

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func spaTestFS() fstest.MapFS {
	return fstest.MapFS{
		"index.html":           {Data: []byte("<!doctype html><title>app</title>")},
		"assets/app-1.js":      {Data: []byte(strings.Repeat("console.log('hi');", 200))},
		"assets/font.woff2":    {Data: []byte("wOF2binary")},
		"favicon.svg":          {Data: []byte("<svg/>")},
		"sw.js":                {Data: []byte("self.addEventListener('fetch', () => {})")},
		"manifest.webmanifest": {Data: []byte(`{"name":"Mealheiro"}`)},
	}
}

func spaGet(t *testing.T, h http.Handler, path, acceptEncoding string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if acceptEncoding != "" {
		req.Header.Set("Accept-Encoding", acceptEncoding)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestSPAHandler(t *testing.T) {
	h := spaHandler(spaTestFS())

	tests := []struct {
		name         string
		path         string
		wantStatus   int
		wantBody     string
		wantCache    string
		wantEncoding string
		accept       string
	}{
		{name: "root serves index", path: "/", wantStatus: 200, wantBody: "<title>app</title>", wantCache: "no-cache"},
		{name: "client route falls back to index", path: "/month", wantStatus: 200, wantBody: "<title>app</title>", wantCache: "no-cache"},
		{name: "nested client route", path: "/settings/budgets", wantStatus: 200, wantBody: "<title>app</title>", wantCache: "no-cache"},
		{name: "hashed asset is immutable", path: "/assets/app-1.js", wantStatus: 200, wantBody: "console.log", wantCache: "public, max-age=31536000, immutable"},
		// A tab opened before an upgrade asks for a chunk that no longer exists: it must 404, not get HTML.
		{name: "missing asset is 404", path: "/assets/app-0.js", wantStatus: 404},
		{name: "gzip when accepted", path: "/assets/app-1.js", accept: "br, gzip", wantStatus: 200, wantBody: "console.log", wantEncoding: "gzip", wantCache: "public, max-age=31536000, immutable"},
		{name: "index gzip when accepted", path: "/", accept: "gzip", wantStatus: 200, wantBody: "<title>app</title>", wantEncoding: "gzip", wantCache: "no-cache"},
		{name: "no gzip for already compressed fonts", path: "/assets/font.woff2", accept: "gzip", wantStatus: 200, wantBody: "wOF2binary"},
		{name: "no gzip when not accepted", path: "/assets/app-1.js", accept: "identity", wantStatus: 200, wantBody: "console.log"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := spaGet(t, h, tt.path, tt.accept)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if tt.wantStatus != 200 {
				return
			}
			if got := rec.Header().Get("Content-Encoding"); got != tt.wantEncoding {
				t.Fatalf("Content-Encoding = %q, want %q", got, tt.wantEncoding)
			}
			if tt.wantCache != "" {
				if got := rec.Header().Get("Cache-Control"); got != tt.wantCache {
					t.Errorf("Cache-Control = %q, want %q", got, tt.wantCache)
				}
			}
			if got := rec.Header().Get("Vary"); tt.accept != "" && tt.wantEncoding != "" && got != "Accept-Encoding" {
				t.Errorf("Vary = %q, want Accept-Encoding", got)
			}
			body := rec.Body.Bytes()
			if tt.wantEncoding == "gzip" {
				zr, err := gzip.NewReader(rec.Body)
				if err != nil {
					t.Fatal(err)
				}
				if body, err = io.ReadAll(zr); err != nil {
					t.Fatal(err)
				}
			}
			if !strings.Contains(string(body), tt.wantBody) {
				t.Errorf("body %q does not contain %q", string(body), tt.wantBody)
			}
		})
	}
}

func TestSPAHandlerPWAFiles(t *testing.T) {
	h := spaHandler(spaTestFS())
	for _, accept := range []string{"", "gzip"} {
		t.Run("sw.js accept="+accept, func(t *testing.T) {
			rec := spaGet(t, h, "/sw.js", accept)
			if rec.Code != 200 {
				t.Fatalf("status = %d", rec.Code)
			}
			for k, want := range map[string]string{
				"Content-Type":           "text/javascript",
				"Cache-Control":          "no-cache",
				"Service-Worker-Allowed": "/",
			} {
				if got := rec.Header().Get(k); got != want {
					t.Errorf("%s = %q, want %q", k, got, want)
				}
			}
		})
		t.Run("manifest accept="+accept, func(t *testing.T) {
			rec := spaGet(t, h, "/manifest.webmanifest", accept)
			if rec.Code != 200 {
				t.Fatalf("status = %d", rec.Code)
			}
			if got := rec.Header().Get("Content-Type"); got != "application/manifest+json" {
				t.Errorf("Content-Type = %q, want application/manifest+json", got)
			}
		})
	}
}

func TestSPAHandlerPWAFilesNeverFallBack(t *testing.T) {
	fsys := spaTestFS()
	delete(fsys, "sw.js")
	delete(fsys, "manifest.webmanifest")
	h := spaHandler(fsys)
	for _, p := range []string{"/sw.js", "/manifest.webmanifest"} {
		if rec := spaGet(t, h, p, "gzip"); rec.Code != http.StatusNotFound {
			t.Errorf("%s: status = %d, want 404 (not index.html)", p, rec.Code)
		}
	}
	// Other unknown paths still get the SPA.
	if rec := spaGet(t, h, "/offline", ""); rec.Code != 200 || !strings.Contains(rec.Body.String(), "<title>app</title>") {
		t.Errorf("/offline: status %d body %q, want index.html", rec.Code, rec.Body.String())
	}
}
