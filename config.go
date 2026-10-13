package main

import (
	"encoding/json"
	"os"
)

type DirectionConfig struct {
	SLDirId      int    `json:"dir"`
	FallbackName string `json:"fallback"`
}

type LineConfig struct {
	Name       string            `json:"name"`
	Type       string            `json:"type"`
	Group      *string           `json:"group"`
	Directions []DirectionConfig `json:"dirs"`
}

type SiteConfig struct {
	Id           int          `json:"id"`
	Name         string       `json:"name"`
	Abbreviation string       `json:"abbr"`
	Lines        []LineConfig `json:"lines"`
}

type Config struct {
	Sites []SiteConfig `json:"sites"`
}

var config Config

func LoadConfig() {
	res, err := os.ReadFile("config.json")
	if err != nil {
		println("UNABLE TO READ CONTENTS OF CONFIG FILE: " + err.Error())
		os.Exit(-1)
	}
	err = json.Unmarshal(res, &config)
	if err != nil {
		println("UNABLE TO READ CONFIG FILE: " + err.Error())
		os.Exit(-1)
	}
	println("Config loaded:")
	res, err = json.Marshal(config)
	if err != nil {
		println("Failed to print config: " + err.Error())
		os.Exit(-1)
	}
	println(string(res))
}

func GetConfig() Config {
	return config
}
