package main

import (
	"encoding/json"
	"os"
)

type Config struct {
	SlackAPIToken       string   `json:"slackApiToken"`
	TimelineChannelID   string   `json:"timelineChannelID"`
	BlackListChannelIDs []string `json:"blackListChannelIDs"`
	Sentry              sentry   `json:"sentry"`
}

type sentry struct {
	DSN *string `json:"dsn"`
}

func ReadConfig(path string) (*Config, error) {
	result := Config{}
	file, openErr := os.Open(path)
	if openErr != nil {
		return nil, openErr
	}
	decoder := json.NewDecoder(file)
	err := decoder.Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
