package api

import (
	"bytes"
	"compress/gzip"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strconv"
	"strings"
	"sync"
)

// compressible lists the text asset types worth gzipping; fonts and images are already compressed.
var compressible = map[string]bool{
	".html": true, ".js": true, ".css": true, ".svg": true, ".json": true, ".webmanifest": true, ".txt": true,
}

func init() {
	// Chrome's installability check wants the registered manifest type; Go has no default for it.
	_ = mime.AddExtensionType(".webmanifest", "application/manifest+json")
}

// noFallback lists root files that must 404 when missing instead of getting
// index.html: a service worker or manifest that parses as HTML breaks installs.
var noFallback = map[string]bool{"sw.js": true, "manifest.webmanifest": true}

// spaHandler serves static files from fsys and falls back to index.html for
// client-side routes. Hashed assets are cached aggressively; text files are
// gzipped (once, then kept in memory) for clients that accept it.
func spaHandler(fsys fs.FS) http.Handler {
	files := http.FileServerFS(fsys)
	var gzCache sync.Map // path -> []byte
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if p == "" {
			p = "index.html"
		}
		if _, err := fs.Stat(fsys, p); err != nil {
			// A stale tab asking for a chunk from a previous build must get a 404, not HTML,
			// so the client can detect it and reload.
			if strings.HasPrefix(p, "assets/") || noFallback[p] {
				http.NotFound(w, r)
				return
			}
			r.URL.Path = "/"
			p = "index.html"
		}
		if strings.HasPrefix(p, "assets/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			w.Header().Set("Cache-Control", "no-cache")
		}
		if p == "sw.js" {
			w.Header().Set("Content-Type", "text/javascript")
			w.Header().Set("Service-Worker-Allowed", "/")
		}

		ext := path.Ext(p)
		if !compressible[ext] {
			files.ServeHTTP(w, r)
			return
		}
		w.Header().Add("Vary", "Accept-Encoding")
		if !acceptsGzip(r) || r.Header.Get("Range") != "" {
			files.ServeHTTP(w, r)
			return
		}
		gz, err := gzipped(&gzCache, fsys, p)
		if err != nil {
			files.ServeHTTP(w, r)
			return
		}
		if w.Header().Get("Content-Type") == "" {
			w.Header().Set("Content-Type", mime.TypeByExtension(ext))
		}
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Content-Length", strconv.Itoa(len(gz)))
		if r.Method == http.MethodHead {
			return
		}
		_, _ = w.Write(gz)
	})
}

func acceptsGzip(r *http.Request) bool {
	for part := range strings.SplitSeq(r.Header.Get("Accept-Encoding"), ",") {
		name, q, _ := strings.Cut(strings.TrimSpace(part), ";")
		if strings.EqualFold(strings.TrimSpace(name), "gzip") {
			return strings.ReplaceAll(strings.TrimSpace(q), " ", "") != "q=0"
		}
	}
	return false
}

func gzipped(cache *sync.Map, fsys fs.FS, p string) ([]byte, error) {
	if v, ok := cache.Load(p); ok {
		return v.([]byte), nil
	}
	raw, err := fs.ReadFile(fsys, p)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	zw, _ := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if _, err := zw.Write(raw); err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	v, _ := cache.LoadOrStore(p, buf.Bytes())
	return v.([]byte), nil
}
