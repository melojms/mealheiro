package api

import (
	"strings"

	"github.com/melojms/mealheiro/internal/store"
)

// Entry is the JSON shape of an entry (docs/API.md "Entry").
type Entry struct {
	ID                 int64    `json:"id"`
	Type               string   `json:"type"`
	Date               string   `json:"date"`
	AmountCents        int64    `json:"amount_cents"`
	CategoryID         int64    `json:"category_id"`
	CategoryName       string   `json:"category_name"`
	ParentCategoryID   *int64   `json:"parent_category_id"`
	ParentCategoryName *string  `json:"parent_category_name"`
	PayerID            int64    `json:"payer_id"`
	PayerName          string   `json:"payer_name"`
	Personal           bool     `json:"personal"`
	Note               string   `json:"note"`
	Tags               []string `json:"tags"`
	Status             string   `json:"status"`
	TemplateID         *int64   `json:"template_id"`
	Recurring          bool     `json:"recurring"`
	CreatedAt          string   `json:"created_at"`
	UpdatedAt          string   `json:"updated_at"`
}

// entryFromView maps an entry_view row to its JSON shape.
func entryFromView(v store.EntryView) Entry {
	e := Entry{
		ID:               v.ID,
		Type:             v.Type,
		Date:             v.Date,
		AmountCents:      v.AmountCents,
		CategoryID:       v.CategoryID,
		CategoryName:     v.CategoryName,
		ParentCategoryID: v.ParentCategoryID,
		PayerID:          v.PayerID,
		PayerName:        v.PayerName,
		Personal:         v.Personal,
		Note:             v.Note,
		Tags:             []string{},
		Status:           v.Status,
		TemplateID:       v.TemplateID,
		Recurring:        v.TemplateMonth != nil,
		CreatedAt:        v.CreatedAt,
		UpdatedAt:        v.UpdatedAt,
	}
	if v.ParentCategoryID != nil {
		name := v.ParentCategoryName
		e.ParentCategoryName = &name
	}
	if v.TagsCsv != "" {
		e.Tags = strings.Split(v.TagsCsv, ",")
	}
	return e
}

// Person is the JSON shape of a person.
type Person struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Kind string `json:"kind"`
}

func personFromStore(p store.Person) Person {
	return Person{ID: p.ID, Name: p.Name, Kind: p.Kind}
}
