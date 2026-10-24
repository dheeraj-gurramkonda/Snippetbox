package main

import (
	"html/template"
	"io/fs"
	"path/filepath"
	"time"

	"snippetbox.dheeraj.net/assets"
	"snippetbox.dheeraj.net/internal/models"
)

// templateData contains the values used by page and layout templates.
type templateData struct {
	CurrentYear int
	Snippet     models.Snippet
	Snippets    []models.Snippet
	Form        any
	Flash       string
	IsAuthenticated bool
	CSRFToken string
}

// humanDate formats stored timestamps in UTC for consistent display, while
// leaving zero values blank for templates that have no date to show.
func humanDate(t time.Time) string {
	if t.IsZero() {
        return ""
    }

    return t.UTC().Format("02 Jan 2006 at 15:04")
}

// functions are available to templates when the cache is parsed.
var functions = template.FuncMap{
	"humanDate": humanDate,
}

// newTemplateCache parses each page together with the shared layout and partials.
// Templates are cached at startup so parsing is not repeated for each request.
func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := fs.Glob(assets.Files, "html/pages/*.tmpl")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		// Use the page's base filename as the cache key used by render.
		name := filepath.Base(page)

		// Each page is composed with the common base layout and reusable partials.
		patterns := []string{
			"html/base.tmpl",
			"html/partials/*.tmpl",
			page,
		}

		ts, err := template.New(name).Funcs(functions).ParseFS(assets.Files, patterns...)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}
