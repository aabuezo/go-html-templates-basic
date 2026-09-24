package main

import (
	"log"
	"myapp/pkg/config"
	"myapp/pkg/handlers"
	"myapp/pkg/render"
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
	}

	renderer := render.NewRenderer(app)
	handler := handlers.NewHandler(renderer)

	http.HandleFunc("/{$}", handler.Home)
	http.HandleFunc("/about", handler.About)

	log.Println("Starting application on port", portNumber)
	_ = http.ListenAndServe(portNumber, nil)
}
