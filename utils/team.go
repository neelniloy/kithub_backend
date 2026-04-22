package utils

import (
	"regexp"
	"strings"
	"unicode"
)

type teamAlias struct {
	ID   string
	Name string
}

var teamAliases = map[string]teamAlias{
	"barcelona":           {"barcelona", "FC Barcelona"},
	"fc barcelona":        {"barcelona", "FC Barcelona"},
	"real madrid":         {"real_madrid", "Real Madrid"},
	"real madird":         {"real_madrid", "Real Madrid"},
	"r madrid":            {"real_madrid", "Real Madrid"},
	"atletico madrid":     {"atletico_madrid", "Atletico Madrid"},
	"atletico de madrid":  {"atletico_madrid", "Atletico Madrid"},
	"sevilla":             {"sevilla", "Sevilla"},
	"man united":          {"manchester_united", "Manchester United"},
	"man utd":             {"manchester_united", "Manchester United"},
	"manchester united":   {"manchester_united", "Manchester United"},
	"manchester city":     {"manchester_city", "Manchester City"},
	"man city":            {"manchester_city", "Manchester City"},
	"liverpool":           {"liverpool", "Liverpool"},
	"chelsea":             {"chelsea", "Chelsea"},
	"arsenal":             {"arsenal", "Arsenal"},
	"tottenham":           {"tottenham_hotspur", "Tottenham Hotspur"},
	"tottenham hotspur":   {"tottenham_hotspur", "Tottenham Hotspur"},
	"bayern munich":       {"bayern_munich", "Bayern Munich"},
	"bayern":              {"bayern_munich", "Bayern Munich"},
	"borussia dortmund":   {"borussia_dortmund", "Borussia Dortmund"},
	"dortmund":            {"borussia_dortmund", "Borussia Dortmund"},
	"bayer 04 leverkusen": {"bayer_leverkusen", "Bayer Leverkusen"},
	"bayer leverkusen":    {"bayer_leverkusen", "Bayer Leverkusen"},
	"paris saint germain": {"paris_saint_germain", "Paris Saint-Germain"},
	"paris sg":            {"paris_saint_germain", "Paris Saint-Germain"},
	"psg":                 {"paris_saint_germain", "Paris Saint-Germain"},
	"juventus":            {"juventus", "Juventus"},
	"inter milan":         {"inter_milan", "Inter Milan"},
	"internazionale":      {"inter_milan", "Inter Milan"},
	"ac milan":            {"ac_milan", "AC Milan"},
	"milan":               {"ac_milan", "AC Milan"},
	"napoli":              {"napoli", "Napoli"},
	"ajax":                {"ajax", "Ajax"},
	"benfica":             {"benfica", "Benfica"},
	"porto":               {"porto", "Porto"},
	"sporting cp":         {"sporting_cp", "Sporting CP"},
	"sporting lisbon":     {"sporting_cp", "Sporting CP"},
	"al nassr":            {"al_nassr", "Al Nassr"},
	"al hilal":            {"al_hilal", "Al Hilal"},
	"inter miami":         {"inter_miami", "Inter Miami"},
	"brentford":           {"brentford", "Brentford"},
	"crystal palace":      {"crystal_palace", "Crystal Palace"},
	"leeds united":        {"leeds_united", "Leeds United"},
	"wolves":              {"wolves", "Wolverhampton Wanderers"},
	"wolverhampton":       {"wolves", "Wolverhampton Wanderers"},
	"valencia":            {"valencia", "Valencia"},
	"galatasaray":         {"galatasaray", "Galatasaray"},
	"santos":              {"santos", "Santos FC"},
	"boca juniors":        {"boca_juniors", "Boca Juniors"},
	"banfield":            {"banfield", "Banfield"},
	"talleres":            {"talleres", "Talleres"},
	"colon":               {"colon", "Colon"},
	"colón":               {"colon", "Colon"},
	"sporting cristal":    {"sporting_cristal", "Sporting Cristal"},
	"atlético tucumán":    {"atlético_tucumán", "Atlético Tucumán"},
	"atletico tucuman":    {"atlético_tucumán", "Atlético Tucumán"},
	"argentina":           {"argentina", "Argentina"},
	"brazil":              {"brazil", "Brazil"},
	"brasileiro":          {"brazil", "Brazil"},
	"colombia":            {"colombia", "Colombia"},
	"france":              {"france", "France"},
	"germany":             {"germany", "Germany"},
	"england":             {"england", "England"},
	"spain":               {"spain", "Spain"},
	"portugal":            {"portugal", "Portugal"},
	"netherlands":         {"netherlands", "Netherlands"},
	"indonesia":           {"indonesia", "Indonesia"},
	"japan":               {"japan", "Japan"},
	"mexico":              {"mexico", "Mexico"},
	"vietnam":             {"vietnam", "Vietnam"},
	"việt nam":            {"vietnam", "Vietnam"},
}

var cleanupPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\b(dream\s+league\s+soccer|dls|fts|kit|kits|url|logo|png)\b`),
	regexp.MustCompile(`(?i)\b(20\d{2}\s*[-/]\s*\d{2}|20\d{2}|\d{2}\s*[-/]\s*\d{2}|\d{1,4})\b`),
	regexp.MustCompile(`(?i)\b(home|away|third|goalkeeper|gk|new|latest|updated)\b`),
	regexp.MustCompile(`(?i)\b(download|for|in|world|cup|football|league|leagues|club|fc|team|game|found|nothing|update|pack)\b`),
}

var blockedFallbackTeams = map[string]bool{
	"game":             true,
	"nothing":          true,
	"nothing found":    true,
	"football":         true,
	"football league":  true,
	"football leagues": true,
	"league":           true,
	"leagues":          true,
	"world cup":        true,
	"saudi pro":        true,
}

func NormalizeTeamName(raw string) (string, string) {
	candidate := normalizeText(raw)
	if candidate == "" {
		return "", ""
	}

	for alias, team := range teamAliases {
		if containsWordSequence(candidate, alias) {
			return team.ID, team.Name
		}
	}

	cleaned := candidate
	for _, pattern := range cleanupPatterns {
		cleaned = pattern.ReplaceAllString(cleaned, " ")
	}
	cleaned = strings.Join(strings.Fields(cleaned), " ")
	cleaned = strings.Trim(cleaned, " -_|:")
	if !validFallbackTeam(cleaned) {
		return "", ""
	}

	words := strings.Fields(cleaned)
	if len(words) > 4 {
		words = words[:4]
	}

	displayName := titleCase(strings.Join(words, " "))
	id := strings.Join(words, "_")
	return id, displayName
}

func normalizeText(raw string) string {
	var builder strings.Builder
	for _, r := range strings.ToLower(raw) {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r):
			builder.WriteRune(r)
		default:
			builder.WriteRune(' ')
		}
	}
	return strings.Join(strings.Fields(builder.String()), " ")
}

func containsWordSequence(text, seq string) bool {
	return regexp.MustCompile(`(^|\s)` + regexp.QuoteMeta(seq) + `(\s|$)`).MatchString(text)
}

func titleCase(value string) string {
	words := strings.Fields(value)
	for i, word := range words {
		runes := []rune(word)
		if len(runes) <= 2 {
			words[i] = strings.ToUpper(word)
			continue
		}
		runes[0] = unicode.ToTitle(runes[0])
		words[i] = string(runes)
	}
	return strings.Join(words, " ")
}

func validFallbackTeam(cleaned string) bool {
	if cleaned == "" || len([]rune(cleaned)) < 3 || blockedFallbackTeams[cleaned] {
		return false
	}

	hasLetter := false
	for _, r := range cleaned {
		if unicode.IsLetter(r) {
			hasLetter = true
			break
		}
	}
	if !hasLetter {
		return false
	}

	words := strings.Fields(cleaned)
	if len(words) == 0 || len(words) > 4 {
		return false
	}
	for _, word := range words {
		if blockedFallbackTeams[word] || len([]rune(word)) == 1 {
			return false
		}
	}

	return true
}
