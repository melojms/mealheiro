// Package money formats integer euro cents. Amounts are always stored as int64 cents.
package money

import (
	"fmt"
	"strconv"
	"strings"
)

// Decimal formats cents as a plain decimal-comma number without currency or
// thousands separators, e.g. 123456 -> "1234,56", -5 -> "-0,05". Used for CSV.
func Decimal(cents int64) string {
	sign := ""
	if cents < 0 {
		sign = "-"
		cents = -cents
	}
	return fmt.Sprintf("%s%d,%02d", sign, cents/100, cents%100)
}

// Parse parses user input like "12,50", "12.5", "1 234,56", "1.234,56" or "7"
// into positive cents. The last '.' or ',' followed by 1-2 digits is the decimal separator.
func Parse(s string) (int64, error) {
	s = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(s), "€"))
	s = strings.ReplaceAll(s, " ", "")
	if s == "" {
		return 0, fmt.Errorf("empty amount")
	}
	intPart, frac := s, ""
	if i := strings.LastIndexAny(s, ".,"); i >= 0 && len(s)-i-1 <= 2 {
		intPart, frac = s[:i], s[i+1:]
	}
	intPart = strings.NewReplacer(".", "", ",", "").Replace(intPart)
	if intPart == "" {
		intPart = "0"
	}
	for len(frac) < 2 {
		frac += "0"
	}
	whole, err := strconv.ParseInt(intPart, 10, 64)
	if err != nil || whole < 0 {
		return 0, fmt.Errorf("invalid amount %q", s)
	}
	f, err := strconv.ParseInt(frac, 10, 64)
	if err != nil || f < 0 {
		return 0, fmt.Errorf("invalid amount %q", s)
	}
	return whole*100 + f, nil
}
