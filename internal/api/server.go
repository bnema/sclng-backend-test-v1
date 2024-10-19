package api

import (
	"scalingo-api-test/internal/cache"
	"scalingo-api-test/internal/services"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Server struct {
	echo        *echo.Echo
	repoService *services.RepoService
	cache       *cache.Cache
}

func NewServer(repoService *services.RepoService, cache *cache.Cache) *Server {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	s := &Server{
		echo:        e,
		repoService: repoService,
		cache:       cache,
	}

	// Use middleware
	s.echo.Use(middleware.Logger())
	s.echo.Use(middleware.Recover())

	// Setup routes
	s.setupRoutes()

	return s
}

func (s *Server) Start(address string) error {
	return s.echo.Start(address)
}
