package env

import (
	"os"
	"strconv"
	"strings"
)

func GetString(key string, fallback string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	return val
}

func GetInt(key string, fallback int) int {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	vallAsInt, error := strconv.Atoi(val)
	if error != nil {
		return fallback
	}
	return vallAsInt
}

func GetBool(key string, fallback bool) bool {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	if boolValue, err := strconv.ParseBool(val); err == nil {
		return boolValue
	}
	return fallback
}

func GetStringSlice(key string, fallback []string) []string {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	return strings.Split(val, ",")
}
