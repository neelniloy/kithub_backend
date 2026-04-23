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
	matcher := metadata.NewMatcher(store)

	sources := []scraper.Source{
		{Name: "dlskits.com", BaseURL: "https://dlskits.com/"},
		{Name: "dlskiturl.com", BaseURL: "https://dlskiturl.com/"},
		{Name: "kitdls.net", BaseURL: "https://kitdls.net/"},
		{Name: "ftsdlskits.com", BaseURL: "https://ftsdlskits.com/"},
		{Name: "dlskitsurl.com", BaseURL: "https://dlskitsurl.com/"},
		{Name: "dreamleaguesoccerkits.com", BaseURL: "https://www.dreamleaguesoccerkits.com/"},
		{Name: "dreamkitsapp.com", BaseURL: "https://dreamkitsapp.com/"},
		{Name: "sakibpro.com", BaseURL: "https://sakibpro.com/"},
		{Name: "kuchalana.com", BaseURL: "https://www.kuchalana.com/", SearchParam: "q"},
		{Name: "dlsguide.com", BaseURL: "https://dlsguide.com/"},
		{Name: "dlskits.club", BaseURL: "https://dlskits.club/"},
	}

	httpClient := scraper.NewHTTPClient(20 * time.Second)
	collector := scraper.NewCollector(httpClient, scraper.Options{
		MaxConcurrency: 15,
		MaxArticles:    10000,
		Retries:        3,
	})

	var pages []scraper.ArticlePage
	if deepSearch {
		teamNames := make([]string, 0, len(store.Teams))
		for _, t := range store.Teams {
			teamNames = append(teamNames, t.Name)
		}
		pages, err = collector.Search(ctx, sources, teamNames)
	} else {
		pages, err = collector.Collect(ctx, sources)
	}
	if err != nil {
		log.Printf("completed with source errors: %v", err)
	}

	catalog := models.Catalog{
		Version:     1,
		LastUpdated: time.Now().UTC().Format(time.DateOnly),
		Leagues:     make(map[string]models.League),
		Teams:       make(map[string]models.Team),
	}

	// Initialize catalog
	for _, l := range store.Leagues {
		catalog.Leagues[l.ID] = models.League{
			ID:        l.ID,
			Name:      l.Name,
			Logo:      l.Logo,
			IsPopular: l.IsPopular,
		}
	}
	for teamID, t := range store.Teams {
		catalog.Teams[teamID] = models.Team{
			Name:      t.Name,
			Logo:      t.Logo,
			League:    t.League,
			IsPopular: t.IsPopular,
			Seasons:   make(map[string]map[string]string),
		}
	}

	var parsedRecords []models.KitRecord
	var parsedLogos []models.LogoRecord

	log.Printf("parsing %d pages...", len(pages))
	for _, page := range pages {
		records, logos := parser.ParseArticle(page, matcher)
		parsedRecords = append(parsedRecords, records...)
		parsedLogos = append(parsedLogos, logos...)
	}

	// EXTREME MODE: Skip validation for speed
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
	for _, t := range catalog.Teams {
		for _, s := range t.Seasons {
			kitCount += len(s)
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
	for id, t := range catalog.Teams {
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

	fmt.Printf("Generated kits.json with %d leagues, %d teams, %d kit URLs\n", len(catalog.Leagues), len(catalog.Teams), kitCount)
	
	teamsWithKits := 0
	for _, team := range catalog.Teams {
		if len(team.Seasons) > 0 {
			teamsWithKits++
		}
	}
	fmt.Printf("Coverage: %d/%d teams have kit data (%.1f%%)\n", teamsWithKits, len(catalog.Teams), float64(teamsWithKits)/float64(len(catalog.Teams))*100)
	return nil
}

func countTeamLogos(catalog models.Catalog) int {
	count := 0
	for _, team := range catalog.Teams {
		if team.Logo != "" {
			count++
		}
	}
	return count
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
