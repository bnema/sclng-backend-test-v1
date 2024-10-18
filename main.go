package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/go-github/v66/github"
)

func main() {
	// Create a new GitHub client
	client := github.NewClient(nil)

	// Get current time and subtract one minute
	now := time.Now().Add(-60 * time.Minute)
	formattedTime := now.Format(time.RFC3339)

	fmt.Println("Search time:", formattedTime)

	// Create a search query using the provided base query
	query := fmt.Sprintf("is:public created:>=%s", formattedTime)

	// Set up search options
	opts := &github.SearchOptions{
		Sort:  "created",
		Order: "desc",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	// Perform the search
	ctx := context.Background()
	result, _, err := client.Search.Repositories(ctx, query, opts)
	if err != nil {
		log.Fatalf("Error searching repositories: %v", err)
	}

	// Filter repositories with non-null language
	var allRepos []*github.Repository
	for _, repo := range result.Repositories {
		if repo.GetLanguage() != "" {
			allRepos = append(allRepos, repo)
		}
	}

	// Print the results
	fmt.Printf("Found %d repositories created within the last minute with non-null language:\n", len(allRepos))
	for i, repo := range allRepos {
		if i >= 100 {
			break
		}
		fmt.Printf("- %s (Language: %s, Created: %s)\n",
			repo.GetFullName(),
			repo.GetLanguage(),
			repo.GetCreatedAt().Format("2006-01-02 15:04:05"))
	}
}
