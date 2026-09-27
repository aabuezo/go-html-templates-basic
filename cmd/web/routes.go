package main

import (
	"myapp/internal/config"
	"myapp/internal/handlers"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func routes(handler *handlers.Handler, app *config.AppConfig) http.Handler {
	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	mux.Use(NoSurf(app))
	mux.Use(SessionLoad(app))

	mux.Get("/", handler.Home)
	mux.Get("/about", handler.About)

	mux.Get("/login", handler.Login)
	mux.Post("/login", handler.LoginForm)

	mux.Get("/user-home", handler.UserHome)

	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	return mux
}
