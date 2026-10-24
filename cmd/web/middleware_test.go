package main

import (
    "bytes"
    "io"
    "net/http"
    "net/http/httptest"
    "testing"

    "snippetbox.dheeraj.net/internal/assert"
)

func TestCommonHeaders(t *testing.T) {
    // Initialize a new httptest.ResponseRecorder and dummy http.Request.
    rr := httptest.NewRecorder()

    req, err := http.NewRequest(http.MethodGet, "/", nil)
    if err != nil {
        t.Fatal(err)
    }

    next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("OK"))
    })

    commonHeaders(next).ServeHTTP(rr, req)

    res := rr.Result()
    defer res.Body.Close()

    expectedValue := "default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com"
    assert.Equal(t, res.Header.Get("Content-Security-Policy"), expectedValue)

    expectedValue = "origin-when-cross-origin"
    assert.Equal(t, res.Header.Get("Referrer-Policy"), expectedValue)

    expectedValue = "nosniff"
    assert.Equal(t, res.Header.Get("X-Content-Type-Options"), expectedValue)

    expectedValue = "deny"
    assert.Equal(t, res.Header.Get("X-Frame-Options"), expectedValue)

    expectedValue = "0"
    assert.Equal(t, res.Header.Get("X-XSS-Protection"), expectedValue)

    expectedValue = "Go"
    assert.Equal(t, res.Header.Get("Server"), expectedValue)

    assert.Equal(t, res.StatusCode, http.StatusOK)

    body, err := io.ReadAll(res.Body)
    if err != nil {
        t.Fatal(err)
    }
    body = bytes.TrimSpace(body)
    
    assert.Equal(t, string(body), "OK")
}