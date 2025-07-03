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

// ConvertStringToTimeRFC3339 converts a RFC3339 string to time.Time
// Example: "2025-07-03T11:48:48.060949386Z" → time.Time
func ConvertStringToTimeRFC3339(dateStr string) (time.Time, error) {
	// Try RFC3339Nano first (with nanoseconds)
	t, err := time.Parse(time.RFC3339Nano, dateStr)
	if err != nil {
		// Fallback to regular RFC3339 (without nanoseconds)
		t, err = time.Parse(time.RFC3339, dateStr)
		if err != nil {
			return time.Time{}, err
		}
	}
	return t, nil
}

// ConvertTimeToRFC3339 converts time.Time to RFC3339 string with nanoseconds
// Example: time.Time → "2025-07-03T11:48:48.060949386Z"
func ConvertTimeToRFC3339(date time.Time) string {
	return date.Format(time.RFC3339Nano)
}

// ConvertTimeToRFC3339Simple converts time.Time to simple RFC3339 string (without nanoseconds)
// Example: time.Time → "2025-07-03T11:48:48Z"
func ConvertTimeToRFC3339Simple(date time.Time) string {
	return date.Format(time.RFC3339)
}

// ConvertISO8601ToRFC3339 converts ISO 8601 date to RFC3339 timestamp
// Example: "2025-07-03" → "2025-07-03T00:00:00Z"
func ConvertISO8601ToRFC3339(dateStr string) (string, error) {
	t, err := ConvertISO8601ToTime(dateStr)
	if err != nil {
		return "", err
	}
	// Convert to UTC and format as RFC3339
	return t.UTC().Format(time.RFC3339), nil
}

// ConvertRFC3339ToISO8601 converts RFC3339 timestamp to ISO 8601 date
// Example: "2025-07-03T11:48:48.060949386Z" → "2025-07-03"
func ConvertRFC3339ToISO8601(rfc3339Str string) (string, error) {
	t, err := ConvertStringToTimeRFC3339(rfc3339Str)
	if err != nil {
		return "", err
	}
	return ConvertTimeToISO8601(t), nil
}

// ConvertRFC3339ToYYYYMMDD converts RFC3339 timestamp to YYYY-MM-DD string format
// Example: "2025-07-03T11:48:48.060949386Z" → "2025-07-03"
func ConvertRFC3339ToYYYYMMDD(rfc3339Str string) (string, error) {
	t, err := ConvertStringToTimeRFC3339(rfc3339Str)
	if err != nil {
		return "", err
	}
	return t.Format("2006-01-02"), nil
}

// ConvertRFC3339ToDateFormat converts RFC3339 timestamp directly to YYYY-MM-DD format
// Example: "2025-07-03T11:48:48.060949386Z" → "2025-07-03"
func ConvertRFC3339ToDateFormat(rfc3339Str time.Time) (string, error) {
	// Force YYYY-MM-DD format
	return rfc3339Str.Format("2006-01-02"), nil
}
