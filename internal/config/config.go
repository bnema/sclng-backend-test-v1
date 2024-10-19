package config

import (
	"log"
	"os"

	_ "github.com/joho/godotenv/autoload"
)

// Config struct to hold configuration values
type Config struct {
	GitHubToken string
	Port        string
}

// LoadConfig reads environment variables and returns a Config struct
func LoadConfig() *Config {
	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		log.Fatal("GITHUB_TOKEN is not set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT is not set")
	}

	return &Config{
		GitHubToken: githubToken,
		Port:        port,
	}
}
