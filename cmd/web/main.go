package main

import (
	"encoding/gob"
	"log"
	"myapp/internal/config"
	"myapp/internal/handlers"
	"myapp/internal/helpers"
	"myapp/internal/models"
	"myapp/internal/render"
	"myapp/internal/session"
	"net/http"
	"os"
)

const portNumber = ":8080"

func main() {

	// what am I going to put in the session
	gob.Register(models.Login{})

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

	app.InfoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.ErrorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	renderer := render.NewRenderer(app)
	handler := handlers.NewHandler(app, renderer)
	helpers.NewHelpers(app)

	log.Println("Starting application on port", portNumber)

	srv := &http.Server{
		Addr:    portNumber,
		Handler: routes(handler, app),
	}

	err = srv.ListenAndServe()
	log.Fatal(err)
}
