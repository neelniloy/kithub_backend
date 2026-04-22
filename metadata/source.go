package metadata

type SourceLeague struct {
	ID         string
	ExternalID string
	SearchName string
	IsPopular  bool
	Aliases    []string
}

var SourceLeagues = []SourceLeague{
	{ID: "premier_league", ExternalID: "4328", SearchName: "English Premier League", IsPopular: true, Aliases: []string{"EPL", "English Premier League"}},
	{ID: "la_liga", ExternalID: "4335", SearchName: "Spanish La Liga", IsPopular: true, Aliases: []string{"Spanish La Liga", "LaLiga"}},
	{ID: "bundesliga", ExternalID: "4331", SearchName: "German Bundesliga", IsPopular: true, Aliases: []string{"German Bundesliga"}},
	{ID: "serie_a", ExternalID: "4332", SearchName: "Italian Serie A", IsPopular: true, Aliases: []string{"Italian Serie A"}},
	{ID: "ligue_1", ExternalID: "4334", SearchName: "French Ligue 1", IsPopular: true, Aliases: []string{"French Ligue 1"}},
	{ID: "mls", ExternalID: "4346", SearchName: "American Major League Soccer", IsPopular: true, Aliases: []string{"Major League Soccer"}},
	{ID: "eredivisie", ExternalID: "4337", SearchName: "Dutch Eredivisie", Aliases: []string{"Dutch Eredivisie"}},
	{ID: "saudi_pro_league", ExternalID: "4668", SearchName: "Saudi-Arabian Pro League", IsPopular: true, Aliases: []string{"Saudi Pro League", "Saudi-Arabian Pro League"}},
	{ID: "super_lig", ExternalID: "4339", SearchName: "Turkish Super Lig", IsPopular: true, Aliases: []string{"Turkish Super Lig", "Süper Lig"}},
	{ID: "brasileirao", ExternalID: "4351", SearchName: "Brazilian Serie A", Aliases: []string{"Brazilian Serie A", "Brasileirão"}},
	{ID: "argentine_primera", ExternalID: "4406", SearchName: "Argentinian Primera Division", Aliases: []string{"Argentinian Primera Division", "Argentine Primera Division"}},
	{ID: "liga_1_peru", ExternalID: "4688", SearchName: "Peruvian Primera Division", Aliases: []string{"Peruvian Primera Division", "Liga 1 Peru"}},
}

var internationalLeague = League{
	ID:         "international",
	ExternalID: "4429",
	Name:       "International",
	Logo:       "https://r2.thesportsdb.com/images/media/league/badge/e7er5g1696521789.png",
	IsPopular:  true,
	Aliases:    []string{"FIFA World Cup", "International"},
	Source:     "manual",
}

var nationalTeams = []Team{
	{ID: "argentina", Name: "Argentina", Logo: "https://r2.thesportsdb.com/images/media/team/badge/1c8m0g1687707876.png", League: "international", IsPopular: true, Aliases: []string{"Argentina"}, Source: "manual"},
	{ID: "brazil", Name: "Brazil", Logo: "https://r2.thesportsdb.com/images/media/team/badge/rbdwne1678526703.png", League: "international", IsPopular: true, Aliases: []string{"Brazil", "Brasil"}, Source: "manual"},
	{ID: "england", Name: "England", Logo: "https://r2.thesportsdb.com/images/media/team/badge/i3w3521687708223.png", League: "international", IsPopular: true, Aliases: []string{"England"}, Source: "manual"},
	{ID: "france", Name: "France", Logo: "https://r2.thesportsdb.com/images/media/team/badge/6n9jin1687707835.png", League: "international", IsPopular: true, Aliases: []string{"France"}, Source: "manual"},
	{ID: "germany", Name: "Germany", Logo: "https://r2.thesportsdb.com/images/media/team/badge/2m9p8n1687707806.png", League: "international", IsPopular: true, Aliases: []string{"Germany"}, Source: "manual"},
	{ID: "colombia", ExternalID: "134501", Name: "Colombia", Logo: "https://r2.thesportsdb.com/images/media/team/badge/4ymyku1691180081.png", League: "international", IsPopular: true, Aliases: []string{"Colombia"}, Source: "thesportsdb"},
	{ID: "indonesia", ExternalID: "140164", Name: "Indonesia", Logo: "https://r2.thesportsdb.com/images/media/team/badge/hkg0st1656933330.png", League: "international", Aliases: []string{"Indonesia"}, Source: "thesportsdb"},
	{ID: "japan", ExternalID: "134503", Name: "Japan", Logo: "https://r2.thesportsdb.com/images/media/team/badge/ffsyxz1591989843.png", League: "international", Aliases: []string{"Japan"}, Source: "thesportsdb"},
	{ID: "mexico", ExternalID: "134497", Name: "Mexico", Logo: "https://r2.thesportsdb.com/images/media/team/badge/3rmosi1748525208.png", League: "international", IsPopular: true, Aliases: []string{"Mexico"}, Source: "thesportsdb"},
	{ID: "netherlands", ExternalID: "133905", Name: "Netherlands", Logo: "https://r2.thesportsdb.com/images/media/team/badge/1p0hr41593787110.png", League: "international", Aliases: []string{"Netherlands", "Holland"}, Source: "thesportsdb"},
	{ID: "portugal", Name: "Portugal", Logo: "https://r2.thesportsdb.com/images/media/team/badge/8p7h461687707841.png", League: "international", IsPopular: true, Aliases: []string{"Portugal"}, Source: "manual"},
	{ID: "spain", Name: "Spain", Logo: "https://r2.thesportsdb.com/images/media/team/badge/ja4it51687628717.png", League: "international", IsPopular: true, Aliases: []string{"Spain"}, Source: "manual"},
	{ID: "vietnam", ExternalID: "140161", Name: "Vietnam", Logo: "https://r2.thesportsdb.com/images/media/team/badge/i18iu41597944056.png", League: "international", Aliases: []string{"Vietnam", "Viet Nam", "Việt Nam"}, Source: "thesportsdb"},
}
