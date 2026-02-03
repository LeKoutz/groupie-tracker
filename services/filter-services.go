package services

func ExtractYearFromDate(dateStr string) int {
	t, err := parseDate(dateStr)
	if err != nil {
		return 0
	}
	return t.Year()
}
