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

		year, err := strconv.Atoi(match[1])
		if err != nil {
			continue
		}

		// Use the end year as the season identifier (e.g. 2024-25 -> 2025)
		shortYear, ok := parseSeasonEndYear(match[2])
		if !ok {
			continue
		}

		// Normalize to 4 digit year
		if shortYear < 100 {
			fullYear := 2000 + shortYear
			if fullYear < year { // Handle century crossover if necessary, though unlikely here
				fullYear += 100
			}
			return strconv.Itoa(fullYear)
		}
		return strconv.Itoa(shortYear)
	}

	for _, match := range shortSeasonRangeRE.FindAllStringSubmatch(text, -1) {
		if len(match) != 3 {
			continue
		}

		endYear, err := strconv.Atoi(match[2])
		if err != nil {
			continue
		}
		return strconv.Itoa(2000 + endYear)
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
