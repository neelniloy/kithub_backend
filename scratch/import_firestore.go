package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"kithub_backend/metadata"
	"kithub_backend/models"
)

type FirestoreTeam struct {
	ID        string   `json:"teamId"`
	Name      string   `json:"teamName"`
	Logo      string   `json:"teamLogo"`
	League    string   `json:"league"`
	Home      string   `json:"homeKit"`
	Away      string   `json:"awayKit"`
	Third     string   `json:"thirdKit"`
	GKHome    string   `json:"goalHome"`
	GKAway    string   `json:"goalAway"`
	GKThird   string   `json:"goalThird"`
	IsPopular bool     `json:"trending"`
}

func main() {
	// 1. Load existing Firestore teams
	teamData, err := os.ReadFile(filepath.Join("temp", "_Team.json"))
	if err != nil {
		log.Fatalf("failed to read _Team.json: %v", err)
	}

	var fTeams []FirestoreTeam
	if err := json.Unmarshal(teamData, &fTeams); err != nil {
		log.Fatalf("failed to unmarshal teams: %v", err)
	}

	// 2. Load metadata store
	store, err := metadata.Load(metadata.DefaultDataDir)
	if err != nil {
		store = metadata.Store{
			Leagues: make(map[string]metadata.League),
			Teams:   make(map[string]metadata.Team),
		}
	}

	// 3. Load current catalog (kits.json) if it exists
	catalog := models.Catalog{
		Version:     1,
		LastUpdated: time.Now().UTC().Format(time.DateOnly),
		Leagues:     make(map[string]models.League),
		Teams:       make(map[string]models.Team),
	}

	existingData, err := os.ReadFile("kits.json")
	if err == nil {
		json.Unmarshal(existingData, &catalog)
	}

	// 4. Merge Firestore data into Catalog and Store
	importedCount := 0
	for _, ft := range fTeams {
		if ft.Name == "" {
			continue
		}

		teamID := metadata.Slug(ft.Name)
		leagueID := metadata.Slug(ft.League)

		// Create league if not exists
		if _, ok := catalog.Leagues[leagueID]; !ok {
			catalog.Leagues[leagueID] = models.League{
				ID:   leagueID,
				Name: ft.League,
			}
			store.Leagues[leagueID] = metadata.League{
				ID:   leagueID,
				Name: ft.League,
			}
		}

		// Create/Update team
		team, ok := catalog.Teams[teamID]
		if !ok {
			team = models.Team{
				Name:      ft.Name,
				Logo:      ft.Logo,
				League:    leagueID,
				IsPopular: ft.IsPopular,
				Seasons:   make(map[string]map[string]string),
			}
			store.Teams[teamID] = metadata.Team{
				ID:     teamID,
				Name:   ft.Name,
				League: leagueID,
				Logo:   ft.Logo,
			}
			importedCount++
		}

		// Add kits (Firestore data usually doesn't have season, so we put it in "Imported")
		season := "Imported"
		if team.Seasons == nil {
			team.Seasons = make(map[string]map[string]string)
		}
		if team.Seasons[season] == nil {
			team.Seasons[season] = make(map[string]string)
		}

		addKit := func(kType, url string) {
			if url != "" && team.Seasons[season][kType] == "" {
				team.Seasons[season][kType] = url
			}
		}

		addKit("home", ft.Home)
		addKit("away", ft.Away)
		addKit("third", ft.Third)
		addKit("gk_home", ft.GKHome)
		addKit("gk_away", ft.GKAway)
		addKit("gk_third", ft.GKThird)

		catalog.Teams[teamID] = team
	}

	// 5. Save updated catalog
	out, _ := json.MarshalIndent(catalog, "", "  ")
	os.WriteFile("kits.json", out, 0644)

	// 6. Save back to metadata store so scraper can use it
	metadata.Save(metadata.DefaultDataDir, store)

	fmt.Printf("Merged %d teams from Firestore into kits.json and Master List\n", importedCount)
	fmt.Printf("Total teams now in catalog: %d\n", len(catalog.Teams))
}
