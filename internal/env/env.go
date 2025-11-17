package env

import (
	"os"
	"strconv"
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
