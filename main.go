package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"kithub_backend/metadata"
	"kithub_backend/models"
	"kithub_backend/parser"
	"kithub_backend/scraper"
	"kithub_backend/utils"
)

func main() {
	ctx := context.Background()

	command := "scrape"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	switch command {
	case "sync-meta":
		if err := syncMetadata(ctx); err != nil {
			log.Fatalf("sync metadata: %v", err)
		}
		return
	case "scrape":
		if err := scrape(ctx, false); err != nil {
			log.Fatalf("scrape: %v", err)
		}
	case "deep-search":
		if err := scrape(ctx, true); err != nil {
			log.Fatalf("deep-search: %v", err)
		}
	default:
		log.Fatalf("unknown command %q; use sync-meta, scrape, or deep-search", command)
	}
}

func syncMetadata(ctx context.Context) error {
	fetcher := metadata.NewFetcher(25 * time.Second)
	store, err := fetcher.Fetch(ctx)
	if saveErr := metadata.Save(metadata.DefaultDataDir, store); saveErr != nil {
		return saveErr
	}
	if err != nil {
		log.Printf("metadata sync completed with partial source errors: %v", err)
	}
	fmt.Printf("Synced metadata: %d leagues and %d teams\n", len(store.Leagues), len(store.Teams))
	return nil
}

func scrape(ctx context.Context, deepSearch bool) error {
	store, err := metadata.Load(metadata.DefaultDataDir)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		log.Printf("metadata files not found; running one-time metadata sync")
		if err := syncMetadata(ctx); err != nil {
			return err
		}
		store, err = metadata.Load(metadata.DefaultDataDir)
		if err != nil {
			return err
		}
	}

	// ALWAYS merge Firestore teams to ensure full coverage (211+ teams)
	firestorePath := filepath.Join("temp", "_Team.json")
	if _, err := os.Stat(firestorePath); err == nil {
		log.Printf("merging existing teams from %s...", firestorePath)
		if err := store.MergeFirestore(firestorePath); err != nil {
			log.Printf("Warning: failed to merge firestore teams: %v", err)
		} else {
			log.Printf("Metadata store now has %d teams", len(store.Teams))
		}
	}

	matcher := metadata.NewMatcher(store)

	discoverySources := []scraper.Source{
		{Name: "DLSKits", BaseURL: "https://dlskits.com/"},
		{Name: "DLSKits-2026", BaseURL: "https://dlskits.com/category/dls-26-kits/"},
		{Name: "KitDLS", BaseURL: "https://kitdls.net/"},
		{Name: "FTSDLSKits", BaseURL: "https://ftsdlskits.com/"},
		{Name: "FTSDLSKits-2026", BaseURL: "https://ftsdlskits.com/category/dream-league-soccer-kits-2026/"},
		{Name: "DreamKits", BaseURL: "https://img.dreamkitsapp.com/"},
		{Name: "DLSKitsURL", BaseURL: "https://dlskitsurl.com/"},
		{Name: "dreamleaguesoccerkits.com", BaseURL: "https://www.dreamleaguesoccerkits.com/"},
		{Name: "dreamkitsapp.com", BaseURL: "https://dreamkitsapp.com/"},
		{Name: "sakibpro.com", BaseURL: "https://sakibpro.com/"},
		{Name: "kuchalana.com", BaseURL: "https://www.kuchalana.com/", SearchParam: "q"},
		{Name: "dlsguide.com", BaseURL: "https://dlsguide.com/"},
		{Name: "DLSKits.club", BaseURL: "https://dlskits.club/"},
		{Name: "DLSKitsHub", BaseURL: "https://dlskitshub.com/"},
		{Name: "DLSKitsHub-2026", BaseURL: "https://dlskitshub.com/category/dream-league-soccer-kits-2026/"},
	}

	httpClient := scraper.NewHTTPClient(20 * time.Second)
	collector := scraper.NewCollector(httpClient, scraper.Options{
		MaxConcurrency: 15,
		MaxArticles:    10000,
		Retries:        3,
	})

	var pages []scraper.ArticlePage
	// HYBRID MODE: Always do general discovery AND targeted search for best results
	log.Printf("starting general discovery (latest posts)...")
	generalPages, genErr := collector.Collect(ctx, discoverySources)
	if genErr != nil {
		log.Printf("general discovery partial error: %v", genErr)
	}
	pages = append(pages, generalPages...)

	log.Printf("starting targeted team-wise scraping for %d known teams...", len(store.Teams))
	teamNames := make([]string, 0, len(store.Teams))
	for _, t := range store.Teams {
		teamNames = append(teamNames, t.Name)
	}
	searchPages, searchErr := collector.Search(ctx, discoverySources, teamNames)
	if searchErr != nil {
		log.Printf("targeted search partial error: %v", searchErr)
	}
	pages = append(pages, searchPages...)
	
	if genErr != nil || searchErr != nil {
		log.Printf("completed with some source errors")
	}

	catalog := models.Catalog{
		Version:     1,
		LastUpdated: time.Now().UTC().Format(time.DateOnly),
		Leagues:     make(map[string]models.League),
		Teams:       make(map[string]models.Team),
	}

	// Initialize leagues
	for _, l := range store.Leagues {
		catalog.Leagues[l.ID] = models.League{
			ID:        l.ID,
			Name:      l.Name,
			Logo:      l.Logo,
			IsPopular: l.IsPopular,
		}
	}

	var parsedRecords []models.KitRecord
	var parsedLogos []models.LogoRecord

	// IMPORT EXISTING KITS FROM FIRESTORE (into 2024 season by default)
	if _, err := os.Stat(firestorePath); err == nil {
		data, _ := os.ReadFile(firestorePath)
		var fTeams []metadata.FirestoreTeam
		if err := json.Unmarshal(data, &fTeams); err == nil {
			for _, ft := range fTeams {
				teamID := metadata.Slug(ft.Name)
				add := func(kType, url string) {
					if url != "" {
						parsedRecords = append(parsedRecords, models.KitRecord{
							TeamID:   teamID,
							TeamName: ft.Name,
							Season:   "2024", // Default for existing data (2024-25 season)
							KitType:  kType,
							URL:      url,
							Source:   "firestore",
							League:   models.League{ID: metadata.Slug(ft.League), Name: ft.League},
						})
					}
				}
				add("home", ft.Home)
				add("away", ft.Away)
				add("third", ft.Third)
				add("gk_home", ft.GKHome)
				add("gk_away", ft.GKAway)
				add("gk_third", ft.GKThird)
			}
		}
	}

	log.Printf("parsing %d scraped pages...", len(pages))
	for _, page := range pages {
		records, logos := parser.ParseArticle(page, matcher)
		parsedRecords = append(parsedRecords, records...)
		parsedLogos = append(parsedLogos, logos...)
	}

	// Add kits to catalog
	log.Printf("found %d potential kits; saving...", len(parsedRecords))
	for _, record := range parsedRecords {
		utils.AddKitRecord(&catalog, record)
	}

	logoMap := make(map[string]string)
	for _, l := range parsedLogos {
		if logoMap[l.TeamID] == "" {
			logoMap[l.TeamID] = l.URL
		}
	}
	utils.ApplyTeamLogos(&catalog, logoMap)

	out, err := json.MarshalIndent(catalog, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal catalog: %w", err)
	}

	if err := os.WriteFile("kits.json", out, 0644); err != nil {
		return fmt.Errorf("write kits.json: %w", err)
	}

	kitCount := 0
	teamsWithKits := 0
	for _, t := range catalog.Teams {
		hasKits := false
		for _, s := range t.Seasons {
			if len(s) > 0 {
				kitCount += len(s)
				hasKits = true
			}
		}
		if hasKits {
			teamsWithKits++
		}
	}

	// Save back to metadata files to make discovered teams permanent
	log.Printf("saving discovered metadata to data/...")
	newLeaguesFile := metadata.LeagueFile{
		Version:     1,
		LastUpdated: time.Now().UTC().Format(time.RFC3339),
		Leagues:     make(map[string]metadata.League),
	}
	for id, l := range catalog.Leagues {
		newLeaguesFile.Leagues[id] = metadata.League{
			ID:        l.ID,
			Name:      l.Name,
			Logo:      l.Logo,
			IsPopular: l.IsPopular,
		}
	}
	leaguesData, _ := json.MarshalIndent(newLeaguesFile, "", "  ")
	os.WriteFile(filepath.Join("data", "leagues.json"), leaguesData, 0644)

	newTeamsFile := metadata.TeamFile{
		Version:     1,
		LastUpdated: time.Now().UTC().Format(time.RFC3339),
		Teams:       make(map[string]metadata.Team),
	}
	// Use store.Teams as base to preserve all teams
	for id, t := range store.Teams {
		newTeamsFile.Teams[id] = metadata.Team{
			ID:        id,
			Name:      t.Name,
			Logo:      t.Logo,
			League:    t.League,
			IsPopular: t.IsPopular,
		}
	}
	teamsData, _ := json.MarshalIndent(newTeamsFile, "", "  ")
	os.WriteFile(filepath.Join("data", "teams.json"), teamsData, 0644)

	fmt.Printf("Generated kits.json with %d leagues, %d teams with kits, %d kit URLs\n", len(catalog.Leagues), teamsWithKits, kitCount)
	fmt.Printf("Coverage: %d/%d teams in master list have kit data (%.1f%%)\n", teamsWithKits, len(store.Teams), float64(teamsWithKits)/float64(len(store.Teams))*100)
	return nil
}



func validateKitURLs(ctx context.Context, client *scraper.HTTPClient, records []models.KitRecord, maxConcurrency int) []models.KitRecord {
	if maxConcurrency <= 0 {
		maxConcurrency = 5
	}
	if maxConcurrency > 10 {
		maxConcurrency = 10
	}

	jobs := make(chan models.KitRecord)
	results := make(chan models.KitRecord)
	var wg sync.WaitGroup

	for i := 0; i < maxConcurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for record := range jobs {
				if client.ValidPNG(ctx, record.URL) {
					results <- record
				}
			}
		}()
	}

	go func() {
		defer close(jobs)
		for _, record := range records {
			select {
			case <-ctx.Done():
				return
			case jobs <- record:
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	valid := make([]models.KitRecord, 0, len(records))
	for record := range results {
		valid = append(valid, record)
	}

	return valid
}

func validateLogoURLs(ctx context.Context, client *scraper.HTTPClient, logos []models.LogoRecord, maxConcurrency int) map[string]string {
	if maxConcurrency <= 0 {
		maxConcurrency = 5
	}
	if maxConcurrency > 10 {
		maxConcurrency = 10
	}

	jobs := make(chan models.LogoRecord)
	results := make(chan models.LogoRecord)
	var wg sync.WaitGroup

	for i := 0; i < maxConcurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for logo := range jobs {
				if client.ValidPNG(ctx, logo.URL) {
					results <- logo
				}
			}
		}()
	}

	go func() {
		defer close(jobs)
		for _, logo := range logos {
			select {
			case <-ctx.Done():
				return
			case jobs <- logo:
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	valid := make(map[string]string)
	for logo := range results {
		if _, exists := valid[logo.TeamID]; !exists {
			valid[logo.TeamID] = logo.URL
		}
	}

	return valid
}
