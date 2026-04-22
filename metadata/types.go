package metadata

type League struct {
	ID         string   `json:"id"`
	ExternalID string   `json:"external_id,omitempty"`
	Name       string   `json:"name"`
	Logo       string   `json:"logo,omitempty"`
	IsPopular  bool     `json:"is_popular,omitempty"`
	Aliases    []string `json:"aliases,omitempty"`
	Source     string   `json:"source,omitempty"`
}

type Team struct {
	ID         string   `json:"id"`
	ExternalID string   `json:"external_id,omitempty"`
	Name       string   `json:"name"`
	Logo       string   `json:"logo,omitempty"`
	League     string   `json:"league"`
	IsPopular  bool     `json:"is_popular,omitempty"`
	Aliases    []string `json:"aliases,omitempty"`
	Source     string   `json:"source,omitempty"`
}

type LeagueFile struct {
	Version     int               `json:"version"`
	LastUpdated string            `json:"last_updated"`
	Leagues     map[string]League `json:"leagues"`
}

type TeamFile struct {
	Version     int             `json:"version"`
	LastUpdated string          `json:"last_updated"`
	Teams       map[string]Team `json:"teams"`
}

type Store struct {
	Leagues map[string]League
	Teams   map[string]Team
}

type Match struct {
	Team   Team
	League League
}
