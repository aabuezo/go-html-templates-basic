package main

import (
	"myapp/internal/config"
	"net/http"

	"github.com/justinas/nosurf"
)

// NoSurf adds CSRF protection to all POST requests
func NoSurf(app *config.AppConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		csrfHandler := nosurf.New(next)

		csrfHandler.SetBaseCookie(http.Cookie{
			HttpOnly: true,
			Path:     "/",              // applies to the entire site
			Secure:   app.InProduction, // true for HTTPS, false for HTTP
			SameSite: http.SameSiteLaxMode,
		})

		return csrfHandler
	}
}

// SessionLoad loads and saves the session on every request
func SessionLoad(app *config.AppConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return app.Session.LoadAndSave(next)
	}
}
