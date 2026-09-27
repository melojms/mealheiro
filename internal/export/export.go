// Package export writes entries as a spreadsheet-friendly CSV (docs/API.md
// "Export & backup"): UTF-8 BOM, ';' separator, CRLF, decimal comma.
package export

import (
	"encoding/csv"
	"io"
	"strings"

	"github.com/melojms/mealheiro/internal/money"
)

// bom marks the file as UTF-8 for spreadsheet apps.
const bom = "\xef\xbb\xbf"

// Header is the fixed CSV header row.
var Header = []string{"date", "type", "category", "subcategory", "amount", "payer", "note", "tags", "recurring", "status", "personal"}

// Row is one exported entry.
type Row struct {
	Date        string
	Type        string
	Category    string // top-level category name
	Subcategory string // leaf name, empty when booked on a top-level category
	AmountCents int64
	Payer       string
	Note        string
	Tags        []string
	Recurring   bool
	Status      string
	Personal    bool
}

// Write writes the BOM, header and rows to w.
func Write(w io.Writer, rows []Row) error {
	if _, err := io.WriteString(w, bom); err != nil {
		return err
	}
	cw := csv.NewWriter(w)
	cw.Comma = ';'
	cw.UseCRLF = true
	if err := cw.Write(Header); err != nil {
		return err
	}
	for _, r := range rows {
		if err := cw.Write([]string{
			r.Date, r.Type, r.Category, r.Subcategory, money.Decimal(r.AmountCents),
			r.Payer, r.Note, strings.Join(r.Tags, ","), yesNo(r.Recurring), r.Status, yesNo(r.Personal),
		}); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
