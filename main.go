package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
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
		if err := scrape(ctx); err != nil {
			log.Fatalf("scrape: %v", err)
		}
	default:
		log.Fatalf("unknown command %q; use sync-meta or scrape", command)
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

func scrape(ctx context.Context) error {
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
	}

	httpClient := scraper.NewHTTPClient(20 * time.Second)
	collector := scraper.NewCollector(httpClient, scraper.Options{
		MaxConcurrency: 8,
		MaxArticles:    80,
		Retries:        2,
	})

	pages, err := collector.Collect(ctx, sources)
	if err != nil {
		log.Printf("completed with source errors: %v", err)
	}

	catalog := models.Catalog{
		Version:     1,
		LastUpdated: time.Now().UTC().Format(time.DateOnly),
		Leagues:     make(map[string]models.League),
		Teams:       make(map[string]models.Team),
	}

	parsedRecords := make([]models.KitRecord, 0)
	parsedLogos := make([]models.LogoRecord, 0)
	seenURLs := make(map[string]bool)
	seenLogos := make(map[string]bool)
	for _, page := range pages {
		records, logos := parser.ParseArticle(page, matcher)
		for _, record := range records {
			if seenURLs[record.URL] {
				continue
			}
			seenURLs[record.URL] = true
			parsedRecords = append(parsedRecords, record)
		}
		for _, logo := range logos {
			if seenLogos[logo.URL] {
				continue
			}
			seenLogos[logo.URL] = true
			parsedLogos = append(parsedLogos, logo)
		}
	}

	validRecords := validateKitURLs(ctx, httpClient, parsedRecords, 8)
	for _, record := range validRecords {
		utils.AddKitRecord(&catalog, record)
	}

	validLogos := validateLogoURLs(ctx, httpClient, parsedLogos, 8)
	utils.ApplyTeamLogos(&catalog, validLogos)

	out, err := json.MarshalIndent(catalog, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal catalog: %w", err)
	}

	if err := os.WriteFile("kits.json", out, 0644); err != nil {
		return fmt.Errorf("write kits.json: %w", err)
	}

	fmt.Printf("Generated kits.json with %d leagues, %d teams, %d valid kit URLs, and %d team logos\n", len(catalog.Leagues), len(catalog.Teams), len(validRecords), countTeamLogos(catalog))
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
