package utils

import "encoding/json"

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
