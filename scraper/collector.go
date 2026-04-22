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
	Name    string
	BaseURL string
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
	var discoveryErrs []error

	for _, source := range sources {
		html, err := c.getWithRetry(ctx, source.BaseURL)
		if err != nil {
			discoveryErrs = append(discoveryErrs, fmt.Errorf("%s discovery: %w", source.Name, err))
			continue
		}

		urls, err := discoverArticleURLs(source, html, c.opts.MaxArticles)
		if err != nil {
			discoveryErrs = append(discoveryErrs, fmt.Errorf("%s parse links: %w", source.Name, err))
			continue
		}

		for _, articleURL := range urls {
			articleURLs[articleURL] = source
		}
	}

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

func discoverArticleURLs(source Source, html string, maxArticles int) ([]string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	base, err := url.Parse(source.BaseURL)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	var urls []string
	doc.Find("article a[href], h1 a[href], h2 a[href], h3 a[href], .post a[href], .entry-title a[href], a[href]").Each(func(_ int, s *goquery.Selection) {
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
		if !looksLikeArticleLink(articleURL, text) {
			return
		}

		seen[articleURL] = true
		urls = append(urls, articleURL)
	})

	return urls, nil
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
		"/category", "/tag", "/page", "/privacy", "/contact", "/about", "/dmca",
		"/disclaimer", "/terms", "/search", "/author", "/wp-content", "/feed",
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
	if strings.Contains(combined, ".png") {
		return false
	}
	keywords := []string{"kit", "kits", "dls", "dream-league", "dream league", "logo"}
	for _, keyword := range keywords {
		if strings.Contains(combined, keyword) {
			return true
		}
	}
	return false
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
