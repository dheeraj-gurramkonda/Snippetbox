package main

// contextKey gives authentication state a private, package-specific context key.
type contextKey string

// contextKeyIsAuthenticated stores whether the session maps to an existing user.
const contextKeyIsAuthenticated = contextKey("isAuthenticated")
