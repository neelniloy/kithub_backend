package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func main() {
	resp, err := http.Get("https://www.thesportsdb.com/api/v1/json/3/all_leagues.php")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var result struct {
		Leagues []struct {
			ID    string `json:"idLeague"`
			Name  string `json:"strLeague"`
			Sport string `json:"strSport"`
		} `json:"leagues"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	soccerCount := 0
	for _, l := range result.Leagues {
		if l.Sport == "Soccer" {
			soccerCount++
			fmt.Printf("%s: %s\n", l.ID, l.Name)
		}
	}
	fmt.Printf("\nTotal Soccer Leagues: %d\n", soccerCount)
}
