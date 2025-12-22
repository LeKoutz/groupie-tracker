package services

import (
	"strconv"
)

func ExtractYearFromDate(dateStr string) int {
	if dateStr == "" {
		return 0
	}
	// extract year from "YYYY-MM-DD" format
	if len(dateStr) < 4 {
		yearStr := dateStr[:4] // YYYY
		year, err := strconv.Atoi(yearStr)
		if err == nil {
			return year
		}
	}
	return 0
}