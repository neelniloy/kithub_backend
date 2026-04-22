package utils

import (
	"fmt"
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

		expectedShortYear := (year + 1) % 100
		shortYear, ok := parseSeasonEndYear(match[2])
		if !ok || shortYear != expectedShortYear {
			continue
		}

		return fmt.Sprintf("%d-%02d", year, shortYear)
	}

	for _, match := range shortSeasonRangeRE.FindAllStringSubmatch(text, -1) {
		if len(match) != 3 {
			continue
		}

		startYear, err := strconv.Atoi(match[1])
		if err != nil {
			continue
		}
		endYear, err := strconv.Atoi(match[2])
		if err != nil {
			continue
		}
		if endYear != startYear+1 {
			continue
		}

		return fmt.Sprintf("20%02d-%02d", startYear, endYear)
	}

	if match := seasonYearRE.FindStringSubmatch(text); len(match) == 2 {
		year, err := strconv.Atoi(match[1])
		if err != nil {
			return ""
		}
		return fmt.Sprintf("%d-%02d", year, (year+1)%100)
	}

	return ""
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
