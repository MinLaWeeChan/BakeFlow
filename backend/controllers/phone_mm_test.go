package controllers

import "testing"

func TestNormalizeMyanmarPhoneE164_Valid(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"09123456789", "+959123456789"},
		{"+959123456789", "+959123456789"},
		{"09 123-456-789", "+959123456789"},
		{"  09-123456789  ", "+959123456789"},
	}

	for _, tc := range tests {
		got, ok := NormalizeMyanmarPhoneE164(tc.in)
		if !ok {
			t.Fatalf("expected ok for %q", tc.in)
		}
		if got != tc.want {
			t.Fatalf("NormalizeMyanmarPhoneE164(%q)=%q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestNormalizeMyanmarPhoneE164_Invalid(t *testing.T) {
	tests := []string{
		"",
		"+",
		"0912345678",
		"091234567890",
		"+95912345678",
		"+9591234567890",
		"08123456789",
		"+958123456789",
		"09abcdefghi",
		"09 123 456 78a",
	}

	for _, in := range tests {
		if got, ok := NormalizeMyanmarPhoneE164(in); ok {
			t.Fatalf("expected invalid for %q, got %q", in, got)
		}
	}
}
