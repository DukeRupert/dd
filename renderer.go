package main

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

//go:embed templates/*
var templateFS embed.FS

// TemplateRenderer handles HTML template rendering
type TemplateRenderer struct {
    templates map[string]*template.Template
    funcMap   template.FuncMap
}

func NewTemplateRenderer() *TemplateRenderer {
    return &TemplateRenderer{
        templates: make(map[string]*template.Template),
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

func (tr *TemplateRenderer) LoadTemplates() error {
    layouts, err := fs.Glob(templateFS, "templates/layouts/*.html")
    if err != nil {
        return err
    }
    
    pages, err := fs.Glob(templateFS, "templates/pages/*.html")
    if err != nil {
        return err
    }
    
    partials, err := fs.Glob(templateFS, "templates/partials/*.html")
    if err != nil {
        return err
    }
    
    for _, page := range pages {
        // home.html -> home
        pageName := filepath.Base(page)
        name := strings.TrimSuffix(pageName, filepath.Ext(pageName))
        
        // files = layouts + page + partials
        files := append(layouts, page)
        files = append(files, partials...)
        
        tmpl := template.Must(
            template.New(pageName).Funcs(tr.funcMap).ParseFiles(files...),
        )
        
        // Debug: Print all defined templates
        fmt.Printf("Templates defined for page '%s':\n", name)
        for _, t := range tmpl.Templates() {
            fmt.Printf("  - %s\n", t.Name())
        }
        fmt.Println()
        
        tr.templates[name] = tmpl
    }
    
    return nil
}

func (tr *TemplateRenderer) Render(w http.ResponseWriter, name string, data interface{}) error {
   tmpl, exists := tr.templates[name]
    if !exists {
        return fmt.Errorf("template %s not found", name)
    }
    
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    
    // Debug: try different template names
    fmt.Printf("Available templates for %s: ", name)
    for _, t := range tmpl.Templates() {
        fmt.Printf("%s ", t.Name())
    }
    fmt.Println()
    
    // Try executing the layout template instead of the page name
    return tmpl.ExecuteTemplate(w, "layout", data)
}