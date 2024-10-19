package main

import (
	"fmt"
	"log"

	"scalingo-api-test/internal/api"
	"scalingo-api-test/internal/cache"
	"scalingo-api-test/internal/config"
	"scalingo-api-test/internal/services"
)

func main() {
	// Load the config
	config := config.LoadConfig()

	// Create a new cache
	cache := cache.NewCache()

	// Pass the config to the service to create a new repo service
	repoService, err := services.NewRepoService(config)
	if err != nil {
		fmt.Printf("Error creating repo service: %v\n", err)
		return
	}

	// Fetch 100 today's public repositories with non-null language
	// This will also fetch the language details for each repository
	// I decided to do it here and block until its done so the cache is always ready when the server start
	repos, err := repoService.GetTodaysPublicRepos()
	if err != nil {
		fmt.Printf("Error fetching repositories: %v\n", err)
		return
	}

	// Transform the repositories to a custom format so we can cache it
	customRepos, err := repoService.TransformToCustomRepositories(repos)
	if err != nil {
		fmt.Printf("Error transforming repositories: %v\n", err)
		return
	}

	// Set the cache with the transformed repositories
	cache.Set(customRepos)

	// Create and start the server
	server := api.NewServer(repoService, cache)

	log.Printf("Starting server on port %s\n", config.Port)
	log.Fatal(server.Start(fmt.Sprintf(":%s", config.Port)))

}
