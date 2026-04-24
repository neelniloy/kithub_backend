package parser

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"kithub_backend/metadata"
	"kithub_backend/models"
	"kithub_backend/scraper"
	"kithub_backend/utils"
)

var rawPNGLinkRE = regexp.MustCompile(`(?i)https?://[^\s"'<>]+?\.png(?:\?[^\s"'<>]*)?`)

func ParseArticle(page scraper.ArticlePage, matcher *metadata.Matcher) ([]models.KitRecord, []models.LogoRecord) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(page.HTML))
	if err != nil {
		return nil, nil
	}

	title := page.Title
	if title == "" {
		title = strings.TrimSpace(doc.Find("title").First().Text())
	}

	// 1. Identify the "Primary Team" for this page (from the title)
	primaryMatch, hasPrimary := matcher.MatchTeam(title)
	if !hasPrimary {
		primaryMatch, hasPrimary = matcher.FallbackTeam(title)
	}

	articleSeason := utils.ExtractSeason(title)
	if articleSeason == "" {
		articleSeason = utils.ExtractSeason(page.HTML)
	}

	found := make(map[string]bool)
	foundLogos := make(map[string]bool)
	var records []models.KitRecord
	var logos []models.LogoRecord

	// Helper to process a URL with a matched team
	processURL := func(rawURL, contextText string, specificTeam *metadata.Match) {
		kitURL, ok := normalizePNGURL(page.URL, rawURL)
		if !ok || found[kitURL] {
			return
		}

		// Use specific team if found, otherwise use page primary team
		match := primaryMatch
		if specificTeam != nil {
			match = *specificTeam
		}

		// If we still have no team, we can't save it
		if match.Team.ID == "" {
			return
		}

		league := modelLeague(match.League)

		if looksLikeLogoAsset(kitURL, contextText) {
			if foundLogos[kitURL] {
				return
			}
			foundLogos[kitURL] = true
			logos = append(logos, models.LogoRecord{
				TeamID:     match.Team.ID,
				TeamName:   match.Team.Name,
				TeamLogo:   match.Team.Logo,
				League:     league,
				URL:        kitURL,
				Source:     page.Source,
				ArticleURL: page.URL,
			})
			return
		}

		season := utils.ExtractSeason(kitURL + " " + contextText)
		if season == "" {
			season = articleSeason
		}
		// Default to 2025 if still empty (most likely current season)
		if season == "" {
			season = "2025"
		}

		kitType := utils.DetectKitType(kitURL)
		if kitType == "unknown" {
			kitType = utils.DetectKitType(contextText)
		}

		found[kitURL] = true
		records = append(records, models.KitRecord{
			TeamID:      match.Team.ID,
			TeamName:    match.Team.Name,
			TeamLogo:    match.Team.Logo,
			TeamPopular: match.Team.IsPopular,
			League:      league,
			Season:      season,
			KitType:     kitType,
			URL:         kitURL,
			Source:      page.Source,
			ArticleURL:  page.URL,
		})
	}

	// 2. Scan structured tags (images and links)
	doc.Find("img[src], img[data-src], img[data-lazy-src], a[href]").Each(func(_ int, s *goquery.Selection) {
		contextText := surroundingText(s)
		
		// Try to find a specific team in the surrounding text (for league posts)
		var specificMatch *metadata.Match
		if m, ok := matcher.MatchTeam(contextText); ok {
			specificMatch = &m
		} else if m, ok := matcher.FallbackTeam(contextText); ok {
			// Create a dynamic team from the context if not in master list
			specificMatch = &m
		}

		rawURL := ""
		for _, attr := range []string{"src", "data-src", "data-lazy-src", "href"} {
			if val, exists := s.Attr(attr); exists {
				rawURL = val
				break
			}
		}

		processURL(rawURL, contextText, specificMatch)
	})

	// 3. Scan raw HTML for any missed links (fallback to primary team)
	for _, rawURL := range rawPNGLinkRE.FindAllString(page.HTML, -1) {
		processURL(rawURL, title, nil)
	}

	return records, logos
}

func modelLeague(league metadata.League) models.League {
	if league.ID == "" {
		return models.League{ID: "unknown", Name: "Unknown"}
	}
	return models.League{
		ID:        league.ID,
		Name:      league.Name,
		Logo:      league.Logo,
		IsPopular: league.IsPopular,
	}
}

func normalizePNGURL(articleURL, rawURL string) (string, bool) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "", false
	}
	rawURL = strings.Trim(rawURL, `"'`)
	rawURL = strings.ReplaceAll(rawURL, `\/`, `/`)

	base, err := url.Parse(articleURL)
	if err != nil {
		return "", false
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", false
	}
	resolved := base.ResolveReference(parsed)
	if resolved.Scheme != "http" && resolved.Scheme != "https" {
		return "", false
	}
	if !strings.HasSuffix(strings.ToLower(resolved.Path), ".png") {
		return "", false
	}
	resolved.Fragment = ""
	return resolved.String(), true
}

func surroundingText(s *goquery.Selection) string {
	parts := []string{
		s.AttrOr("alt", ""),
		s.AttrOr("title", ""),
		s.Parent().Text(),
		s.Parent().Prev().Text(),
		s.Parent().Next().Text(),
	}
	return strings.Join(parts, " ")
}



func looksLikeLogoAsset(kitURL, _ string) bool {
	assetText := strings.ToLower(kitURL)
	return strings.Contains(assetText, "logo") ||
		strings.Contains(assetText, "icon") ||
		strings.Contains(assetText, "512x512") ||
		strings.Contains(assetText, "512-x-512")
}

func logoBelongsToTeam(logoURL, teamID, teamName string) bool {
	urlText := normalizeComparableText(logoURL)
	return strings.Contains(urlText, normalizeComparableText(teamID)) ||
		strings.Contains(urlText, normalizeComparableText(teamName))
}

func normalizeComparableText(value string) string {
	value = strings.ToLower(value)
	replacer := strings.NewReplacer("_", "", "-", "", " ", "", ".", "", "/", "", "%20", "")
	return replacer.Replace(value)
}
