package main

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-playground/form/v4"
	"github.com/justinas/nosurf"
)

// serverError logs the request error and responds with a generic 500 error page.
func (app *application) serverError(w http.ResponseWriter, r *http.Request, err error) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
	)

	app.logger.Error(err.Error(), slog.String("method", method), slog.String("uri", uri))
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// clientError writes a standard HTTP error response for the given status code.
func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

// render executes a cached page into a buffer before writing the response.
// Buffering prevents a partially rendered page from being sent if execution fails.
func (app *application) render(w http.ResponseWriter, r *http.Request, status int, page string, data templateData) error {
	ts, ok := app.templateCache[page]
	if !ok {
		return fmt.Errorf("the template %s does not exist in the cache", page)
	}

	buf := new(bytes.Buffer)
	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		return err
	}

	w.WriteHeader(status)
	_, err = buf.WriteTo(w)

	return err
}

// newTemplateData fills fields shared by all pages, including one-time session
// messages and request-specific authentication and CSRF state.
func (app *application) newTemplateData(r *http.Request) templateData {
	return templateData{
		CurrentYear: time.Now().Year(),
		Flash:       app.sessionManager.PopString(r.Context(), "flash"),
		IsAuthenticated: app.isAuthenticated(r),
		CSRFToken: nosurf.Token(r),
	}
}

// decodePostForm parses URL-encoded request data and maps it onto dst.
func (app *application) decodePostForm(r *http.Request, dst any) error {
	err := r.ParseForm()
	if err != nil {
		return err
	}

	err = app.formDecoder.Decode(dst, r.PostForm)
	if err != nil {
		if _, ok := errors.AsType[*form.InvalidDecoderError](err); ok {
			// An invalid decoder configuration is a programming error, rather
			// than a malformed user submission.
			panic(err)
		}
		return err
	}
	return nil
}

// isAuthenticated reads the flag set by the authentication middleware.
func (app *application) isAuthenticated(r *http.Request) bool {
	isAuthenticated, ok := r.Context().Value(contextKeyIsAuthenticated).(bool)
	if !ok {
		return false
	}
	
	return isAuthenticated
}
