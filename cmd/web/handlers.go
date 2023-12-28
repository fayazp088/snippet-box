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

type userSignupForm struct {
	Name                string `form:"name"`
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

type userLoginForm struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
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

	a.sessionManager.Put(r.Context(), "flash", "Snippet successfully created!")

	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
}

func (a *Application) snippetCreate(w http.ResponseWriter, r *http.Request) {
	data := a.newTemplateData(r)
	data.Form = snippetCreateForm{
		Expires: 7,
	}
	a.render(w, r, http.StatusOK, "create.gohtml", data)
}

func (a *Application) userSignup(w http.ResponseWriter, r *http.Request) {
	data := a.newTemplateData(r)
	data.Form = userSignupForm{}
	a.render(w, r, http.StatusOK, "signup.gohtml", data)
}
func (a *Application) userSignupPost(w http.ResponseWriter, r *http.Request) {
	var form userSignupForm

	err := a.decodePostForm(r, &form)

	if err != nil {
		a.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(!validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.Password, 8), "password", "This field must be at least 8 characters long")

	if !form.Valid() {
		data := a.newTemplateData(r)
		data.Form = form
		a.render(w, r, http.StatusUnprocessableEntity, "signup.gohtml", data)
		return
	}

	err = a.users.Insert(form.Name, form.Email, form.Password)

	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			form.AddFieldError("email", "Email address is already in use")
			data := a.newTemplateData(r)
			data.Form = form
			a.render(w, r, http.StatusUnprocessableEntity, "signup.gohtml", data)
		} else {
			a.serverError(w, r, err)
		}
		return
	}

	a.sessionManager.Put(r.Context(), "flash", "Your signup was successful. Please log in.")
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}
func (a *Application) userLogin(w http.ResponseWriter, r *http.Request) {
	data := a.newTemplateData(r)
	data.Form = userLoginForm{}
	a.render(w, r, http.StatusOK, "login.gohtml", data)
}
func (a *Application) userLoginPost(w http.ResponseWriter, r *http.Request) {
	var form userLoginForm

	err := a.decodePostForm(r, &form)
	if err != nil {
		a.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(!validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")

	if !form.Valid() {
		data := a.newTemplateData(r)
		data.Form = form
		a.render(w, r, http.StatusUnprocessableEntity, "login.gohtml", data)
		return
	}

	id, err := a.users.Authenticate(form.Email, form.Password)

	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.AddNonFieldError("Email or password is incorrect")
			data := a.newTemplateData(r)
			data.Form = form
			a.render(w, r, http.StatusUnprocessableEntity, "login.gohtml", data)
		} else {
			a.serverError(w, r, err)
		}
		return
	}

	err = a.sessionManager.RenewToken(r.Context())
	if err != nil {
		a.serverError(w, r, err)
		return
	}

	a.sessionManager.Put(r.Context(), "authenticatedUserID", id)

	http.Redirect(w, r, "/snippet/create", http.StatusSeeOther)
}
func (a *Application) userLogoutPost(w http.ResponseWriter, r *http.Request) {
	err := a.sessionManager.RenewToken(r.Context())
	if err != nil {
		a.serverError(w, r, err)
		return
	}

	a.sessionManager.Remove(r.Context(), "authenticatedUserID")

	a.sessionManager.Put(r.Context(), "flash", "You've been logged out successfully!")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
