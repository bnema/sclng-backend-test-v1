package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"

	"scalingo-api-test/internal/cache"
	"scalingo-api-test/internal/models"
	"scalingo-api-test/internal/services"
)

// SearchHandler handles search requests and manages access to the repository cache
type SearchHandler struct {
	repoService *services.RepoService
	cache       *cache.Cache
	mu          sync.Mutex // Mutex to ensure thread-safe access to the cache
}

// NewSearchHandler creates a new instance of SearchHandler
func NewSearchHandler(repoService *services.RepoService, cache *cache.Cache) *SearchHandler {
	return &SearchHandler{
		repoService: repoService,
		cache:       cache,
	}
}

// SearchResponse represents the structure of the JSON response for search requests
type SearchResponse struct {
	LastUpdated  time.Time                  `json:"last_updated"`
	TotalResults int                        `json:"total_results"`
	Repositories []*models.CustomRepository `json:"repositories"`
}

// Handle processes the search request and returns filtered repositories
func (sh *SearchHandler) Handle(c echo.Context) error {
	// Access the cache
	sh.mu.Lock()
	defer sh.mu.Unlock()
	repos, lastFetch := sh.cache.Get()

	// Get query parameters
	params := c.QueryParams()

	// Apply filters to the repositories based on query parameters
	filteredRepos := filterRepos(repos, params)

	response := SearchResponse{
		LastUpdated:  lastFetch,
		TotalResults: len(filteredRepos),
		Repositories: filteredRepos,
	}

	return c.JSON(http.StatusOK, response)
}

// filterRepos applies the specified filters to the list of repositories
func filterRepos(repos []*models.CustomRepository, params map[string][]string) []*models.CustomRepository {
	filtered := make([]*models.CustomRepository, 0)

	for _, repo := range repos {
		if matchesFilters(repo, params) {
			filtered = append(filtered, repo)
		}
	}

	return filtered
}

// matchesFilters checks if a repository matches all specified filters
func matchesFilters(repo *models.CustomRepository, params map[string][]string) bool {
	for key, values := range params {
		switch key {
		case "lang":
			if !matchesLanguage(repo, values[0]) {
				return false
			}
		case "license":
			if !matchesLicense(repo, values[0]) {
				return false
			}
		case "stars":
			if !matchesStars(repo, values[0]) {
				return false
			}
		}

		// We can add more filters if needed
	}
	return true
}

// matchesLanguage checks if the repository uses the specified programming language
func matchesLanguage(repo *models.CustomRepository, lang string) bool {

	if lang == "" {
		return true
	}

	if repo.Languages[lang].Bytes > 0 {
		return true
	}

	return false
}

// matchesLicense checks if the repository's license matches the specified license
func matchesLicense(repo *models.CustomRepository, license string) bool {

	// If no license is specified, consider it a match
	if license == "" {
		return true
	}
	// Case-insensitive check for the license in the repository's license field (example: MIT, GPL, etc.)
	if strings.Contains(strings.ToLower(repo.License), strings.ToLower(license)) {
		return true
	}

	return false
}

// matchesStars checks if the repository has at least the specified number of stars
func matchesStars(repo *models.CustomRepository, stars string) bool {
	if stars == "" {
		return true
	}
	count, err := strconv.Atoi(stars)
	if err != nil {
		return false // Invalid star count, consider it a non-match
	}

	if repo.Stars >= count {
		return true
	}

	return false
}
