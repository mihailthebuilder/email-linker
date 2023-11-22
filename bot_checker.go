package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

var botIps = []string{
	// linkedin message preview
	"172.176.75.89",
	"52.165.149.97",
}

func (lpc *LinkPreviewChecker) IsBotRequest(req *http.Request) (bool, error) {
	if lpc.isBotBasedOnIp(req) {
		log.Printf("redirect request url path %s coming from bot ip", req.URL.Path)
		return true, nil
	}

	ua, err := lpc.isBotBasedOnUserAgent(req)
	if err != nil {
		return false, fmt.Errorf("error checking if request for url path %s is bot based on user agent: %s", req.URL.Path, err)
	}

	if ua {
		log.Printf("redirect request url path %s coming from bot user-agent", req.URL.Path)
	}

	return ua, nil
}

func (lpc *LinkPreviewChecker) isBotBasedOnIp(req *http.Request) bool {
	requestIps := req.Header.Values("X-Forwarded-For")

	for _, requestIp := range requestIps {
		for _, botIp := range botIps {
			if requestIp == botIp {
				return true
			}
		}
	}

	return false
}

func (lpc *LinkPreviewChecker) isBotBasedOnUserAgent(req *http.Request) (bool, error) {
	ua := req.Header.Get("User-Agent")
	for _, bot := range lpc.Bots {
		matched, err := regexp.Match(bot.Pattern, []byte(ua))
		if err != nil {
			return false, fmt.Errorf("error regexp.Match on pattern %s against userAgent %s: %s", bot.Pattern, ua, err)
		}

		if matched {
			log.Printf("redirect request url %s identified as bot", req.URL)
			return true, nil
		}
	}

	return false, nil
}
