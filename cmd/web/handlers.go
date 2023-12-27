package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/fayazp088/snippet-box/internal/models"
	"github.com/fayazp088/snippet-box/internal/validator"
	"github.com/julienschmidt/httprouter"
)

type snippetCreateForm struct {
	Title               string `form:"title"`
	Content             string `form:"content"`
	Expires             int    `form:"expires"`
	validator.Validator `form:"-"`
}

func (a *Application) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		a.notFound(w)
		return
	}

	snippets, err := a.snippets.Latest()
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	data := a.newTemplateData(r)
	data.Snippets = snippets
	a.render(w, r, http.StatusOK, "home.gohtml", data)
}

func (a *Application) snippetView(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	log.Print("Id", id)
	if err != nil || id < 1 {
		a.notFound(w)
		return
	}

	snippet, err := a.snippets.Get(id)

	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			a.notFound(w)
		} else {
			a.serverError(w, r, err)
		}
		return
	}
	data := a.newTemplateData(r)
	data.Snippet = snippet

	a.render(w, r, http.StatusOK, "view.gohtml", data)
}

func (a *Application) snippetCreatePost(w http.ResponseWriter, r *http.Request) {
	var form snippetCreateForm
	err := a.decodePostForm(r, &form)
	if err != nil {
		a.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Title), "title", "This field cannot be blank")
	form.CheckField(validator.MaxChars(form.Title, 100), "title", "This field cannot be more than 100 characters long")
	form.CheckField(validator.NotBlank(form.Content), "content", "This field cannot be blank")
	form.CheckField(validator.PermittedValue(form.Expires, 1, 7, 365), "expires", "This field must equal 1, 7 or 365")

	if !form.Valid() {
		data := a.newTemplateData(r)
		data.Form = form
		a.render(w, r, http.StatusUnprocessableEntity, "create.gohtml", data)
		return
	}

	id, err := a.snippets.Insert(form.Title, form.Content, form.Expires)

	if err != nil {
		a.serverError(w, r, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
}

func (a *Application) snippetCreate(w http.ResponseWriter, r *http.Request) {
	data := a.newTemplateData(r)
	data.Form = snippetCreateForm{
		Expires: 7,
	}
	a.render(w, r, http.StatusOK, "create.gohtml", data)
}
