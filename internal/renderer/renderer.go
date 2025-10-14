package renderer

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

// Renderer handles HTML template rendering
type Renderer struct {
	templates map[string]*template.Template
	funcMap   template.FuncMap
	fs        embed.FS
}

// New creates a new template renderer
func New(templateFS embed.FS) *Renderer {
	return &Renderer{
		templates: make(map[string]*template.Template),
		fs:        templateFS,
		funcMap: template.FuncMap{
			"safeHTML": func(s string) template.HTML {
				return template.HTML(s)
			},
			"formatDate": func(t time.Time) string {
				return t.Format("2006-01-02")
			},
		},
	}
}

// LoadTemplates loads all templates from the embedded filesystem
func (r *Renderer) LoadTemplates() error {
	fmt.Println("LoadTemplates()")
	layouts, err := fs.Glob(r.fs, "layouts/*.html")
	if err != nil {
		return err
	}

	pages, err := fs.Glob(r.fs, "pages/*.html")
	if err != nil {
		return err
	}

	partials, err := fs.Glob(r.fs, "partials/*.html")
	if err != nil {
		return err
	}

	for _, page := range pages {
		// home.html -> home
		pageName := filepath.Base(page)
		name := strings.TrimSuffix(pageName, filepath.Ext(pageName))

		// DEBUG
		fmt.Printf("Loading template: %s from file: %s\n", name, page)

		// files = layouts + page + partials
		files := append(layouts, page)
		files = append(files, partials...)

		tmpl := template.Must(
			template.New(pageName).Funcs(r.funcMap).ParseFS(r.fs, files...),
		)

		r.templates[name] = tmpl
	}

	return nil
}

// Render renders a template with the given data
func (r *Renderer) Render(w http.ResponseWriter, name string, data interface{}) error {
	tmpl, exists := r.templates[name]
	if !exists {
		return fmt.Errorf("template %s not found", name)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// Execute base.html or minimal.html, not the page template
	// The page will be included via the layout's {{block}} directives
	if tmpl.Lookup("base.html") != nil {
		return tmpl.ExecuteTemplate(w, "base.html", data)
	}
	if tmpl.Lookup("minimal.html") != nil {
		return tmpl.ExecuteTemplate(w, "minimal.html", data)
	}
	
	return tmpl.Execute(w, data)
}

// RenderPartial renders a partial template by name from any loaded template set
func (r *Renderer) RenderPartial(w http.ResponseWriter, partialName string, data interface{}) error {
	// Use any template set since they all have the same partials loaded
	for _, tmpl := range r.templates {
		if tmpl.Lookup(partialName) != nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			return tmpl.ExecuteTemplate(w, partialName, data)
		}
	}
	return fmt.Errorf("partial %s not found", partialName)
}
