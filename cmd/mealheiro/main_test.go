package main

import "testing"

func TestParsePort(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    int
		wantErr bool
	}{
		{"valid", "9090", 9090, false},
		{"lowest", "1", 1, false},
		{"highest", "65535", 65535, false},
		{"empty defaults to 7447", "", 7447, false},
		{"non-numeric", "http", 0, true},
		{"addr form", ":7447", 0, true},
		{"zero", "0", 0, true},
		{"negative", "-1", 0, true},
		{"too high", "65536", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePort(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parsePort(%q) err = %v, wantErr %v", tt.in, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("parsePort(%q) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}
