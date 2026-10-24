package main

import (
	"net/http"

    "snippetbox.dheeraj.net/assets"

	"github.com/justinas/alice"
)

// routes registers public and protected endpoints, then applies the standard
// recovery, request-logging, and response-header middleware to the whole mux.
func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	// Static assets and the health check do not need a user session.
    mux.Handle("GET /static/", http.FileServerFS(assets.Files))

    mux.HandleFunc("GET /ping", ping)

    // Dynamic pages load/save sessions, enforce CSRF protection, and resolve
    // any logged-in user before reaching their route handlers.
    dynamic := alice.New(app.sessionManager.LoadAndSave, preventCSRF, app.authenticate)

    mux.Handle("GET /{$}", dynamic.ThenFunc(app.home))
    mux.Handle("GET /snippet/view/{id}", dynamic.ThenFunc(app.snippetView))
    mux.Handle("GET /user/signup", dynamic.ThenFunc(app.userSignup))
    mux.Handle("POST /user/signup", dynamic.ThenFunc(app.userSignupPost))
    mux.Handle("GET /user/login", dynamic.ThenFunc(app.userLogin))
    mux.Handle("POST /user/login", dynamic.ThenFunc(app.userLoginPost))

    // Append the access check only to routes that require a signed-in account.
    protected := dynamic.Append(app.requireAuthentication)

    mux.Handle("GET /snippet/create", protected.ThenFunc(app.snippetCreate))
    mux.Handle("POST /snippet/create", protected.ThenFunc(app.snippetCreatePost))
    mux.Handle("POST /user/logout", protected.ThenFunc(app.userLogoutPost))

    standard := alice.New(app.recoverPanic, app.logRequest, commonHeaders)
    return standard.Then(mux)
}
