package utils

import "kithub_backend/models"

func AddKitRecord(catalog *models.Catalog, record models.KitRecord) {
	if record.URL == "" || record.TeamID == "" || record.Season == "" || record.KitType == "" {
		return
	}

	// Filter for only valid seasons (2024, 2025, 2026)
	if !IsValidSeason(record.Season) {
		// If it's "Imported", we'll default it to 2024 if it's not already in a better season
		if record.Season == "Imported" {
			record.Season = "2024"
		} else {
			return
		}
	}

	league := record.League
	if league.Logo == "" {
		league.Logo = leagueLogos[league.ID]
	}
	league.IsPopular = league.IsPopular || isPopularLeague(league.ID)
	catalog.Leagues[league.ID] = league

	if catalog.Seasons == nil {
		catalog.Seasons = make(map[string]models.Season)
	}

	season, ok := catalog.Seasons[record.Season]
	if !ok {
		season = models.Season{
			Teams: make(map[string]models.TeamKits),
		}
	}

	team, ok := season.Teams[record.TeamID]
	if !ok {
		logo := record.TeamLogo
		if teamLogos[record.TeamID] != "" {
			logo = teamLogos[record.TeamID]
		}
		team = models.TeamKits{
			Name:      record.TeamName,
			Logo:      logo,
			League:    league.ID,
			IsPopular: record.TeamPopular || isPopularTeam(record.TeamID) || league.IsPopular,
			Kits:      make(map[string]string),
		}
	}

	if team.Kits == nil {
		team.Kits = make(map[string]string)
	}

	if _, exists := team.Kits[record.KitType]; !exists {
		team.Kits[record.KitType] = record.URL
	}

	season.Teams[record.TeamID] = team
	catalog.Seasons[record.Season] = season
}

func ApplyTeamLogos(catalog *models.Catalog, logos map[string]string) {
	for seasonID, season := range catalog.Seasons {
		for teamID, team := range season.Teams {
			if team.Logo != "" {
				continue
			}

			if scrapedLogo, ok := logos[teamID]; ok && scrapedLogo != "" {
				if predefinedLogo := teamLogos[teamID]; predefinedLogo != "" {
					team.Logo = predefinedLogo
				} else {
					team.Logo = scrapedLogo
				}
				season.Teams[teamID] = team
			}
		}
		catalog.Seasons[seasonID] = season
	}
}
