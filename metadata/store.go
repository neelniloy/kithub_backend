package metadata

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

const (
	DefaultDataDir = "data"
)

func Save(dir string, store Store) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.DateOnly)

	leagueFile := LeagueFile{Version: 1, LastUpdated: now, Leagues: store.Leagues}
	if err := writeJSON(filepath.Join(dir, "leagues.json"), leagueFile); err != nil {
		return err
	}

	teamFile := TeamFile{Version: 1, LastUpdated: now, Teams: store.Teams}
	if err := writeJSON(filepath.Join(dir, "teams.json"), teamFile); err != nil {
		return err
	}

	return ensureContributionFiles(dir)
}

func Load(dir string) (Store, error) {
	store := Store{
		Leagues: make(map[string]League),
		Teams:   make(map[string]Team),
	}

	leagueFile, err := readLeagueFile(filepath.Join(dir, "leagues.json"))
	if err != nil {
		return store, err
	}
	teamFile, err := readTeamFile(filepath.Join(dir, "teams.json"))
	if err != nil {
		return store, err
	}

	for id, league := range leagueFile.Leagues {
		store.Leagues[id] = league
	}
	for id, team := range teamFile.Teams {
		store.Teams[id] = team
	}

	contribLeagues, err := readLeagueFile(filepath.Join(dir, "contrib", "leagues.json"))
	if err == nil {
		for id, league := range contribLeagues.Leagues {
			store.Leagues[id] = mergeLeague(store.Leagues[id], league)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return store, err
	}

	contribTeams, err := readTeamFile(filepath.Join(dir, "contrib", "teams.json"))
	if err == nil {
		for id, team := range contribTeams.Teams {
			store.Teams[id] = mergeTeam(store.Teams[id], team)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return store, err
	}

	return store, nil
}

func ensureContributionFiles(dir string) error {
	contribDir := filepath.Join(dir, "contrib")
	if err := os.MkdirAll(contribDir, 0755); err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.DateOnly)

	leaguePath := filepath.Join(contribDir, "leagues.json")
	if _, err := os.Stat(leaguePath); errors.Is(err, os.ErrNotExist) {
		if err := writeJSON(leaguePath, LeagueFile{Version: 1, LastUpdated: now, Leagues: map[string]League{}}); err != nil {
			return err
		}
	}

	teamPath := filepath.Join(contribDir, "teams.json")
	if _, err := os.Stat(teamPath); errors.Is(err, os.ErrNotExist) {
		if err := writeJSON(teamPath, TeamFile{Version: 1, LastUpdated: now, Teams: map[string]Team{}}); err != nil {
			return err
		}
	}

	return nil
}

func readLeagueFile(path string) (LeagueFile, error) {
	var file LeagueFile
	if err := readJSON(path, &file); err != nil {
		return file, err
	}
	if file.Leagues == nil {
		file.Leagues = map[string]League{}
	}
	return file, nil
}

func readTeamFile(path string) (TeamFile, error) {
	var file TeamFile
	if err := readJSON(path, &file); err != nil {
		return file, err
	}
	if file.Teams == nil {
		file.Teams = map[string]Team{}
	}
	return file, nil
}

func writeJSON(path string, value any) error {
	out, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, out, 0644)
}

func readJSON(path string, target any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

func mergeLeague(base, override League) League {
	if override.ID != "" {
		base.ID = override.ID
	}
	if override.ExternalID != "" {
		base.ExternalID = override.ExternalID
	}
	if override.Name != "" {
		base.Name = override.Name
	}
	if override.Logo != "" {
		base.Logo = override.Logo
	}
	if override.IsPopular {
		base.IsPopular = true
	}
	base.Aliases = uniqueStrings(append(base.Aliases, override.Aliases...))
	if override.Source != "" {
		base.Source = override.Source
	}
	return base
}

func mergeTeam(base, override Team) Team {
	if override.ID != "" {
		base.ID = override.ID
	}
	if override.ExternalID != "" {
		base.ExternalID = override.ExternalID
	}
	if override.Name != "" {
		base.Name = override.Name
	}
	if override.Logo != "" {
		base.Logo = override.Logo
	}
	if override.League != "" {
		base.League = override.League
	}
	if override.IsPopular {
		base.IsPopular = true
	}
	base.Aliases = uniqueStrings(append(base.Aliases, override.Aliases...))
	if override.Source != "" {
		base.Source = override.Source
	}
	return base
}
