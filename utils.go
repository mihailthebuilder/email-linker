package main

import (
	"log"
	"os"
	"strconv"
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

func getInt(val string) int {
	valInt, err := strconv.Atoi(val)
	if err != nil {
		log.Fatalf("couldn't convert val = %s from string to int: %s", val, err)
	}
	return valInt
}
