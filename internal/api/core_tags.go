package api

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/melojms/mealheiro/internal/store"
)

// coreNormalizeTags lowercases, trims, dedupes and sorts tags, dropping empties.
func coreNormalizeTags(raw []string) ([]string, error) {
	out := make([]string, 0, len(raw))
	for _, t := range raw {
		t = strings.ToLower(strings.TrimSpace(t))
		if t == "" {
			continue
		}
		if utf8.RuneCountInString(t) > coreMaxNameLen {
			return nil, fmt.Errorf("tag %q is longer than %d characters", t, coreMaxNameLen)
		}
		out = append(out, t)
	}
	slices.Sort(out)
	return slices.Compact(out), nil
}

// coreReplaceEntryTags upserts tags and makes them the entry's full tag set.
func coreReplaceEntryTags(ctx context.Context, q *store.Queries, entryID int64, tags []string) error {
	if err := q.ClearEntryTags(ctx, entryID); err != nil {
		return err
	}
	for _, name := range tags {
		tagID, err := q.UpsertTag(ctx, name)
		if err != nil {
			return err
		}
		if err := q.AddEntryTag(ctx, store.AddEntryTagParams{EntryID: entryID, TagID: tagID}); err != nil {
			return err
		}
	}
	return nil
}

// TagCount is an element of GET /api/tags.
type TagCount struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

func (s *Server) handleListTags(w http.ResponseWriter, r *http.Request) {
	prefix := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	rows, err := s.Q.ListUsedTags(r.Context(), coreLikeEscape(prefix)+"%")
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	out := make([]TagCount, 0, len(rows))
	for _, t := range rows {
		out = append(out, TagCount(t))
	}
	writeJSON(w, http.StatusOK, out)
}
