package api

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/melojms/mm-budget/internal/clock"
	"github.com/melojms/mm-budget/internal/store"
)

// coreUsageWindowDays is the quick-add ordering window: entries dated within the
// last N days, today included.
const coreUsageWindowDays = 90

const coreDefaultIcon = "circle"

var (
	coreIconRe  = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	coreColorRe = regexp.MustCompile(`^#[0-9a-f]{6}$`)
)

// Category is the JSON shape of a category (docs/API.md "Category").
type Category struct {
	ID         int64  `json:"id"`
	ParentID   *int64 `json:"parent_id"`
	Type       string `json:"type"`
	Name       string `json:"name"`
	Icon       string `json:"icon"`
	Color      string `json:"color"`
	Archived   bool   `json:"archived"`
	UsageCount int64  `json:"usage_count"`
	EntryCount int64  `json:"entry_count"`
}

// coreUsageWindow returns the inclusive date range used for usage_count.
func (s *Server) coreUsageWindow() (since, until string) {
	now := s.Clock.Now()
	return now.AddDate(0, 0, -(coreUsageWindowDays - 1)).Format(clock.DateLayout), now.Format(clock.DateLayout)
}

func (s *Server) coreCategoryJSON(ctx context.Context, id int64) (Category, error) {
	since, until := s.coreUsageWindow()
	c, err := s.Q.GetCategoryWithCounts(ctx, store.GetCategoryWithCountsParams{Since: since, Until: until, ID: id})
	if err != nil {
		return Category{}, err
	}
	return Category(c), nil
}

func (s *Server) handleListCategories(w http.ResponseWriter, r *http.Request) {
	typ := r.URL.Query().Get("type")
	if typ != "" && !coreValidEntryType(typ) {
		writeError(w, http.StatusBadRequest, "invalid type")
		return
	}
	includeArchived := false
	if raw := r.URL.Query().Get("include_archived"); raw != "" {
		v, err := strconv.ParseBool(raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid include_archived")
			return
		}
		includeArchived = v
	}

	since, until := s.coreUsageWindow()
	rows, err := s.Q.ListCategoriesWithCounts(r.Context(), store.ListCategoriesWithCountsParams{Since: since, Until: until})
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	out := make([]Category, 0, len(rows))
	for _, c := range rows {
		if (typ != "" && c.Type != typ) || (c.Archived && !includeArchived) {
			continue
		}
		out = append(out, Category(c))
	}
	writeJSON(w, http.StatusOK, out)
}

// coreValidateCategoryName trims and checks a category name.
func coreValidateCategoryName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" || utf8.RuneCountInString(name) > coreMaxNameLen {
		return "", errors.New("name must be 1-40 characters")
	}
	return name, nil
}

func coreValidateIcon(icon string) (string, error) {
	icon = strings.TrimSpace(icon)
	if len(icon) > 64 || !coreIconRe.MatchString(icon) {
		return "", errors.New("icon must be a kebab-case lucide icon name")
	}
	return icon, nil
}

func coreValidateColor(color string) (string, error) {
	color = strings.ToLower(strings.TrimSpace(color))
	if !coreColorRe.MatchString(color) {
		return "", errors.New("color must be #rrggbb")
	}
	return color, nil
}

// coreBadRequest marks validation failures discovered inside a transaction.
type coreBadRequest struct{ msg string }

func (e coreBadRequest) Error() string { return e.msg }

func (s *Server) handleCreateCategory(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Type     string  `json:"type"`
		Name     string  `json:"name"`
		ParentID *int64  `json:"parent_id"`
		Icon     *string `json:"icon"`
		Color    *string `json:"color"`
	}
	if err := decodeJSON(w, r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body: "+err.Error())
		return
	}
	if !coreValidEntryType(body.Type) {
		writeError(w, http.StatusBadRequest, "invalid type")
		return
	}
	name, err := coreValidateCategoryName(body.Name)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	var icon, color string
	if body.Icon != nil {
		if icon, err = coreValidateIcon(*body.Icon); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	if body.Color != nil {
		if color, err = coreValidateColor(*body.Color); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	ctx := r.Context()
	var id int64
	err = s.coreInTx(ctx, func(q *store.Queries) error {
		if body.ParentID != nil {
			parent, err := q.GetCategory(ctx, *body.ParentID)
			if coreIsNoRows(err) {
				return coreBadRequest{"parent category not found"}
			}
			if err != nil {
				return err
			}
			switch {
			case parent.ParentID != nil:
				return coreBadRequest{"parent must be a top-level category"}
			case parent.Archived:
				return coreBadRequest{"parent category is archived"}
			case parent.Type != body.Type:
				return coreBadRequest{"parent category has a different type"}
			}
			if icon == "" {
				icon = parent.Icon
			}
			if color == "" {
				color = parent.Color
			}
		}
		if icon == "" {
			icon = coreDefaultIcon
		}
		if color == "" {
			used, err := q.ListTopLevelColors(ctx, body.Type)
			if err != nil {
				return err
			}
			color = corePickColor(used)
		}
		id, err = q.CreateCategory(ctx, store.CreateCategoryParams{
			ParentID: body.ParentID, Type: body.Type, Name: name, Icon: icon, Color: color, CreatedAt: s.coreTimestamp(),
		})
		return err
	})
	if s.coreCategoryWriteFailed(w, r, err) {
		return
	}
	c, err := s.coreCategoryJSON(ctx, id)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, c)
}

// corePickColor returns the first palette color not in used, cycling when all are taken.
func corePickColor(used []string) string {
	for _, c := range palette {
		if !slices.ContainsFunc(used, func(u string) bool { return strings.EqualFold(u, c) }) {
			return c
		}
	}
	return palette[len(used)%len(palette)]
}

// coreCategoryWriteFailed maps a category write error to a response; false if err is nil.
func (s *Server) coreCategoryWriteFailed(w http.ResponseWriter, r *http.Request, err error) bool {
	var bad coreBadRequest
	switch {
	case err == nil:
		return false
	case errors.As(err, &bad):
		writeError(w, http.StatusBadRequest, bad.msg)
	case coreIsNoRows(err):
		writeError(w, http.StatusNotFound, "category not found")
	case coreIsUniqueViolation(err):
		writeError(w, http.StatusConflict, "a category with this name already exists here")
	default:
		s.internalError(w, r, err)
	}
	return true
}

func (s *Server) handleUpdateCategory(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r)
	if !ok {
		writeError(w, http.StatusNotFound, "category not found")
		return
	}
	var body struct {
		Name     *string `json:"name"`
		Icon     *string `json:"icon"`
		Color    *string `json:"color"`
		Archived *bool   `json:"archived"`
	}
	if err := decodeJSON(w, r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body: "+err.Error())
		return
	}
	var err error
	var name, icon, color string
	if body.Name != nil {
		if name, err = coreValidateCategoryName(*body.Name); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	if body.Icon != nil {
		if icon, err = coreValidateIcon(*body.Icon); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	if body.Color != nil {
		if color, err = coreValidateColor(*body.Color); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	ctx := r.Context()
	err = s.coreInTx(ctx, func(q *store.Queries) error {
		c, err := q.GetCategory(ctx, id)
		if err != nil {
			return err
		}
		p := store.UpdateCategoryParams{ID: id, Name: c.Name, Icon: c.Icon, Color: c.Color, Archived: c.Archived}
		if body.Name != nil {
			p.Name = name
		}
		if body.Icon != nil {
			p.Icon = icon
		}
		if body.Color != nil {
			p.Color = color
		}
		if body.Archived != nil {
			p.Archived = *body.Archived
		}
		if err := q.UpdateCategory(ctx, p); err != nil {
			return err
		}
		switch {
		case body.Archived == nil:
		case *body.Archived && c.ParentID == nil:
			return q.ArchiveChildCategories(ctx, &id)
		case !*body.Archived && c.ParentID != nil:
			return q.SetCategoryArchived(ctx, store.SetCategoryArchivedParams{Archived: false, ID: *c.ParentID})
		}
		return nil
	})
	if s.coreCategoryWriteFailed(w, r, err) {
		return
	}
	c, err := s.coreCategoryJSON(ctx, id)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (s *Server) handleDeleteCategory(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r)
	if !ok {
		writeError(w, http.StatusNotFound, "category not found")
		return
	}
	ctx := r.Context()
	var inUse bool
	err := s.coreInTx(ctx, func(q *store.Queries) error {
		if _, err := q.GetCategory(ctx, id); err != nil {
			return err
		}
		refs, err := q.CountCategoryReferences(ctx, id)
		if err != nil {
			return err
		}
		if refs.Entries+refs.Templates+refs.Budgets+refs.Children > 0 {
			inUse = true
			return nil
		}
		return q.DeleteCategory(ctx, id)
	})
	if s.coreCategoryWriteFailed(w, r, err) {
		return
	}
	if inUse {
		writeError(w, http.StatusConflict, "category is in use by entries, templates, budgets or subcategories; archive it instead")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
