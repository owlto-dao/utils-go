package convert

import (
	"encoding/json"
	"strconv"
	"time"
)

func ConvertIntToString[T ~int | ~int8 | ~int16 | ~int32 | ~int64](value T) string {
	return strconv.FormatInt(int64(value), 10)
}

func ConvertStringToInt[T ~int | ~int8 | ~int16 | ~int32 | ~int64](value string) T {
	result, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0
	}
	return T(result)
}

func ConvertStringToUint64(value string) uint64 {
	result, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0
	}
	return result
}

func ConvertStringToPtrTime(value string) *time.Time {
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil
	}
	return &t
}

func ConvertStringToTime(value string) time.Time {
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}
	}
	return t
}

func ConvertStringToFloat64(value string) float64 {
	result, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	return result
}

func ConvertToJsonString(obj interface{}) string {
	jsonData, _ := json.Marshal(obj)
	return string(jsonData)
}
