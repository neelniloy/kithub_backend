package scraper

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Source struct {
	Name        string
	BaseURL     string
	SearchParam string // "s" for WordPress, "q" for Blogger
}

type ArticlePage struct {
	Source     string
	URL        string
	Title      string
	HTML       string
	ResolvedAt time.Time
}

type Options struct {
	MaxConcurrency int
	MaxArticles    int
	Retries        int
}

type Collector struct {
	http *HTTPClient
	opts Options
}

func NewCollector(httpClient *HTTPClient, opts Options) *Collector {
	if opts.MaxConcurrency <= 0 {
		opts.MaxConcurrency = 5
	}
	if opts.MaxConcurrency > 10 {
		opts.MaxConcurrency = 10
	}
	if opts.MaxArticles <= 0 {
		opts.MaxArticles = 50
	}
	if opts.Retries < 0 {
		opts.Retries = 0
	}
	return &Collector{http: httpClient, opts: opts}
}

func (c *Collector) Collect(ctx context.Context, sources []Source) ([]ArticlePage, error) {
	articleURLs := make(map[string]Source)
	var articleMu sync.Mutex
	var discoveryErrs []error
	var discoveryMu sync.Mutex

	var wgDiscovery sync.WaitGroup
	for _, source := range sources {
		wgDiscovery.Add(1)
		go func(s Source) {
			defer wgDiscovery.Done()
			log.Printf("Starting discovery for source: %s", s.Name)
			
			currentURL := s.BaseURL
			foundForSource := 0
			page := 1

			for foundForSource < c.opts.MaxArticles && page <= 50 {
				html, err := c.getWithRetry(ctx, currentURL)
				if err != nil {
					discoveryMu.Lock()
					discoveryErrs = append(discoveryErrs, fmt.Errorf("%s discovery at %s: %w", s.Name, currentURL, err))
					discoveryMu.Unlock()
					break
				}

				urls, nextURL, err := discoverArticleURLs(s, html, c.opts.MaxArticles-foundForSource)
				if err != nil {
					discoveryMu.Lock()
					discoveryErrs = append(discoveryErrs, fmt.Errorf("%s parse links: %w", s.Name, err))
					discoveryMu.Unlock()
					break
				}

				articleMu.Lock()
				for _, articleURL := range urls {
					if _, exists := articleURLs[articleURL]; !exists {
						articleURLs[articleURL] = s
						foundForSource++
					}
				}
				articleMu.Unlock()

				if foundForSource >= c.opts.MaxArticles {
					break
				}

				if nextURL == "" {
					page++
					nextURL = fmt.Sprintf("%s/page/%d/", strings.TrimRight(s.BaseURL, "/"), page)
				} else {
					page++
				}
				currentURL = nextURL
			}
			log.Printf("Finished discovery for %s: found %d articles", s.Name, foundForSource)
		}(source)
	}
	wgDiscovery.Wait()

	jobs := make(chan string)
	results := make(chan ArticlePage)
	var wg sync.WaitGroup

	for i := 0; i < c.opts.MaxConcurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for articleURL := range jobs {
				source := articleURLs[articleURL]
				html, err := c.getWithRetry(ctx, articleURL)
				if err != nil {
					log.Printf("skip article %s: %v", articleURL, err)
					continue
				}
				results <- ArticlePage{
					Source:     source.Name,
					URL:        articleURL,
					Title:      extractTitle(html),
					HTML:       html,
					ResolvedAt: time.Now().UTC(),
				}
			}
		}()
	}

	go func() {
		defer close(jobs)
		urls := make([]string, 0, len(articleURLs))
		for articleURL := range articleURLs {
			urls = append(urls, articleURL)
		}
		sort.Strings(urls)
		for _, articleURL := range urls {
			select {
			case <-ctx.Done():
				return
			case jobs <- articleURL:
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	pages := make([]ArticlePage, 0, len(articleURLs))
	for page := range results {
		pages = append(pages, page)
	}

	if len(discoveryErrs) > 0 {
		return pages, errors.Join(discoveryErrs...)
	}
	return pages, nil
}

func (c *Collector) Search(ctx context.Context, sources []Source, teamNames []string) ([]ArticlePage, error) {
	articleURLs := make(map[string]Source)
	var articleMu sync.Mutex

	type searchJob struct {
		source Source
		team   string
	}

	jobs := make(chan searchJob)
	var searchWg sync.WaitGroup

	log.Printf("Starting targeted search discovery for %d teams across %d sources (Rate Limited)...", len(teamNames), len(sources))
	
	// Use a small number of workers for searching to avoid instant bans
	for i := 0; i < 5; i++ {
		searchWg.Add(1)
		go func() {
			defer searchWg.Done()
			for job := range jobs {
				param := job.source.SearchParam
				if param == "" {
					param = "s"
				}
				searchURL := fmt.Sprintf("%s?%s=%s", strings.TrimRight(job.source.BaseURL, "/"), param, url.QueryEscape(job.team))
				html, err := c.getWithRetry(ctx, searchURL)
				if err != nil {
					continue
				}

				urls, _, _ := discoverArticleURLs(job.source, html, 3)
				if len(urls) > 0 {
					log.Printf("Found %d articles for %s on %s", len(urls), job.team, job.source.Name)
				}
				
				articleMu.Lock()
				for _, u := range urls {
					articleURLs[u] = job.source
				}
				articleMu.Unlock()
				
				// Small delay to be polite
				time.Sleep(200 * time.Millisecond)
			}
		}()
	}

	go func() {
		for _, s := range sources {
			for _, team := range teamNames {
				jobs <- searchJob{s, team}
			}
		}
		close(jobs)
	}()

	searchWg.Wait()

	log.Printf("Targeted search found %d potential articles. Starting parallel crawl...", len(articleURLs))

	crawlJobs := make(chan string)
	results := make(chan ArticlePage)
	var workerWg sync.WaitGroup

	for i := 0; i < c.opts.MaxConcurrency; i++ {
		workerWg.Add(1)
		go func() {
			defer workerWg.Done()
			for url := range crawlJobs {
				html, err := c.getWithRetry(ctx, url)
				if err != nil {
					continue
				}
				results <- ArticlePage{
					URL:    url,
					Source: articleURLs[url].Name,
					HTML:   html,
				}
			}
		}()
	}

	go func() {
		for url := range articleURLs {
			crawlJobs <- url
		}
		close(crawlJobs)
	}()

	go func() {
		workerWg.Wait()
		close(results)
	}()

	var pages []ArticlePage
	for page := range results {
		pages = append(pages, page)
	}

	return pages, nil
}

func (c *Collector) getWithRetry(ctx context.Context, rawURL string) (string, error) {
	var lastErr error
	for attempt := 0; attempt <= c.opts.Retries; attempt++ {
		html, _, err := c.http.Get(ctx, rawURL)
		if err == nil {
			return html, nil
		}
		lastErr = err
		if attempt < c.opts.Retries {
			time.Sleep(time.Duration(attempt+1) * 700 * time.Millisecond)
		}
	}
	return "", lastErr
}

func discoverArticleURLs(source Source, html string, maxArticles int) ([]string, string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, "", err
	}

	base, err := url.Parse(source.BaseURL)
	if err != nil {
		return nil, "", err
	}

	seen := make(map[string]bool)
	var urls []string
	doc.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
		if len(urls) >= maxArticles {
			return
		}

		href, ok := s.Attr("href")
		if !ok {
			return
		}
		articleURL, ok := normalizeInternalArticleURL(base, href)
		if !ok || seen[articleURL] {
			return
		}

		text := strings.ToLower(strings.TrimSpace(s.Text()))
		
		// HYPER-AGGRESSIVE: If we're on a search page, we follow anything that looks like a kit page
		isKitLink := strings.Contains(text, "kit") || strings.Contains(text, "dls") || strings.Contains(text, "202") || strings.Contains(articleURL, "kit")
		
		if !isKitLink && !looksLikeArticleLink(articleURL, text) {
			return
		}

		seen[articleURL] = true
		urls = append(urls, articleURL)
	})

	// Find Next Page URL
	nextURL := ""
	doc.Find("a.next, a.next-page, a.page-numbers.next, .pagination a").EachWithBreak(func(_ int, s *goquery.Selection) bool {
		text := strings.ToLower(s.Text())
		if strings.Contains(text, "next") || strings.Contains(text, "»") || strings.Contains(text, ">") {
			if href, ok := s.Attr("href"); ok {
				if absolute, ok := normalizeInternalArticleURL(base, href); ok {
					nextURL = absolute
					return false
				}
			}
		}
		return true
	})

	return urls, nextURL, nil
}

func normalizeInternalArticleURL(base *url.URL, href string) (string, bool) {
	href = strings.TrimSpace(href)
	if href == "" || strings.HasPrefix(href, "#") || strings.HasPrefix(strings.ToLower(href), "javascript:") {
		return "", false
	}

	parsed, err := url.Parse(href)
	if err != nil {
		return "", false
	}

	resolved := base.ResolveReference(parsed)
	if resolved.Hostname() != base.Hostname() {
		return "", false
	}
	if resolved.Scheme != "http" && resolved.Scheme != "https" {
		return "", false
	}

	resolved.Fragment = ""
	resolved.RawQuery = ""
	path := strings.TrimRight(resolved.EscapedPath(), "/")
	if path == "" {
		return "", false
	}

	lowerPath := strings.ToLower(path)
	blocked := []string{
		"/privacy", "/contact", "/about", "/dmca", "/disclaimer", "/terms", "/search",
		"/author", "/wp-content", "/feed",
	}
	for _, prefix := range blocked {
		if strings.HasPrefix(lowerPath, prefix) {
			return "", false
		}
	}

	resolved.Path = path + "/"
	return resolved.String(), true
}

func looksLikeArticleLink(articleURL, text string) bool {
	combined := strings.ToLower(articleURL + " " + text)
	if strings.Contains(combined, ".png") || strings.Contains(combined, ".jpg") || strings.Contains(combined, ".jpeg") {
		return false
	}
	// Many team kit pages are just the team name (e.g. dlskits.com/ac-milan/)
	// If it's an internal link and not a known junk path (handled in normalize), it's probably worth checking.
	return true
}

func extractTitle(html string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return ""
	}
	for _, selector := range []string{"article h1", "h1.entry-title", "h1.post-title", "h1"} {
		title := strings.TrimSpace(doc.Find(selector).First().Text())
		if title != "" {
			return title
		}
	}
	return strings.TrimSpace(doc.Find("title").First().Text())
}
