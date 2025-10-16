package renderer

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
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

func (r *Renderer) LoadTemplates() error {
	slog.Default().Info("loading templates...")

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
		tmpl := template.Must(template.New(name).ParseFS(r.fs, t))
		r.templates[name] = tmpl
	}

	// Parse all layouts + partials
	baseTemplates := template.Must(
		template.New("tmpl").Funcs(r.funcMap).ParseFS(r.fs, append(layouts, partials...)...),
	)

	// For each page, clone layouts and add the specific page
	for _, page := range pages {
		name := strings.TrimSuffix(filepath.Base(page), filepath.Ext(filepath.Base(page)))
		slog.Default().Debug("page template parsed", "name", name)
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
