package api

import (
	"net/http"
	"scalingo-api-test/internal/api/handlers"

	"github.com/labstack/echo/v4"
)

func (server *Server) setupRoutes() {
	// Create a new search handler
	searchHandler := handlers.NewSearchHandler(server.repoService, server.cache)

	// Ping endpoint
	server.echo.GET("/ping", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "pong"})
	})

	// API group
	api := server.echo.Group("/api")

	// Search endpoint
	api.GET("/search", searchHandler.Handle)

	// We can add more endpoints to the api group if needed
}
