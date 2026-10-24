package main

import (
	"bytes"
	"html"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"

	"snippetbox.dheeraj.net/internal/models/mocks"

	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
)

const testServerBaseURL = "https://testserver"

func newTestApplication(t *testing.T) *application {
    templateCache, err := newTemplateCache()
    if err != nil {
        t.Fatal(err)
    }

	formDecoder := form.NewDecoder()

	sessionManager := scs.New()
    sessionManager.Lifetime = 12 * time.Hour
    sessionManager.Cookie.Secure = true

	return &application{
        logger:         slog.New(slog.DiscardHandler),
        snippets:       &mocks.SnippetModel{}, 
        users:          &mocks.UserModel{},   
        templateCache:  templateCache,
        formDecoder:    formDecoder,
        sessionManager: sessionManager,
    }
}

type testServer struct {
    *httptest.Server
}

func newTestServer(t *testing.T, h http.Handler) *testServer {
    ts := httptest.NewTestServer(t, h)

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}

	ts.Client().Jar = jar

	ts.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
        return http.ErrUseLastResponse
    }


    return &testServer{ts}
}

func (ts *testServer) resetClientCookieJar(t *testing.T) {
    jar, err := cookiejar.New(nil)
    if err != nil {
        t.Fatal(err)
    }

    ts.Client().Jar = jar
}


type testResponse struct {
    status  int
    headers http.Header
    cookies []*http.Cookie
    body    string
}

func (ts *testServer) get(t *testing.T, urlPath string) testResponse {
    req, err := http.NewRequest(http.MethodGet, testServerBaseURL+urlPath, nil)
    if err != nil {
        t.Fatal(err)
    }

    res, err := ts.Client().Do(req)
    if err != nil {
        t.Fatal(err)
    }
    defer res.Body.Close()

    body, err := io.ReadAll(res.Body)
    if err != nil {
        t.Fatal(err)
    }

    return testResponse{
        status:  res.StatusCode,
        headers: res.Header,
        cookies: res.Cookies(),
        body:    string(bytes.TrimSpace(body)),
    }
}

func extractCSRFToken(t *testing.T, body string) string {
    csrfTokenRX := regexp.MustCompile(`<input type='hidden' name='csrf_token' value='(.+)'>`)
    matches := csrfTokenRX.FindStringSubmatch(body)
    if len(matches) < 2 {
        t.Fatal("no csrf token found in body")
    }
    return html.UnescapeString(matches[1])
}

func (ts *testServer) postForm(t *testing.T, urlPath string, form url.Values) testResponse {
    req, err := http.NewRequest(http.MethodPost, testServerBaseURL+urlPath, strings.NewReader(form.Encode()))
    if err != nil {
        t.Fatal(err)
    }

    // Set the appropriate Content-Type header for form data and the Sec-Fetch-Site
    // header.
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    req.Header.Set("Sec-Fetch-Site", "same-origin")

    res, err := ts.Client().Do(req)
    if err != nil {
        t.Fatal(err)
    }
    defer res.Body.Close()

    body, err := io.ReadAll(res.Body)
    if err != nil {
        t.Fatal(err)
    }

    return testResponse{
        status:  res.StatusCode,
        headers: res.Header,
        cookies: res.Cookies(),
        body:    string(bytes.TrimSpace(body)),
    }
}