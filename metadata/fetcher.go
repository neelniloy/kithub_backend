package metadata

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const sportsDBBaseURL = "https://www.thesportsdb.com/api/v1/json/3"

type Fetcher struct {
	client *http.Client
}

func NewFetcher(timeout time.Duration) *Fetcher {
	return &Fetcher{client: &http.Client{Timeout: timeout}}
}

type sportsDBLeagueResponse struct {
	Leagues []sportsDBLeague `json:"leagues"`
}

type sportsDBTeamsResponse struct {
	Teams []sportsDBTeam `json:"teams"`
}

type sportsDBLeague struct {
	ID        string `json:"idLeague"`
	Name      string `json:"strLeague"`
	Alternate string `json:"strLeagueAlternate"`
	Badge     string `json:"strBadge"`
	Logo      string `json:"strLogo"`
}

type sportsDBTeam struct {
	ID        string `json:"idTeam"`
	Name      string `json:"strTeam"`
	Alternate string `json:"strAlternate"`
	ShortName string `json:"strTeamShort"`
	Badge     string `json:"strBadge"`
	Logo      string `json:"strLogo"`
}

func (f *Fetcher) Fetch(ctx context.Context) (Store, error) {
	store := Store{
		Leagues: make(map[string]League),
		Teams:   make(map[string]Team),
	}

	var fetchErr error
	for _, source := range SourceLeagues {
		league, err := f.fetchLeague(ctx, source)
		if err != nil {
			fetchErr = err
			continue
		}
		store.Leagues[league.ID] = league

		teams, err := f.fetchTeams(ctx, source, league.ID)
		if err != nil {
			fetchErr = err
			continue
		}
		for _, team := range teams {
			store.Teams[team.ID] = team
		}
	}

	store.Leagues[internationalLeague.ID] = internationalLeague
	for _, team := range nationalTeams {
		store.Teams[team.ID] = team
	}

	return store, fetchErr
}

func (f *Fetcher) fetchLeague(ctx context.Context, source SourceLeague) (League, error) {
	var response sportsDBLeagueResponse
	if err := f.getJSON(ctx, fmt.Sprintf("%s/lookupleague.php?id=%s", sportsDBBaseURL, source.ExternalID), &response); err != nil {
		return League{}, err
	}
	if len(response.Leagues) == 0 {
		return League{}, fmt.Errorf("league %s not found", source.ExternalID)
	}

	league := response.Leagues[0]
	return League{
		ID:         source.ID,
		ExternalID: league.ID,
		Name:       cleanLeagueName(league.Name),
		Logo:       firstNonEmpty(league.Badge, league.Logo),
		IsPopular:  source.IsPopular,
		Aliases:    uniqueStrings(append(source.Aliases, splitAliases(league.Alternate)...)),
		Source:     "thesportsdb",
	}, nil
}

func (f *Fetcher) fetchTeams(ctx context.Context, source SourceLeague, leagueID string) ([]Team, error) {
	var response sportsDBTeamsResponse
	searchName := source.SearchName
	if searchName == "" {
		searchName = source.ID
	}
	if err := f.getJSON(ctx, fmt.Sprintf("%s/search_all_teams.php?l=%s", sportsDBBaseURL, url.QueryEscape(searchName)), &response); err != nil {
		return nil, err
	}

	teams := make([]Team, 0, len(response.Teams))
	for _, item := range response.Teams {
		if item.Name == "" {
			continue
		}
		aliases := []string{item.ShortName}
		aliases = append(aliases, splitAliases(item.Alternate)...)
		teamID := teamIDForName(item.Name)
		aliases = append(aliases, extraAliasesForTeam(teamID)...)
		teams = append(teams, Team{
			ID:         teamID,
			ExternalID: item.ID,
			Name:       item.Name,
			Logo:       firstNonEmpty(item.Badge, item.Logo),
			League:     leagueID,
			IsPopular:  isKnownPopularTeam(item.Name),
			Aliases:    uniqueStrings(aliases),
			Source:     "thesportsdb",
		})
	}

	return teams, nil
}

func (f *Fetcher) getJSON(ctx context.Context, rawURL string, target any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "DLSKitScraper/1.0")
	req.Header.Set("Accept", "application/json")

	resp, err := f.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s returned %d", rawURL, resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(target)
}

func cleanLeagueName(name string) string {
	switch name {
	case "English Premier League":
		return "Premier League"
	case "Spanish La Liga":
		return "La Liga"
	case "German Bundesliga":
		return "Bundesliga"
	case "Italian Serie A":
		return "Serie A"
	case "French Ligue 1":
		return "Ligue 1"
	case "American Major League Soccer":
		return "Major League Soccer"
	case "Brazilian Serie A":
		return "Brasileirao"
	case "Argentinian Primera Division":
		return "Argentine Primera Division"
	case "Peruvian Primera Division":
		return "Liga 1 Peru"
	default:
		return name
	}
}

func extraAliasesForTeam(teamID string) []string {
	switch teamID {
	case "atletico_madrid":
		return []string{"Atletico Madrid", "Atletico de Madrid", "Atlético de Madrid"}
	case "barcelona":
		return []string{"FC Barcelona", "Barca", "F.C. Barcelona"}
	case "real_madrid":
		return []string{"Real Madird", "R Madrid"}
	case "manchester_united":
		return []string{"Man United", "Man Utd"}
	case "manchester_city":
		return []string{"Man City"}
	case "paris_saint_germain":
		return []string{"PSG", "Paris SG"}
	case "inter_milan":
		return []string{"Inter Milan", "Internazionale"}
	case "talleres_de_cordoba":
		return []string{"Talleres", "CA Talleres"}
	case "tottenham_hotspur":
		return []string{"Tottenham", "Tottenham Premier"}
	case "atletico_tucuman":
		return []string{"Atlético Tucumán", "Atletico Tucuman"}
	case "santos":
		return []string{"Santos FC"}
	default:
		return nil
	}
}

func teamIDForName(name string) string {
	switch NormalizeAlias(name) {
	case "fc barcelona":
		return "barcelona"
	case "real madrid":
		return "real_madrid"
	case "manchester united":
		return "manchester_united"
	case "manchester city":
		return "manchester_city"
	case "tottenham hotspur":
		return "tottenham_hotspur"
	case "wolverhampton wanderers":
		return "wolves"
	case "paris saint germain":
		return "paris_saint_germain"
	case "internazionale":
		return "inter_milan"
	case "ac milan":
		return "ac_milan"
	case "bayer leverkusen":
		return "bayer_leverkusen"
	case "paris sg":
		return "paris_saint_germain"
	default:
		return Slug(name)
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func isKnownPopularTeam(name string) bool {
	switch NormalizeAlias(name) {
	case "fc barcelona", "real madrid", "manchester united", "manchester city", "liverpool",
		"chelsea", "arsenal", "bayern munich", "borussia dortmund", "paris saint germain",
		"juventus", "internazionale", "ac milan", "al nassr", "inter miami", "galatasaray",
		"boca juniors", "argentina", "brazil", "france", "portugal", "mexico":
		return true
	default:
		return false
	}
}
