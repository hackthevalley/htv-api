package main

import (
	"log"
	"os"
	"strings"
)

func getEnv(key string, fallback string) string {
	key = strings.TrimSpace(key)
	fallback = strings.TrimSpace(fallback)
	value := strings.TrimSpace(os.Getenv(key))
	if len(value) == 0 {
		log.Printf("Could to get provided env var: %s, using default value instead: %s", key, fallback)
		return fallback
	}
	return value
}
