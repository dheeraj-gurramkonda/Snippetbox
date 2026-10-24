package main

import (
	"net/http"
	"strings"
    "net/url"
	"testing"

	"snippetbox.dheeraj.net/internal/assert"
)

func TestPing(t *testing.T) {
    app := newTestApplication(t)

    ts := newTestServer(t, app.routes())
    defer ts.Close()

    res := ts.get(t, "/ping")
    assert.Equal(t, res.status, http.StatusOK)
    assert.Equal(t, res.body, "OK")
}

func TestSnippetView(t *testing.T) {
    app := newTestApplication(t)

    ts := newTestServer(t, app.routes())
    defer ts.Close()

    tests := []struct {
        name     string
        urlPath  string
        wantStatus int
        wantBody string
    }{
        {
            name:       "Valid ID",
            urlPath:    "/snippet/view/1",
            wantStatus: http.StatusOK,
            wantBody:   "An old silent pond...",
        },
        {
            name:      "Non-existent ID",
            urlPath:   "/snippet/view/2",
            wantStatus: http.StatusNotFound,
        },
        {
            name:      "Negative ID",
            urlPath:   "/snippet/view/-1",
            wantStatus: http.StatusNotFound,
        },
        {
            name:      "Decimal ID",
            urlPath:   "/snippet/view/1.23",
            wantStatus: http.StatusNotFound,
        },
        {
            name:      "String ID",
            urlPath:   "/snippet/view/foo",
            wantStatus: http.StatusNotFound,
        },
        {
            name:      "Empty ID",
            urlPath:   "/snippet/view/",
            wantStatus: http.StatusNotFound,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ts.resetClientCookieJar(t)

            res := ts.get(t, tt.urlPath) 
            assert.Equal(t, res.status, tt.wantStatus)
            assert.True(t, strings.Contains(res.body, tt.wantBody))
        })
    }
}

func TestUserSignup(t *testing.T) {
    app := newTestApplication(t)
    ts := newTestServer(t, app.routes())
    defer ts.Close()

    const (
        validName     = "Bob"
        validPassword = "validPa$$word"
        validEmail    = "bob@example.com"
        formTag       = "<form action='/user/signup' method='POST' novalidate>"
    )

    tests := []struct {
        name              string
        userName          string
        userEmail         string
        userPassword      string
        useValidCSRFToken bool
        wantStatus        int
        wantFormTag       string
    }{
        {
            name:              "Valid submission",
            userName:          validName,
            userEmail:         validEmail,
            userPassword:      validPassword,
            useValidCSRFToken: true,
            wantStatus:        http.StatusSeeOther,
        },
        {
            name:              "Invalid CSRF Token",
            userName:          validName,
            userEmail:         validEmail,
            userPassword:      validPassword,
            useValidCSRFToken: false,
            wantStatus:        http.StatusBadRequest,
        },
        {
            name:              "Empty name",
            userName:          "",
            userEmail:         validEmail,
            userPassword:      validPassword,
            useValidCSRFToken: true,
            wantStatus:        http.StatusUnprocessableEntity,
            wantFormTag:       formTag,
        },
        {
            name:              "Empty email",
            userName:          validName,
            userEmail:         "",
            userPassword:      validPassword,
            useValidCSRFToken: true,
            wantStatus:        http.StatusUnprocessableEntity,
            wantFormTag:       formTag,
        },
        {
            name:              "Empty password",
            userName:          validName,
            userEmail:         validEmail,
            userPassword:      "",
            useValidCSRFToken: true,
            wantStatus:        http.StatusUnprocessableEntity,
            wantFormTag:       formTag,
        },
        {
            name:              "Invalid email",
            userName:          validName,
            userEmail:         "bob@example.",
            userPassword:      validPassword,
            useValidCSRFToken: true,
            wantStatus:        http.StatusUnprocessableEntity,
            wantFormTag:       formTag,
        },
        {
            name:              "Short password",
            userName:          validName,
            userEmail:         validEmail,
            userPassword:      "pa$$",
            useValidCSRFToken: true,
            wantStatus:        http.StatusUnprocessableEntity,
            wantFormTag:       formTag,
        },
        {
            name:              "Duplicate email",
            userName:          validName,
            userEmail:         "dupe@example.com",
            userPassword:      validPassword,
            useValidCSRFToken: true,
            wantStatus:        http.StatusUnprocessableEntity,
            wantFormTag:       formTag,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ts.resetClientCookieJar(t)

            res := ts.get(t, "/user/signup")

            form := url.Values{}
            form.Add("name", tt.userName)
            form.Add("email", tt.userEmail)
            form.Add("password", tt.userPassword)
            if tt.useValidCSRFToken {
                form.Add("csrf_token", extractCSRFToken(t, res.body))
            }

            res = ts.postForm(t, "/user/signup", form)

            assert.Equal(t, res.status, tt.wantStatus)
            assert.True(t, strings.Contains(res.body, tt.wantFormTag))
        })
    }
}