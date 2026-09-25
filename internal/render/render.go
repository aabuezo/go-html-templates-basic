package render

import (
	"bytes"
	"errors"
	"html/template"
	"log"
	"myapp/internal/config"
	"myapp/internal/models"
	"net/http"
	"path/filepath"
)

type Renderer struct {
	app *config.AppConfig
}

// NewRenderer sets the config for the template package
func NewRenderer(app *config.AppConfig) *Renderer {
	return &Renderer{app: app}
}

func AddDefaultData(td *models.TemplateData) *models.TemplateData {

	return td
}

// RenderTemplate renders a template
func (r *Renderer) RenderTemplate(w http.ResponseWriter, tmpl string, td *models.TemplateData) {
	var tc map[string]*template.Template
	if r.app.UseCache {
		// get the template cache from the app config
		tc = r.app.TemplateCache
	} else {
		tc, _ = CreateTemplateCache()
	}

	// get the template from cache
	t, ok := tc[tmpl]
	if !ok {
		log.Fatal("could not get template from template cache")
	}

	buf := new(bytes.Buffer)

	td = AddDefaultData(td)

	err := t.Execute(buf, td)
	if err != nil {
		log.Println(err)
	}

	// render the template
	_, err = buf.WriteTo(w)
	if err != nil {
		log.Println(err)
	}
}

func CreateTemplateCache() (map[string]*template.Template, error) {
	myCache := map[string]*template.Template{}

	// get all of the files named *.page.tmpl from ./templates
	pages, err := filepath.Glob("./templates/*.page.tmpl")
	if err != nil {
		return myCache, err
	}

	if len(pages) == 0 {
		return myCache, errors.New("could not find any *.page.tmpl")
	}

	// get all of the files named *.layout.tmpl from ./templates
	layouts, err := filepath.Glob("./templates/*.layout.tmpl")
	if err != nil {
		return myCache, err
	}

	if len(layouts) == 0 {
		return myCache, errors.New("could not find any *.layout.tmpl")
	}

	// range through all files ending with *.page.tmpl
	for _, page := range pages {
		name := filepath.Base(page)
		ts, err := template.New(name).ParseFiles(page)
		if err != nil {
			return myCache, err
		}

		ts, err = ts.ParseFiles(layouts...)
		if err != nil {
			return myCache, err
		}

		myCache[name] = ts
	}

	return myCache, nil
}

// var (
// 	tc   = make(map[string]*template.Template)
// 	tcMu sync.Mutex
// )

// func RenderTemplate(w http.ResponseWriter, t string) {
// 	tcMu.Lock()

// 	tmpl, inMap := tc[t]
// 	if !inMap {
// 		// create template cache
// 		log.Println("creating template and adding to cache")

// 		var err error
// 		tmpl, err = createTemplate(t)
// 		if err != nil {
// 			tcMu.Unlock()
// 			http.Error(w, "error loading template", http.StatusInternalServerError)
// 			log.Println(err)
// 			return
// 		}

// 		// add template to cache (map)
// 		tc[t] = tmpl
// 	} else {
// 		log.Println("using cached template")
// 	}

// 	tcMu.Unlock()

// 	if err := tmpl.Execute(w, nil); err != nil {
// 		log.Println(err)
// 	}
// }

// func createTemplate(t string) (*template.Template, error) {
// 	templates := []string{
// 		fmt.Sprintf("./templates/%s", t),
// 		"./templates/base.layout.tmpl",
// 	}

// 	// parse template
// 	return template.ParseFiles(templates...)
// }
