package reports

import (
	"fmt"
	"strconv"
	"strings"
)

// FormatCents formats cents like the UI (pt-PT): 41230 -> "412,30 €",
// 123456 -> "1234,56 €", 1234567 -> "12 345,67 €".
func FormatCents(cents int64) string {
	sign := ""
	if cents < 0 {
		sign = "-"
		cents = -cents
	}
	whole := strconv.FormatInt(cents/100, 10)
	// pt-PT only groups thousands from five integer digits on.
	if len(whole) > 4 {
		var b strings.Builder
		for i, r := range whole {
			if i > 0 && (len(whole)-i)%3 == 0 {
				b.WriteByte(' ')
			}
			b.WriteRune(r)
		}
		whole = b.String()
	}
	return fmt.Sprintf("%s%s,%02d €", sign, whole, cents%100)
}
