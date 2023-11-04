package main

import (
	"log"
	"os"
)

func lengthOfString(input string) int {
	return len([]rune(input))
}

func getEnv(env string) string {
	val := os.Getenv(env)

	if val == "" {
		log.Panicf("environment variable %s not found", env)
	}

	return val
}
