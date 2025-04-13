package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	MaxRecordsPerSecond int        `yaml:"maxRecordsPerSecond"`
	TCP                 Connection `yaml:"TCP"`
}

var Cfg Config

type Connection struct {
	Port     string `yaml:"port"`
	Host     string `yaml:"host"`
	Protocol string `yaml:"protocol"`
}

func LoadConfig(filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Error opening config file: %v", err)
	}
	defer file.Close()

	Cfg := Config{}

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&Cfg); err != nil {
		log.Fatalf("Error decoding config file: %v", err)
	}
}
