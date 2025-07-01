package utils

import (
	"encoding/json"
	"time"
)

func StructToMap(data interface{}) (map[string]interface{}, error) {
	var result map[string]interface{}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(jsonData, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func StringToTime(format string, date string) (time.Time, error) {
	t, err := time.Parse(format, date)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}
