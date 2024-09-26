package convert

import (
	"encoding/json"
	"strconv"
	"time"
)

func IntToBool[T ~int | ~int8 | ~int16 | ~int32 | ~int64](value T) bool {
	return value != 0
}

func BoolToInt[T ~int | ~int8 | ~int16 | ~int32 | ~int64](value bool) T {
	if value {
		return T(1)
	}
	return T(0)
}

func IntToString[T ~int | ~int8 | ~int16 | ~int32 | ~int64](value T) string {
	return strconv.FormatInt(int64(value), 10)
}

func StringToInt[T ~int | ~int8 | ~int16 | ~int32 | ~int64](value string) T {
	result, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0
	}
	return T(result)
}

func StringToUint64(value string) uint64 {
	result, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0
	}
	return result
}

func StringToPtrTime(value string) *time.Time {
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil
	}
	return &t
}

func StringToTime(value string) time.Time {
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}
	}
	return t
}

func StringToFloat64(value string) float64 {
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
