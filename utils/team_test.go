package utils

import "testing"

func TestNormalizeTeamName(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		id   string
		name string
	}{
		"FC Barcelona DLS Kits 2025":          {"barcelona", "FC Barcelona"},
		"Man United home kit 2025/26":         {"manchester_united", "Manchester United"},
		"Real Madird DLS Kit 25-26":           {"real_madrid", "Real Madrid"},
		"DLS 512 x 512 Logo Pack":             {"", ""},
		"Nothing Found":                       {"", ""},
		"Wolves DLS Kit 25-26 Home":           {"wolves", "Wolverhampton Wanderers"},
		"Bayer 04 Leverkusen DLS Kits 2025":   {"bayer_leverkusen", "Bayer Leverkusen"},
		"Atlético Tucumán Football League 26": {"atlético_tucumán", "Atlético Tucumán"},
	}

	for input, want := range tests {
		gotID, gotName := NormalizeTeamName(input)
		if gotID != want.id || gotName != want.name {
			t.Fatalf("NormalizeTeamName(%q) = (%q, %q), want (%q, %q)", input, gotID, gotName, want.id, want.name)
		}
	}
}
