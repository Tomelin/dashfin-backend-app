package utils

import "time"

func ConvertStringToTime(format string, dateStr string) (time.Time, error) {
	// Parse the date string into a time.Time object
	t, err := time.Parse(format, dateStr)
	if err != nil {
		return time.Time{}, err
	}

	// Format the time.Time object back to a string in the desired format
	return t, nil
}

func ConvertTimeToString(format string, date time.Time) string {
	// Format the time.Time object to a string in the desired format
	return date.Format(format)
}

func ConvertTimeToISO8601(date time.Time) string {
	// Format the time.Time object to a string in ISO 8601 format (YYYY-MM-DD)
	return date.Format("2006-01-02")
}

func ConvertISO8601ToTime(dateStr string) (time.Time, error) {
	// Parse the date string into a time.Time object in ISO 8601 format
	return time.Parse("2006-01-02", dateStr)
}

func ConvertTimeToYYYYMM(date time.Time) string {
	// Format the time.Time object to a string in ISO 8601 format (YYYY-MM)
	return date.Format("2006-01")
}

func GetFirstDayOfCurrentMonth() time.Time {
	// Get the current time
	currentTime := time.Now()

	// Create a new time.Time object for the first day of the current month
	firstDayOfMonth := time.Date(currentTime.Year(), currentTime.Month(), 1, 0, 0, 0, 0, currentTime.Location())

	return firstDayOfMonth
}

// Get último dia do mês atual
func GetLastDayOfCurrentMonth() time.Time {
	// Get the current time
	currentTime := time.Now()

	// Create a new time.Time object for the last day of the current month
	lastDayOfMonth := time.Date(currentTime.Year(), currentTime.Month()+1, 0, 0, 0, 0, 0, currentTime.Location())

	return lastDayOfMonth
}
