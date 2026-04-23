package metadata

import (
	"encoding/json"
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

	return nil
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

	return store, nil
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
