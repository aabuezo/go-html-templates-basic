package main

import (
	"log"
	"myapp/internal/config"
	"myapp/internal/handlers"
	"myapp/internal/render"
	"myapp/internal/session"
	"net/http"
)

const portNumber = ":8080"

func main() {

	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("cannot create template cache")
	}

	app := &config.AppConfig{
		UseCache:      false,
		TemplateCache: tc,
		InProduction:  false, // change to true for Production
	}

	app.Session = session.New(app.InProduction)

	renderer := render.NewRenderer(app)
	handler := handlers.NewHandler(app, renderer)

	log.Println("Starting application on port", portNumber)

	srv := &http.Server{
		Addr:    portNumber,
		Handler: routes(handler, app),
	}

	err = srv.ListenAndServe()
	log.Fatal(err)
}
