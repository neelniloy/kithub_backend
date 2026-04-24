package utils

import "kithub_backend/models"

func AddKitRecord(catalog *models.Catalog, record models.KitRecord) {
	if record.URL == "" || record.TeamID == "" || record.Season == "" || record.KitType == "" {
		return
	}

	// Filter for only valid seasons (2024, 2025, 2026)
	if !IsValidSeason(record.Season) {
		// Default to 2024 if it's "Imported" but not recognized as 2025/26
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

	if catalog.Teams == nil {
		catalog.Teams = make(map[string]models.Team)
	}

	team, ok := catalog.Teams[record.TeamID]
	if !ok {
		logo := record.TeamLogo
		if teamLogos[record.TeamID] != "" {
			logo = teamLogos[record.TeamID]
		}
		team = models.Team{
			Name:      record.TeamName,
			Logo:      logo,
			League:    league.ID,
			IsPopular: record.TeamPopular || isPopularTeam(record.TeamID) || league.IsPopular,
			Seasons:   make(map[string]map[string]string),
		}
	}

	if team.Seasons == nil {
		team.Seasons = make(map[string]map[string]string)
	}

	if team.Seasons[record.Season] == nil {
		team.Seasons[record.Season] = make(map[string]string)
	}

	// Don't overwrite if kit already exists from a previous article (unless it was unknown)
	if _, exists := team.Seasons[record.Season][record.KitType]; !exists || team.Seasons[record.Season][record.KitType] == "" {
		team.Seasons[record.Season][record.KitType] = record.URL
	}

	catalog.Teams[record.TeamID] = team
}

func ApplyTeamLogos(catalog *models.Catalog, logos map[string]string) {
	for teamID, team := range catalog.Teams {
		if team.Logo != "" {
			continue
		}

		if scrapedLogo, ok := logos[teamID]; ok && scrapedLogo != "" {
			if predefinedLogo := teamLogos[teamID]; predefinedLogo != "" {
				team.Logo = predefinedLogo
			} else {
				team.Logo = scrapedLogo
			}
			catalog.Teams[teamID] = team
		}
	}
}
