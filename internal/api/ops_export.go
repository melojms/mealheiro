package api

import (
	"bytes"
	"cmp"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/melojms/mealheiro/internal/backup"
	"github.com/melojms/mealheiro/internal/clock"
	"github.com/melojms/mealheiro/internal/export"
	"github.com/melojms/mealheiro/internal/store"
)

// Bounds used when the export range is open-ended (dates compare as strings).
const (
	exportMinDate = "0000-01-01"
	exportMaxDate = "9999-12-31"
)

func (s *Server) handleExportCSV(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	from, to := q.Get("from"), q.Get("to")
	for _, d := range []string{from, to} {
		if d == "" {
			continue
		}
		if _, err := clock.ParseDate(d); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	if from != "" && to != "" && from > to {
		writeError(w, http.StatusBadRequest, "from must be <= to")
		return
	}
	types := map[string]bool{}
	if raw := q.Get("types"); raw != "" {
		for t := range strings.SplitSeq(raw, ",") {
			t = strings.TrimSpace(t)
			if !knownEntryType(t) {
				writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid type %q", t))
				return
			}
			types[t] = true
		}
	}

	views, err := s.Q.ListExportEntries(r.Context(), store.ListExportEntriesParams{
		FromDate: cmp.Or(from, exportMinDate),
		ToDate:   cmp.Or(to, exportMaxDate),
	})
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	rows := make([]export.Row, 0, len(views))
	for _, v := range views {
		if len(types) > 0 && !types[v.Type] {
			continue
		}
		rows = append(rows, exportRow(v))
	}

	// Render to memory first so a failure can still produce a proper error response.
	var buf bytes.Buffer
	if err := export.Write(&buf, rows); err != nil {
		s.internalError(w, r, err)
		return
	}
	name := fmt.Sprintf("mealheiro_%s_%s.csv", cmp.Or(from, "all"), cmp.Or(to, "all"))
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="`+name+`"`)
	w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
	_, _ = buf.WriteTo(w)
}

func exportRow(v store.EntryView) export.Row {
	row := export.Row{
		Date:        v.Date,
		Type:        v.Type,
		Category:    v.CategoryName,
		AmountCents: v.AmountCents,
		Payer:       v.PayerName,
		Note:        v.Note,
		Recurring:   v.TemplateMonth != nil,
		Status:      v.Status,
		Personal:    v.Personal,
	}
	if v.ParentCategoryID != nil {
		row.Category, row.Subcategory = v.ParentCategoryName, v.CategoryName
	}
	if v.TagsCsv != "" {
		row.Tags = strings.Split(v.TagsCsv, ",")
	}
	return row
}

func (s *Server) handleBackup(w http.ResponseWriter, r *http.Request) {
	dir, err := s.backupTempDir()
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	defer func() { _ = os.RemoveAll(dir) }()

	path := filepath.Join(dir, "snapshot.db")
	if err := backup.Snapshot(r.Context(), s.DB, path); err != nil {
		s.internalError(w, r, err)
		return
	}
	f, err := os.Open(path)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	defer func() { _ = f.Close() }()
	info, err := f.Stat()
	if err != nil {
		s.internalError(w, r, err)
		return
	}

	name := "mealheiro-" + s.Clock.Now().Format("20060102-150405") + ".db"
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", `attachment; filename="`+name+`"`)
	w.Header().Set("Content-Length", strconv.FormatInt(info.Size(), 10))
	if _, err := io.Copy(w, f); err != nil {
		s.Log.Warn("backup download interrupted", "err", err)
	}
}

// backupTempDir creates a scratch dir on the data volume (the container root FS
// is read-only), falling back to the OS temp dir.
func (s *Server) backupTempDir() (string, error) {
	if s.BackupDir != "" {
		if err := os.MkdirAll(s.BackupDir, 0o750); err == nil {
			if dir, err := os.MkdirTemp(s.BackupDir, ".download-*"); err == nil {
				return dir, nil
			}
		}
	}
	return os.MkdirTemp("", "mealheiro-download-*")
}
