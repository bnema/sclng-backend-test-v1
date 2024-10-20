package services

import (
	"context"
	"fmt"
	"log"
	"scalingo-api-test/internal/config"
	"scalingo-api-test/internal/models"
	"sort"
	"sync"
	"time"

	"github.com/google/go-github/v66/github"
)

// RepoService provides methods to interact with GitHub repositories
type RepoService struct {
	client *github.Client
	config *config.Config
}

// RepoInfo contains repository information including languages
type RepoInfo struct {
	Repo      *github.Repository
	Languages map[string]int
}

// NewRepoService creates a new RepoService with an authenticated GitHub client
func NewRepoService(config *config.Config) (*RepoService, error) {
	// Create a GitHub client
	client := github.NewClient(nil)
	// If a token is provided, use it to authenticate the client
	client = client.WithAuthToken(config.GitHubToken)
	log.Println("Repo service started using GitHub's authenticated client")

	return &RepoService{
		client: client,
		config: config,
	}, nil
}

// GetTodaysPublicRepos fetches today's public repositories and returns them sorted by creation time
func (s *RepoService) GetTodaysPublicRepos() ([]*RepoInfo, error) {
	ctx := context.Background()

	// Fetch 100 repositories with non-null language
	repos, err := s.fetchReposWithLanguage(ctx)
	if err != nil {
		return nil, err
	}

	// Create RepoInfo slice
	repoInfos := make([]*RepoInfo, len(repos))
	for i, repo := range repos {
		repoInfos[i] = &RepoInfo{
			Repo:      repo,
			Languages: make(map[string]int),
		}
	}

	// Fetch language details for the repositories
	s.FetchLanguageDetails(repoInfos)

	// Sort repositories by creation time (newest first)
	sort.Slice(repoInfos, func(i, j int) bool {
		return repoInfos[i].Repo.GetCreatedAt().After(repoInfos[j].Repo.GetCreatedAt().Time)
	})

	return repoInfos, nil
}

// fetchReposWithLanguage fetches repositories with non-null language
func (s *RepoService) fetchReposWithLanguage(ctx context.Context) ([]*github.Repository, error) {

	// Get the current time in UTC
	now := time.Now().UTC()

	// For the time range, we will use the current time and substract 10 minutes
	endTime := now
	// Substract 5 minutes to the end time
	startTime := endTime.Add(-5 * time.Minute)

	// Set the time as RFC3339 format and add the time range
	query := fmt.Sprintf("is:public created:%s..%s", startTime.Format(time.RFC3339), endTime.Format(time.RFC3339))

	// Set the options for the GitHub API request
	opts := &github.SearchOptions{
		Sort:  "created",
		Order: "desc",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	var allRepos []*github.Repository

	log.Println("Start fetching repositories with non-null language...")
	for len(allRepos) < 100 {
		result, _, err := s.client.Search.Repositories(ctx, query, opts)
		if err != nil {
			return nil, fmt.Errorf("error searching repositories: %w", err)
		}

		for _, repo := range result.Repositories {
			if repo.GetLanguage() != "" {
				allRepos = append(allRepos, repo)
				if len(allRepos) == 100 {
					break
				}
			}
		}

		if len(result.Repositories) < opts.PerPage {
			break
		}

		opts.Page++
	}

	log.Printf("Fetched %d repositories", len(allRepos))

	return allRepos, nil
}

// FetchLanguageDetails fetches language details for the given RepoInfo slice
func (s *RepoService) FetchLanguageDetails(repoInfos []*RepoInfo) {
	log.Println("Start fetching language details for our 100 repos...")
	ctx := context.Background()
	var wg sync.WaitGroup
	maxRequests := make(chan struct{}, 100) // Just in case we need to slow it down

	for _, repoInfo := range repoInfos {
		wg.Add(1)
		go func(ri *RepoInfo) {
			defer wg.Done()
			maxRequests <- struct{}{}
			defer func() { <-maxRequests }()
			ri.Languages = s.fetchLanguageDetails(ctx, ri.Repo)
		}(repoInfo)
	}

	wg.Wait()
}

// fetchLanguageDetails fetches language details for a given repository
func (s *RepoService) fetchLanguageDetails(ctx context.Context, repo *github.Repository) map[string]int {
	languages, _, err := s.client.Repositories.ListLanguages(ctx, repo.GetOwner().GetLogin(), repo.GetName())
	if err != nil {
		fmt.Printf("Error fetching languages for %s: %v\n", repo.GetFullName(), err)
		return nil
	}

	return languages
}

// TransformToCustomRepositories transforms a slice of RepoInfo to a slice of CustomRepository
func (s *RepoService) TransformToCustomRepositories(repoInfos []*RepoInfo) ([]*models.CustomRepository, error) {
	customRepos := make([]*models.CustomRepository, len(repoInfos))
	var wg sync.WaitGroup
	errChan := make(chan error, len(repoInfos))

	for i, repoInfo := range repoInfos {
		wg.Add(1)
		go func(i int, repoInfo *RepoInfo) {
			defer wg.Done()
			customRepo, err := s.transformSingleRepo(repoInfo)
			if err != nil {
				errChan <- err
				return
			}
			customRepos[i] = customRepo
		}(i, repoInfo)
	}

	wg.Wait()
	close(errChan)

	if len(errChan) > 0 {
		return nil, <-errChan // Return the first error encountered
	}

	return customRepos, nil
}

// transformSingleRepo transforms a single RepoInfo to a CustomRepository
func (s *RepoService) transformSingleRepo(repoInfo *RepoInfo) (*models.CustomRepository, error) {
	repo := repoInfo.Repo

	// Transform languages to CustomRepository languages
	languages := make(map[string]models.Language)
	for lang, bytes := range repoInfo.Languages {
		languages[lang] = models.Language{Bytes: bytes}
	}

	// Update and return the CustomRepository
	return &models.CustomRepository{
		FullName:   repo.GetFullName(),
		Owner:      repo.GetOwner().GetLogin(),
		Repository: repo.GetName(),
		Languages:  languages,
		License:    repo.GetLicense().GetName(),
		CreatedAt:  repo.GetCreatedAt().Time,
		UpdatedAt:  repo.GetUpdatedAt().Time,
		PushedAt:   repo.GetPushedAt().Time,
		Stars:      repo.GetStargazersCount(),
		Forks:      repo.GetForksCount(),
		Issues:     repo.GetOpenIssuesCount(),
	}, nil
}
