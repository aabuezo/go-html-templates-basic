package handlers

import (
	"log"
	"net/http"
	"strings"

	"myapp/internal/config"
	"myapp/internal/forms"
	"myapp/internal/helpers"
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
	h.Renderer.RenderTemplate(w, r, "home.page.tmpl", &models.TemplateData{})
}

// About is the about page handler
func (h *Handler) About(w http.ResponseWriter, r *http.Request) {
	h.Renderer.RenderTemplate(w, r, "about.page.tmpl", &models.TemplateData{})
}

// Form shows the login page
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var emptyLogin models.Login
	data := make(map[string]any)
	data["login"] = emptyLogin

	h.Renderer.RenderTemplate(w, r, "login.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
		Data: data,
	})
}

// PostForm handles the posting of a login
func (h *Handler) LoginForm(w http.ResponseWriter, r *http.Request) {
	// Parse the form
	err := r.ParseForm()

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// Fill the object
	login := models.Login{
		Name:     r.Form.Get("name"),
		Email:    r.Form.Get("email"),
		Password: r.Form.Get("password"),
	}

	// Create a form validator
	form := forms.New(r.PostForm)

	form.Required("name", "email", "password")
	form.MinLength("name", 3)
	form.CheckEmail("email")
	form.CheckPassword("password")

	// Check if form is valid
	if !form.Valid() {
		data := make(map[string]any)
		data["login"] = login

		h.Renderer.RenderTemplate(w, r, "login.page.tmpl", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	// Add object to the session
	h.App.Session.Put(r.Context(), "user_name", login.Name)

	// If form is valid, redirect
	http.Redirect(w, r, "/user-home", http.StatusSeeOther)
}

func (h *Handler) UserHome(w http.ResponseWriter, r *http.Request) {
	userName := strings.TrimSpace(
		h.App.Session.GetString(r.Context(), "user_name"),
	)

	if userName == "" {
		log.Println("someone tried to acces user-home without logging in")
		h.App.Session.Put(r.Context(), "error", "Not a logged-in user")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	h.Renderer.RenderTemplate(w, r, "user-home.page.tmpl", &models.TemplateData{
		StringMap: map[string]string{
			"user_name": userName,
		},
	})
}
