package utils

import "strings"

func DetectKitType(text string) string {
	normalized := strings.ToLower(strings.ReplaceAll(text, "-", " "))

	isGK := containsAny(normalized, []string{
		"goalkeeper", "goal keeper", " gk ", "_gk_", "/gk", "keeper",
	})

	switch {
	case isGK && containsAny(normalized, []string{"third", "3rd"}):
		return "gk_third"
	case isGK && containsAny(normalized, []string{"away", "visitor"}):
		return "gk_away"
	case isGK:
		return "gk_home"
	case containsAny(normalized, []string{"third", "3rd"}):
		return "third"
	case containsAny(normalized, []string{"away", "visitor"}):
		return "away"
	case containsAny(normalized, []string{"home", "local"}):
		return "home"
	default:
		return "unknown"
	}
}

func containsAny(text string, needles []string) bool {
	for _, needle := range needles {
		if strings.Contains(text, needle) {
			return true
		}
	}
	return false
}
