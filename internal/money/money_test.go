package money

import "testing"

func TestDecimal(t *testing.T) {
	cases := map[int64]string{0: "0,00", 5: "0,05", 1250: "12,50", 123456: "1234,56", -5: "-0,05"}
	for in, want := range cases {
		if got := Decimal(in); got != want {
			t.Errorf("Decimal(%d) = %q, want %q", in, got, want)
		}
	}
}

func TestParse(t *testing.T) {
	cases := map[string]int64{
		"12,50": 1250, "12.50": 1250, "12.5": 1250, "7": 700, "0,05": 5,
		"1 234,56": 123456, "1.234,56": 123456, "1,234.56": 123456, "12 €": 1200, "1.234": 123400,
	}
	for in, want := range cases {
		got, err := Parse(in)
		if err != nil || got != want {
			t.Errorf("Parse(%q) = %d, %v; want %d", in, got, err, want)
		}
	}
	for _, bad := range []string{"", "abc", "-5", "1,2,3x"} {
		if _, err := Parse(bad); err == nil {
			t.Errorf("Parse(%q) expected error", bad)
		}
	}
}
