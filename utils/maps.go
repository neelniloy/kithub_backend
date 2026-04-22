package utils

import "kithub_backend/models"

type leagueInfo struct {
	ID   string
	Name string
}

var leagueMapping = map[string]leagueInfo{
	"barcelona":           {"la_liga", "La Liga"},
	"real_madrid":         {"la_liga", "La Liga"},
	"atletico_madrid":     {"la_liga", "La Liga"},
	"sevilla":             {"la_liga", "La Liga"},
	"manchester_united":   {"premier_league", "Premier League"},
	"manchester_city":     {"premier_league", "Premier League"},
	"liverpool":           {"premier_league", "Premier League"},
	"chelsea":             {"premier_league", "Premier League"},
	"arsenal":             {"premier_league", "Premier League"},
	"tottenham_hotspur":   {"premier_league", "Premier League"},
	"bayern_munich":       {"bundesliga", "Bundesliga"},
	"bayer_leverkusen":    {"bundesliga", "Bundesliga"},
	"borussia_dortmund":   {"bundesliga", "Bundesliga"},
	"paris_saint_germain": {"ligue_1", "Ligue 1"},
	"juventus":            {"serie_a", "Serie A"},
	"inter_milan":         {"serie_a", "Serie A"},
	"ac_milan":            {"serie_a", "Serie A"},
	"napoli":              {"serie_a", "Serie A"},
	"ajax":                {"eredivisie", "Eredivisie"},
	"benfica":             {"primeira_liga", "Primeira Liga"},
	"porto":               {"primeira_liga", "Primeira Liga"},
	"sporting_cp":         {"primeira_liga", "Primeira Liga"},
	"al_nassr":            {"saudi_pro_league", "Saudi Pro League"},
	"al_hilal":            {"saudi_pro_league", "Saudi Pro League"},
	"inter_miami":         {"mls", "Major League Soccer"},
	"brentford":           {"premier_league", "Premier League"},
	"crystal_palace":      {"premier_league", "Premier League"},
	"leeds_united":        {"premier_league", "Premier League"},
	"wolves":              {"premier_league", "Premier League"},
	"valencia":            {"la_liga", "La Liga"},
	"galatasaray":         {"super_lig", "Super Lig"},
	"santos":              {"brasileirao", "Brasileirao"},
	"boca_juniors":        {"argentine_primera", "Argentine Primera Division"},
	"banfield":            {"argentine_primera", "Argentine Primera Division"},
	"talleres":            {"argentine_primera", "Argentine Primera Division"},
	"colon":               {"argentine_primera", "Argentine Primera Division"},
	"atlético_tucumán":    {"argentine_primera", "Argentine Primera Division"},
	"sporting_cristal":    {"liga_1_peru", "Liga 1 Peru"},
	"argentina":           {"international", "International"},
	"brazil":              {"international", "International"},
	"colombia":            {"international", "International"},
	"france":              {"international", "International"},
	"germany":             {"international", "International"},
	"england":             {"international", "International"},
	"spain":               {"international", "International"},
	"portugal":            {"international", "International"},
	"netherlands":         {"international", "International"},
	"indonesia":           {"international", "International"},
	"japan":               {"international", "International"},
	"mexico":              {"international", "International"},
	"vietnam":             {"international", "International"},
}

var teamLogos = map[string]string{}

var leagueLogos = map[string]string{
	"argentine_primera": "https://r2.thesportsdb.com/images/media/league/badge/rk9xhx1768238251.png",
	"brasileirao":       "https://r2.thesportsdb.com/images/media/league/badge/lywv7t1766787179.png",
	"bundesliga":        "https://r2.thesportsdb.com/images/media/league/badge/teqh1b1679952008.png",
	"eredivisie":        "https://r2.thesportsdb.com/images/media/league/badge/5cdsu21725984946.png",
	"international":     "https://r2.thesportsdb.com/images/media/league/badge/e7er5g1696521789.png",
	"la_liga":           "https://r2.thesportsdb.com/images/media/league/badge/ja4it51687628717.png",
	"liga_1_peru":       "https://r2.thesportsdb.com/images/media/league/badge/1ujpwc1580040216.png",
	"ligue_1":           "https://r2.thesportsdb.com/images/media/league/badge/9f7z9d1742983155.png",
	"mls":               "https://r2.thesportsdb.com/images/media/league/badge/dqo6r91549878326.png",
	"premier_league":    "https://r2.thesportsdb.com/images/media/league/badge/gasy9d1737743125.png",
	"saudi_pro_league":  "https://r2.thesportsdb.com/images/media/league/badge/w67i621701772123.png",
	"serie_a":           "https://r2.thesportsdb.com/images/media/league/badge/67q3q21679951383.png",
	"super_lig":         "https://r2.thesportsdb.com/images/media/league/badge/h7xx231601671132.png",
}

var popularTeams = map[string]bool{
	"barcelona":           true,
	"real_madrid":         true,
	"manchester_united":   true,
	"manchester_city":     true,
	"liverpool":           true,
	"chelsea":             true,
	"arsenal":             true,
	"bayern_munich":       true,
	"bayer_leverkusen":    true,
	"borussia_dortmund":   true,
	"paris_saint_germain": true,
	"juventus":            true,
	"inter_milan":         true,
	"ac_milan":            true,
	"al_nassr":            true,
	"inter_miami":         true,
	"galatasaray":         true,
	"boca_juniors":        true,
	"argentina":           true,
	"brazil":              true,
	"colombia":            true,
	"france":              true,
	"portugal":            true,
	"mexico":              true,
}

var popularLeagues = map[string]bool{
	"la_liga":          true,
	"premier_league":   true,
	"bundesliga":       true,
	"serie_a":          true,
	"ligue_1":          true,
	"international":    true,
	"saudi_pro_league": true,
	"mls":              true,
	"super_lig":        true,
}

func LeagueForTeam(teamID string) models.League {
	if league, ok := leagueMapping[teamID]; ok {
		return models.League{ID: league.ID, Name: league.Name}
	}
	return models.League{ID: "unknown", Name: "Unknown"}
}

func isPopularTeam(teamID string) bool {
	return popularTeams[teamID]
}

func isPopularLeague(leagueID string) bool {
	return popularLeagues[leagueID]
}
