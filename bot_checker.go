package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
)

func newBotChecker() *LinkPreviewChecker {
	path := "crawler-user-agents.json"

	file, err := os.ReadFile(path)
	if err != nil {
		log.Panicf("error reading user-agent JSON file: %s", err)
	}

	var bots []Bot

	err = json.Unmarshal(file, &bots)
	if err != nil {
		log.Panicf("error parsing user-agent JSON file: %s", err)
	}

	return &LinkPreviewChecker{Bots: bots}
}

type LinkPreviewChecker struct {
	Bots []Bot
}

type Bot struct {
	Pattern string `json:"pattern"`
}

func (lpc *LinkPreviewChecker) IsLinkPreviewRequest(userAgent string) (bool, error) {
	for _, bot := range lpc.Bots {
		matched, err := regexp.Match(bot.Pattern, []byte(userAgent))
		if err != nil {
			return false, fmt.Errorf("error regexp.Match on pattern %s against userAgent %s: %s", bot.Pattern, userAgent, err)
		}

		if matched {
			return true, nil
		}
	}

	return false, nil
}
