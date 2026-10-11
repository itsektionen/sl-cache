package main

import (
	"encoding/json"
	"os"
)

type SiteConfig struct {
	Name         string `json:"name"`
	Abbreviation string `json:"abbr"`
}

type Config struct {
	Sites map[string]SiteConfig `json:"sites"`
}

var config Config

func LoadConfig() {
	res, err := os.ReadFile("./config.json")
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
