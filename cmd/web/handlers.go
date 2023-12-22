package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

func (a *Application) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		a.notFound(w)
		return
	}

	files := []string{
		"./ui/html/base.gohtml",
		"./ui/html/pages/home.gohtml",
		"./ui/html/partials/nav.gohtml",
	}

	tmpl, err := template.ParseFiles(files...)

	if err != nil {
		a.logger.Error(err.Error())
		a.serverError(w, r, err)
		return
	}

	err = tmpl.ExecuteTemplate(w, "base", nil)

	if err != nil {
		a.logger.Error(err.Error())
		a.serverError(w, r, err)
	}
}

func (a *Application) snippetView(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id < 1 {
		a.notFound(w)
		return
	}
	fmt.Fprintf(w, "Display a specific snippet with ID %d...", id)
}

func (a *Application) snippetCreate(w http.ResponseWriter, r *http.Request) {
	log.Print("hello snippetCreate")
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		a.clientError(w, http.StatusMethodNotAllowed)
		return
	}

	title := "O snail"
	content := "O snail\nClimb Mount Fuji,\nBut slowly, slowly!\n\n– Kobayashi Issa"
	expires := 7

	id, err := a.snippets.Insert(title, content, expires)

	if err != nil {
		a.serverError(w, r, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/snippet/view?id=%d", id), http.StatusSeeOther)
}
