package handlers

import (
	"net/http"

	"myapp/pkg/config"
	"myapp/pkg/models"
	"myapp/pkg/render"
)

// Handler contains the dependencies used by the HTTP handlers.
type Handler struct {
	Renderer *render.Renderer
	App      *config.AppConfig
}

// NewHandler creates a new HTTP handler.
func NewHandler(app *config.AppConfig, r *render.Renderer) *Handler {
	return &Handler{
		Renderer: r,
		App:      app,
	}
}

// Home is the home page handler
func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	h.Renderer.RenderTemplate(w, "home.page.tmpl", &models.TemplateData{})
}

// About is the about page handler
func (h *Handler) About(w http.ResponseWriter, r *http.Request) {
	// perform some logic
	stringMap := make(map[string]string)
	stringMap["test"] = "Hello, again."

	// send the data to the template
	h.Renderer.RenderTemplate(w, "about.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
	})
}
