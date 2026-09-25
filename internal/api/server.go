// Package api exposes the JSON HTTP API (see docs/API.md) and serves the SPA.
//
// Route registration is split per domain so features can be developed in
// parallel: routes_core.go (entries, categories, people, tags),
// routes_ops.go (templates, pending, budgets, export, backup) and
// routes_reports.go (reports, insights).
package api

import (
	"database/sql"
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/melojms/mm-budget/internal/clock"
	"github.com/melojms/mm-budget/internal/store"
)

// Version is set at build time via -ldflags.
var Version = "dev"

type Server struct {
	DB        *sql.DB
	Q         *store.Queries
	Clock     clock.Clock
	Log       *slog.Logger
	BackupDir string
	// Static is the built SPA (web/dist). Nil disables SPA serving (tests).
	Static fs.FS
}

func New(db *sql.DB, c clock.Clock, log *slog.Logger, backupDir string, static fs.FS) *Server {
	return &Server{DB: db, Q: store.New(db), Clock: c, Log: log, BackupDir: backupDir, Static: static}
}

// Routes builds the full HTTP handler.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", s.handleHealth)
	mux.HandleFunc("GET /api/meta", s.handleMeta)

	s.routesCore(mux)
	s.routesOps(mux)
	s.routesReports(mux)

	mux.Handle("/api/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotFound, "not found")
	}))
	if s.Static != nil {
		mux.Handle("/", spaHandler(s.Static))
	}
	return s.logRequests(mux)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if err := s.DB.PingContext(r.Context()); err != nil {
		writeError(w, http.StatusServiceUnavailable, "db unavailable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleMeta(w http.ResponseWriter, _ *http.Request) {
	now := s.Clock.Now()
	writeJSON(w, http.StatusOK, map[string]string{
		"today":    now.Format(clock.DateLayout),
		"month":    now.Format(clock.MonthLayout),
		"timezone": now.Location().String(),
		"version":  Version,
	})
}
