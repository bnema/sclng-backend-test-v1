# GitHub Repository Explorer

## Table of Contents
- [Reasoning](#reasoning)
- [Explanation of the project structure](#explanation-of-the-project-structure)
- [Usage](#usage)
  - [Ping](#ping)
  - [Search Repositories](#search-repositories)
    - [Query Parameters](#query-parameters)
    - [Response](#response)
- [Execution](#execution)
- [Possible optimizations and features](#possible-optimizations-and-features)
- [Final note](#final-note)

## Reasoning

List of steps:

1. Retrieve the 100 latest public repositories
2. Get the list of languages used for each of these repositories (as fast as possible)
3. Transform the fetched GitHub repository data and language information into our CustomRepository model (as fast as possible)
4. Set up a cache to serve data to clients
5. Serve the data on our /search endpoint
6. Offer query params to filter the data (language, license, etc.)

It's important to consider performance and scalability from the start. If we have 1 million requests/s on our endpoint, we can't afford to make 1 million requests/s to the GitHub API. It's therefore crucial to set up a caching system to serve the same data in our possession to all clients.

The exercise involves performing an initial fetch to retrieve the 100 latest public GitHub repositories and then, for each of these repositories, retrieve the list of languages used. That's a total of 101 requests to be made to the GitHub API each time we want to update the cache.

For information, the GitHub API has the following rate limits:

- Unauthenticated: 60 requests per hour
- Authenticated: 5000 requests per hour	

It's therefore imperative to use the authenticated (via a personal access token) API to avoid exceeding the rate limit and getting blocked.

## Explanation of the project structure

### `cmd` folder

- `cmd/api/main.go`: Application entry point. 

### `internal` folder

Contains application-specific code, not intended to be imported by other projects.

- `api/`: Logic related to the HTTP API.
  - `handlers/search_handler.go`: Handling of HTTP requests for repositories.
  - `routes.go`: Definition of API routes.
  - `server.go`: HTTP server configuration and launch.
- `cache/cache.go`: Implementation of caching logic.
- `config/config.go`: Application configuration management.
- `models/custom_repository.go`: Data structures for repositories.
- `services/repository_service.go`: Logic to fetch and transform repositories.


## Usage

The API provides the following endpoint:

### Ping

- **Endpoint**: `/ping`
- **Method**: GET
- **Description**: Check if the API is running.

### Search Repositories

- **Endpoint**: `/api/search`
- **Method**: GET
- **Description**: Retrieves and filters the latest public repositories from GitHub.


#### Query Parameters

1. `lang` (optional): Filter repositories by programming language.
   - Example: `/api/search?lang=go`

2. `license` (optional): Filter repositories by license type.
   - Example: `/api/search?license=MIT`

3. `stars` (optional): Filter repositories by minimum number of stars.
   - Example: `/api/search?stars=1000`

You can combine multiple parameters:
```
/api/search?lang=python&license=MIT&stars=500
```

#### Response

The endpoint returns a JSON object with the following structure:

```json
{
  "last_updated": "2024-01-00T00:00:00Z",
  "total_results": 42,
  "repositories": [
    {
      "full_name": "owner/repo",
      "owner": "owner",
      "repository": "repo",
      "languages": {
        "go": {
          "bytes": 123456
        }
      },
      "license": "MIT License",
      "created_at": "2024-01-00T00:00:00Z",
      "updated_at": "2024-01-00T00:00:00Z",
      "pushed_at": "2024-01-00T00:00:00Z",
      "stars": 1000,
      "forks": 50,
      "issues": 10
    },
    // ... more repositories
  ]
}
```

## Execution

Create a `.env` file and set those variables:
```
GITHUB_TOKEN=your_github_token
PORT=5000
``` 

Run the application:

```
docker compose up
or
podman-compose up
```

Application will be then running on port `5000`


## Possible optimizations and features
   - Fetch and update the cache at regular intervals (cronjob in background)
   - Implement a rate limiting middleware on our endpoints to protect against abuse.
   - Use a distributed cache (like Redis) if the application needs to be deployed on multiple instances (horizontal scaling).
   - Aggregate and store data to have an history
   - Offer other endpoints (example: stats like "most used language in the last month")
   - A proper logging system/package and metrics


## Final note

I had fun with this. I struggled to retrieve the most recent repos from GitHub by "created" and order by "desc" (on a large query), my compromise was to retrieve all those between a range of 5 minutes ago and now (in UTC). Also, it doesnt seems possible to retrieve languages details on a large query, so I had to make a request for each repository.

From my testing it seems there is an average of 100 repositories created each minute but they are for the most part empty or with no language. So I set the range to 5 minutes (~500 repositories) and I ignore the ones with no language. It consistently returns me 100 results so I think it's a good compromise.

With performance in mind, I allowed myself to create a goroutine that executes 100 requests in parallel to retrieve the languages details. This doesn't seem to bother GitHub if I'm authenticated, and the application's cold start is significantly faster.