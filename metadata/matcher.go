package metadata

import (
	"regexp"
	"sort"
	"strings"
)

type Matcher struct {
	store  Store
	lookup []lookupEntry
}

type lookupEntry struct {
	alias  string
	teamID string
}

var noisyTitlePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\b(dream\s+league\s+soccer|dls|fts|kit|kits|url|logo|png|download)\b`),
	regexp.MustCompile(`(?i)\b(20\d{2}\s*[-/]\s*\d{2}|20\d{2}|\d{2}\s*[-/]\s*\d{2}|\d{1,4})\b`),
	regexp.MustCompile(`(?i)\b(home|away|third|goalkeeper|gk|new|latest|updated|leaked|for|in|world|cup|football|league|leagues|game|found|nothing)\b`),
}

func NewMatcher(store Store) *Matcher {
	var entries []lookupEntry
	for teamID, team := range store.Teams {
		aliases := []string{team.ID, team.Name}
		aliases = append(aliases, team.Aliases...)
		for _, alias := range aliases {
			normalized := NormalizeAlias(alias)
			if normalized == "" {
				continue
			}
			entries = append(entries, lookupEntry{alias: normalized, teamID: teamID})
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return len(entries[i].alias) > len(entries[j].alias)
	})

	return &Matcher{store: store, lookup: entries}
}

func (m *Matcher) MatchTeam(text string) (Match, bool) {
	if text == "" {
		return Match{}, false
	}
	normalized := NormalizeAlias(text)
	textSlug := Slug(text)

	// 1. Direct Alias Match (Highest Priority)
	for _, entry := range m.lookup {
		if containsAlias(normalized, entry.alias) {
			team := m.store.Teams[entry.teamID]
			league := m.store.Leagues[team.League]
			return Match{Team: team, League: league}, true
		}
	}

	// 2. Slug-based Fuzzy Match (Secondary Priority)
	for _, team := range m.store.Teams {
		teamSlug := Slug(team.Name)
		if strings.Contains(textSlug, teamSlug) || strings.Contains(teamSlug, textSlug) {
			if len(teamSlug) > 3 { // Avoid matching very short names like "AFC" too aggressively
				league := m.store.Leagues[team.League]
				return Match{Team: team, League: league}, true
			}
		}
		
		// Check aliases slugs too
		for _, alias := range team.Aliases {
			aliasSlug := Slug(alias)
			if aliasSlug != "" && (strings.Contains(textSlug, aliasSlug) || strings.Contains(aliasSlug, textSlug)) {
				if len(aliasSlug) > 3 {
					league := m.store.Leagues[team.League]
					return Match{Team: team, League: league}, true
				}
			}
		}
	}

	return Match{}, false
}

func (m *Matcher) FallbackTeam(text string) (Match, bool) {
	cleaned := text
	for _, pattern := range noisyTitlePatterns {
		cleaned = pattern.ReplaceAllString(cleaned, " ")
	}
	cleaned = strings.Join(strings.Fields(NormalizeAlias(cleaned)), " ")
	if !validFallback(cleaned) {
		return Match{}, false
	}

	id := Slug(cleaned)
	if id == "" {
		return Match{}, false
	}
	team := Team{ID: id, Name: titleCase(cleaned), League: "unknown", Source: "scraped_fallback"}
	league := League{ID: "unknown", Name: "Unknown", Source: "scraped_fallback"}
	return Match{Team: team, League: league}, true
}

func (m *Matcher) Store() Store {
	return m.store
}

func containsAlias(text, alias string) bool {
	return regexp.MustCompile(`(^|\s)` + regexp.QuoteMeta(alias) + `(\s|$)`).MatchString(text)
}

func validFallback(cleaned string) bool {
	if cleaned == "" {
		return false
	}

	words := strings.Fields(cleaned)
	if len(words) == 0 || len(words) > 4 {
		return false
	}

	blocked := map[string]bool{
		"football": true,
		"found":    true,
		"game":     true,
		"league":   true,
		"leagues":  true,
		"nothing":  true,
		"pro":      true,
		"saudi":    true,
		"world":    true,
		"cup":      true,
	}
	for _, word := range words {
		if blocked[word] || len(word) == 1 {
			return false
		}
	}

	return true
}

func titleCase(value string) string {
	words := strings.Fields(value)
	for i, word := range words {
		if len(word) <= 2 {
			words[i] = strings.ToUpper(word)
			continue
		}
		words[i] = strings.ToUpper(word[:1]) + word[1:]
	}
	return strings.Join(words, " ")
}
