package utils

import (
	"regexp"
	"strconv"
)

var (
	seasonRangeRE      = regexp.MustCompile(`\b(20\d{2})\s*[-/]\s*(\d{2}|20\d{2})\b`)
	shortSeasonRangeRE = regexp.MustCompile(`\b(2\d)\s*[-/]\s*(\d{2})\b`)
	seasonYearRE       = regexp.MustCompile(`\b(20\d{2})\b`)
)

func ExtractSeason(text string) string {
	for _, match := range seasonRangeRE.FindAllStringSubmatch(text, -1) {
		if len(match) != 3 {
			continue
		}

		// Use the START year as the season identifier (e.g. 2024-25 -> 2024)
		// This matches the user's expectation where 2025 means the 2025-26 season.
		year := match[1]
		if len(year) == 2 {
			year = "20" + year
		}
		return year
	}

	for _, match := range shortSeasonRangeRE.FindAllStringSubmatch(text, -1) {
		if len(match) != 3 {
			continue
		}

		startYear := match[1]
		return "20" + startYear
	}

	if match := seasonYearRE.FindStringSubmatch(text); len(match) == 2 {
		return match[1]
	}

	return ""
}

func IsValidSeason(season string) bool {
	// Only allow 2024, 2025, 2026
	return season == "2024" || season == "2025" || season == "2026"
}

func parseSeasonEndYear(value string) (int, bool) {
	year, err := strconv.Atoi(value)
	if err != nil {
		return 0, false
	}
	if len(value) == 4 {
		return year % 100, true
	}
	return year, true
}
