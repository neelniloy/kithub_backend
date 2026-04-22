package utils

import "testing"

func TestExtractSeason(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"Barcelona 2025-26 DLS kits":       "2025-26",
		"Barcelona 2025/26 DLS kits":       "2025-26",
		"AC Milan 2024-2025 kits":          "2024-25",
		"AC Milan Home Kit 25-26 DLS 25":   "2025-26",
		"Manchester United 2025 kits":      "2025-26",
		"published 2024-01-10":             "2024-25",
		"bad synthetic range 2025-29 kits": "2025-26",
	}

	for input, want := range tests {
		got := ExtractSeason(input)
		if got != want {
			t.Fatalf("ExtractSeason(%q) = %q, want %q", input, got, want)
		}
	}
}
