package models

type Catalog struct {
	Version     int               `json:"version"`
	LastUpdated string            `json:"last_updated"`
	Leagues     map[string]League `json:"leagues"`
	Seasons     map[string]Season `json:"seasons"`
}

type Season struct {
	Teams map[string]TeamKits `json:"teams"`
}

type League struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Logo      string `json:"logo,omitempty"`
	IsPopular bool   `json:"is_popular,omitempty"`
}

type TeamKits struct {
	Name      string            `json:"name"`
	Logo      string            `json:"logo,omitempty"`
	League    string            `json:"league"`
	IsPopular bool              `json:"is_popular,omitempty"`
	Kits      map[string]string `json:"kits"`
}

type Team struct {
	Name      string                       `json:"name"`
	Logo      string                       `json:"logo,omitempty"`
	League    string                       `json:"league"`
	IsPopular bool                         `json:"is_popular,omitempty"`
	Seasons   map[string]map[string]string `json:"seasons"`
}

type KitRecord struct {
	TeamID      string
	TeamName    string
	TeamLogo    string
	TeamPopular bool
	League      League
	Season      string
	KitType     string
	URL         string
	Source      string
	ArticleURL  string
}

type LogoRecord struct {
	TeamID     string
	TeamName   string
	TeamLogo   string
	League     League
	URL        string
	Source     string
	ArticleURL string
}
