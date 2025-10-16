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

	"golang.org/x/exp/slog"
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

func (r *Renderer) LoadTemplates() error {
	slog.Info("loading templates...")

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

	// Parse all layouts and partials
	templates := append(layouts, partials...)
	for _, t := range templates {
		name := strings.TrimSuffix(filepath.Base(t), filepath.Ext(filepath.Base(t)))
		// logger.Debug("Loading template: %s from file: %s\n", name, t)
		slog.Debug("parsing layouts and partials...")

		tmpl := template.Must(template.New(name).ParseFS(r.fs, t))
		r.templates[name] = tmpl
	}

	// Parse all layouts + partials
	baseTemplates := template.Must(
		template.New("tmpl").Funcs(r.funcMap).ParseFS(r.fs, append(layouts, partials...)...),
	)
	slog.Info("parsing layouts and partials...")
	slog.Debug("testing if debug level is active in renderer")

	// For each page, clone layouts and add the specific page
	for _, page := range pages {
		name := strings.TrimSuffix(filepath.Base(page), filepath.Ext(filepath.Base(page)))
		fmt.Printf("Loading page template: %s from file: %s\n", name, page)

		tmpl := template.Must(template.Must(baseTemplates.Clone()).ParseFS(r.fs, page))
		r.templates[name] = tmpl
	}

	return nil
}

func (r *Renderer) Render(w http.ResponseWriter, name string, data interface{}) error {	
	tmpl, ok := r.templates[name]
	if !ok {
		return fmt.Errorf("template %s not found", name)
	}

	return tmpl.ExecuteTemplate(w, name, data)
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

func (r *Renderer) RenderPage(w http.ResponseWriter, page string, data interface{}) {
	tmpl, ok := r.templates[page]
	if !ok {
		fmt.Errorf("template %s not found", page)
	}
    // Executes: app.html layout → page defines blocks → partials
    err := tmpl.ExecuteTemplate(w, "app.html", data)
	if err != nil {
		fmt.Errorf("failed to ExecuteTemplate", page)
	}
}

// func (r *Renderer) renderPartial(w http.ResponseWriter, partial string, data interface{}) {
//     tmpl, ok := r.templates[partial]
// 	if !ok {
// 		fmt.Errorf("template %s not found", page)
// 	}
// 	// Executes just the partial for htmx
//     err := h.templates.ExecuteTemplate(w, partial, data)
// }
