package main

import (
	"path/filepath"
	"text/template"
	"time"

	"github.com/fayazp088/snippet-box/internal/models"
)

type templateData struct {
	CurrentYear int
	Snippet     models.Snippet
	Snippets    []models.Snippet
	Form any
}

var functions = template.FuncMap{
	"humanDate": humanDate,
}

func humanDate(t time.Time) string {
	return t.Format("02 Jan 2006 at 15:04")
}

func templateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := filepath.Glob("./ui/html/pages/*.gohtml")

	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		ts, err := template.New(name).Funcs(functions).ParseFiles("./ui/html/base.gohtml")

		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseGlob("./ui/html/partials/*.gohtml")

		if err != nil {
			return nil, err
		}

		tmpl, err := ts.ParseFiles(page)

		if err != nil {
			return nil, err
		}

		cache[name] = tmpl
	}

	return cache, nil
}
