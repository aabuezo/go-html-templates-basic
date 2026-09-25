package handlers

import (
	"net/http"

	"myapp/internal/config"
	"myapp/internal/models"
	"myapp/internal/render"
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
	remoteIP := r.RemoteAddr
	h.App.Session.Put(r.Context(), "remote_ip", remoteIP)

	h.Renderer.RenderTemplate(w, "home.page.tmpl", &models.TemplateData{})
}

// About is the about page handler
func (h *Handler) About(w http.ResponseWriter, r *http.Request) {
	// perform some logic
	stringMap := make(map[string]string)
	stringMap["test"] = "Hello, again."

	remoteIP := h.App.Session.GetString(r.Context(), "remote_ip")
	stringMap["remote_ip"] = remoteIP

	// send the data to the template
	h.Renderer.RenderTemplate(w, "about.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
	})
}
