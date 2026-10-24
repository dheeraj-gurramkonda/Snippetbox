package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"snippetbox.dheeraj.net/internal/models"
	"snippetbox.dheeraj.net/internal/validator"
)

// Form types bind submitted fields to Go values and carry their validation
// errors back to the templates for display.
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

// home displays the most recently created, non-expired snippets.
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	snippets, err := app.snippets.Latest()
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data := app.newTemplateData(r)
	data.Snippets = snippets

	err = app.render(w, r, http.StatusOK, "home.tmpl", data)
	if err != nil {
		app.serverError(w, r, err)
	}
}

// snippetView validates the path ID, loads the matching snippet, and renders it.
func (app *application) snippetView(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	snippet, err := app.snippets.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrNoRecord):
			http.NotFound(w, r)
		default:
			app.serverError(w, r, err)
		}
		return
	}

	data := app.newTemplateData(r)
	data.Snippet = snippet

	err = app.render(w, r, http.StatusOK, "view.tmpl", data)
	if err != nil {
		app.serverError(w, r, err)
	}
}

// snippetCreate shows the HTML form users fill in to create a new snippet.
func (app *application) snippetCreate(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = snippetCreateForm{
		Expires: 365,
	}
	err := app.render(w, r, http.StatusOK, "create.tmpl", data)
	if err != nil {
		app.serverError(w, r, err)
	}
}

// snippetCreatePost accepts the submitted snippet data, saves it, and redirects to the new snippet page.
func (app *application) snippetCreatePost(w http.ResponseWriter, r *http.Request) {
	var form snippetCreateForm

	// Reject malformed form data before applying field-level validation.
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Title), "title", "This field cannot be blank")
	form.CheckField(validator.MaxChars(form.Title, 100), "title", "This field cannot be more than 100 characters long")
	form.CheckField(validator.NotBlank(form.Content), "content", "This field cannot be blank")
	form.CheckField(validator.PermittedValue(form.Expires, 1, 7, 365), "expires", "This field must equal 1, 7 or 365")

	// Re-render invalid submissions so entered values and field errors survive.
	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form

		err := app.render(w, r, http.StatusUnprocessableEntity, "create.tmpl", data)
		if err != nil {
			app.serverError(w, r, err)
		}
		return
	}

	// After persistence, show a one-time confirmation on the destination page.
	id, err := app.snippets.Insert(form.Title, form.Content, form.Expires)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Snippet successfully created!")

	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
}

// userSignup renders a blank account-registration form.
func (app *application) userSignup(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userSignupForm{}

	err := app.render(w, r, http.StatusOK, "signup.tmpl", data)
	if err != nil {
		app.serverError(w, r, err)
	}
}

// userSignupPost validates registration fields, creates the account, and
// redirects the new user to the login page.
func (app *application) userSignupPost(w http.ResponseWriter, r *http.Request) {
	var form userSignupForm

	// Decode first so malformed requests are separated from ordinary validation
	// errors that can be shown next to form fields.
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.Password, 8), "password", "This field must be at least 8 characters long")
	form.CheckField(validator.MaxBytes(form.Password, 72), "password", "This field must not be more than 72 bytes long")

	// Keep validation errors beside the submitted values for correction.
	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form

		err := app.render(w, r, http.StatusUnprocessableEntity, "signup.tmpl", data)
		if err != nil {
			app.serverError(w, r, err)
		}
		return
	}

	err = app.users.Insert(form.Name, form.Email, form.Password)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrDuplicateEmail):
			// Translate the model's duplicate-key result into a form-level error.
			form.AddFieldError("email", "Email address is already in use")
			data := app.newTemplateData(r)
			data.Form = form

			err = app.render(w, r, http.StatusUnprocessableEntity, "signup.tmpl", data)
			if err != nil {
				app.serverError(w, r, err)
			}
		default:
			app.serverError(w, r, err)
		}
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Your signup was successful. Please log in.")

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

// userLogin renders a blank login form.
func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userLoginForm{}

	err := app.render(w, r, http.StatusOK, "login.tmpl", data)
	if err != nil {
		app.serverError(w, r, err)
	}
}

// userLoginPost validates credentials, starts a fresh authenticated session,
// and sends the user to the snippet creation page.
func (app *application) userLoginPost(w http.ResponseWriter, r *http.Request) {
	var form userLoginForm

	// Malformed form encoding is a bad request; incorrect credentials are
	// handled as a regular form result below.
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	form.CheckField(validator.MaxBytes(form.Password, 72), "password", "This field must not be more than 72 bytes long")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form

		err := app.render(w, r, http.StatusUnprocessableEntity, "login.tmpl", data)
		if err != nil {
			app.serverError(w, r, err)
		}
		return
	}

	id, err := app.users.Authenticate(form.Email, form.Password)
	if err != nil {
		switch {
        case errors.Is(err, models.ErrInvalidCredentials):
			// Use one generic message so the response does not reveal whether
			// the email address or password was incorrect.
            form.AddNonFieldError("Email or password is incorrect")
            data := app.newTemplateData(r)
            data.Form = form

            err := app.render(w, r, http.StatusUnprocessableEntity, "login.tmpl", data)
            if err != nil {
                app.serverError(w, r, err)
            }
        default:
            app.serverError(w, r, err)
        }
        return
	}

	// Rotate the session token before marking the session authenticated.
	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "authenticatedUserID", id)

	http.Redirect(w, r, "/snippet/create", http.StatusSeeOther)
}

func (app *application) userLogoutPost(w http.ResponseWriter, r *http.Request) {
	// Rotate the token on logout as well, before clearing the authenticated ID.
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Remove(r.Context(), "authenticatedUserID")
	app.sessionManager.Put(r.Context(), "flash", "You've been logged out successfully!")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// ping is a minimal endpoint for checking whether the HTTP server is responding.
func ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}
