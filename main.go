package main

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/google/go-github/v66/github"
)

func main() {
	// Create a new GitHub client
	client := github.NewClient(nil)

	// Get current time and set it to the start of the day
	// This will be used to search for repositories created today
	// Because otherwise I end up with wrong created_at results for some reason
	today := time.Now().Format("2006-01-02")

	// Print the search time for debugging purposes
	fmt.Println("Search time:", today)

	// Create a search query
	// This query will find public repositories created today
	query := fmt.Sprintf("is:public created:%s", today)

	// Set up search options
	opts := &github.SearchOptions{
		Sort:  "created", // Sort results by creation time
		Order: "desc",    // Sort in descending order (newest first)
		ListOptions: github.ListOptions{
			PerPage: 100, // Request 100 results per page (maximum allowed by GitHub API)
		},
	}

	// Create a context for the API requests
	ctx := context.Background()

	// Initialize a slice to store all repositories
	var allRepos []*github.Repository

	// Initialize the page number for pagination
	page := 1

	// Loop to fetch repositories until we have at least 100 or there are no more results
	for len(allRepos) < 100 {
		// Set the current page number
		opts.Page = page

		// Perform the search request to the GitHub API
		result, _, err := client.Search.Repositories(ctx, query, opts)
		if err != nil {
			// If there's an error, log it and exit the program
			log.Fatalf("Error searching repositories: %v", err)
		}

		// Filter out repositories with null languages because they're not useful for our test
		for _, repo := range result.Repositories {
			if repo.GetLanguage() != "" {
				// Add repositories with non-null languages to our list
				allRepos = append(allRepos, repo)
				if len(allRepos) >= 100 {
					// Break if we've reached 100 repositories
					break
				}
			}
		}

		// If we received fewer results than requested, we've reached the end
		if len(result.Repositories) < opts.PerPage {
			break
		}

		// If we reach this point, we need to move to the next page because either:
		// 1. We haven't collected 100 repos with non-null languages yet, or
		// 2. We received a full page of results, but some had null languages
		page++
	}

	// Sort the repositories by creation time (newest first)
	sort.Slice(allRepos, func(i, j int) bool {
		return allRepos[i].GetCreatedAt().After(allRepos[j].GetCreatedAt().Time)
	})

	// Trim the results to 100 if we got more
	if len(allRepos) > 100 {
		allRepos = allRepos[:100]
	}

	// Print the results for debugging
	fmt.Printf("Found %d repositories created today with non-null language:\n", len(allRepos))
	for _, repo := range allRepos {
		fmt.Printf("- %s (Language: %s, Created: %s)\n",
			repo.GetFullName(), // Print the full name of the repository (owner/repo)
			repo.GetLanguage(), // Print the primary language of the repository
			repo.GetCreatedAt().Format("2006-01-02 15:04:05")) // Print the creation time in a readable format
	}
}
