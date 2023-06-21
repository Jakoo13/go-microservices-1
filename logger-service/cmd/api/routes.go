package main

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// Will hold all our routes for the application
func (app *Config) routes() http.Handler {
	router := chi.NewRouter()

	// specify who is allowed to connect
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// built into chi, will allow to check if service is alive by checking its heartbeat
	router.Use(middleware.Heartbeat("/ping"))

	router.Post("/log", app.WriteLog)

	return router
}
