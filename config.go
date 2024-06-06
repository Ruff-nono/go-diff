package main

import (
	"encoding/json"
	"log"
	"os"
)

var config Config

// Config stores the configuration options
type Config struct {
	Host1                 string   `json:"host1"`
	Host2                 string   `json:"host2"`
	HeadersInclude        []string `json:"headers_include"`
	CompareStatusCode     bool     `json:"compare_status_code"`
	EquivalentStatusCodes [][]int  `json:"equivalent_status_codes"`
	CompareBody           bool     `json:"compare_body"`
	BodiesExclude         []string `json:"bodies_exclude"`
}

func init() {
	configFile, err := os.Open("config.json")
	if err != nil {
		log.Fatalf("Failed to open config file: %v\n", err)
	}
	defer configFile.Close()

	err = json.NewDecoder(configFile).Decode(&config)
	if err != nil {
		log.Fatalf("Failed to decode config file: %v\n", err)
	}
}
