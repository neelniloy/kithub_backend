package metadata

type SourceLeague struct {
	ID         string
	ExternalID string
	SearchName string
	IsPopular  bool
	Aliases    []string
}

var SourceLeagues = []SourceLeague{
	// Europe
	{ID: "premier_league", ExternalID: "4328", SearchName: "English Premier League", IsPopular: true},
	{ID: "la_liga", ExternalID: "4335", SearchName: "Spanish La Liga", IsPopular: true},
	{ID: "bundesliga", ExternalID: "4331", SearchName: "German Bundesliga", IsPopular: true},
	{ID: "serie_a", ExternalID: "4332", SearchName: "Italian Serie A", IsPopular: true},
	{ID: "ligue_1", ExternalID: "4334", SearchName: "French Ligue 1", IsPopular: true},
	{ID: "eredivisie", ExternalID: "4337", SearchName: "Dutch Eredivisie"},
	{ID: "primeira_liga", ExternalID: "4344", SearchName: "Portuguese Primeira Liga"},
	{ID: "championship", ExternalID: "4329", SearchName: "English League Championship"},
	{ID: "scottish_premiership", ExternalID: "4330", SearchName: "Scottish Premiership"},
	{ID: "belgian_pro_league", ExternalID: "4338", SearchName: "Belgian First Division A"},
	{ID: "super_lig", ExternalID: "4339", SearchName: "Turkish Super Lig", IsPopular: true},
	{ID: "greek_super_league", ExternalID: "4336", SearchName: "Greek Superleague Greece"},
	{ID: "russian_premier_league", ExternalID: "4355", SearchName: "Russian Premier League"},
	{ID: "austrian_bundesliga", ExternalID: "4345", SearchName: "Austrian Football Bundesliga"},
	{ID: "danish_superliga", ExternalID: "4340", SearchName: "Danish Superliga"},
	{ID: "swiss_super_league", ExternalID: "4341", SearchName: "Swiss Super League"},

	// Americas
	{ID: "mls", ExternalID: "4346", SearchName: "American Major League Soccer", IsPopular: true},
	{ID: "brasileirao", ExternalID: "4351", SearchName: "Brazilian Serie A", Aliases: []string{"Brazilian Serie A", "Brasileirão"}},
	{ID: "argentine_primera", ExternalID: "4406", SearchName: "Argentinian Primera Division", Aliases: []string{"Argentinian Primera Division", "Argentine Primera Division"}},
	{ID: "liga_mx", ExternalID: "4350", SearchName: "Mexican Liga MX", Aliases: []string{"Liga MX", "Primera División de México"}},
	{ID: "colombian_primera_a", ExternalID: "4403", SearchName: "Colombian Primera A"},
	{ID: "liga_1_peru", ExternalID: "4688", SearchName: "Peruvian Primera Division", Aliases: []string{"Peruvian Primera Division", "Liga 1 Peru"}},

	// Asia & Others
	{ID: "saudi_pro_league", ExternalID: "4668", SearchName: "Saudi-Arabian Pro League", IsPopular: true, Aliases: []string{"Saudi Pro League", "Saudi-Arabian Pro League"}},
	{ID: "j1_league", ExternalID: "4400", SearchName: "Japanese J1 League", Aliases: []string{"J1 League"}},
	{ID: "k_league_1", ExternalID: "4401", SearchName: "South Korean K League 1"},
	{ID: "chinese_super_league", ExternalID: "4356", SearchName: "Chinese Super League"},
	{ID: "australian_a_league", ExternalID: "4347", SearchName: "Australian A-League"},
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
