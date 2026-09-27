package export

import (
	"bytes"
	"testing"
)

func TestWrite(t *testing.T) {
	tests := []struct {
		name string
		rows []Row
		want string
	}{
		{
			name: "header only",
			want: "\xef\xbb\xbfdate;type;category;subcategory;amount;payer;note;tags;recurring;status;personal\r\n",
		},
		{
			name: "rows with quoting",
			rows: []Row{
				{Date: "2026-03-01", Type: "expense", Category: "House", Subcategory: "Mortgage", AmountCents: 80000, Payer: "Joint", Note: "", Recurring: true, Status: "confirmed"},
				{Date: "2026-03-02", Type: "expense", Category: "Groceries", AmountCents: 1250, Payer: "Person A", Note: `a;b "c"`, Tags: []string{"food", "weekly"}, Status: "pending", Personal: true},
				{Date: "2026-03-03", Type: "income", Category: "Salary", AmountCents: 5, Payer: "Person B", Note: "line1\nline2"},
			},
			want: "\xef\xbb\xbfdate;type;category;subcategory;amount;payer;note;tags;recurring;status;personal\r\n" +
				"2026-03-01;expense;House;Mortgage;800,00;Joint;;;yes;confirmed;no\r\n" +
				"2026-03-02;expense;Groceries;;12,50;Person A;\"a;b \"\"c\"\"\";food,weekly;no;pending;yes\r\n" +
				"2026-03-03;income;Salary;;0,05;Person B;\"line1\r\nline2\";;no;;no\r\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := Write(&buf, tt.rows); err != nil {
				t.Fatal(err)
			}
			if got := buf.String(); got != tt.want {
				t.Fatalf("got\n%q\nwant\n%q", got, tt.want)
			}
		})
	}
}
