package controllers

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"

	"github.com/eirka/eirka-libs/user"
)

// Common test utilities for all controller tests

// Helper for standard requests
func performRequest(r http.Handler, method, path string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)
	req.Header.Set("X-Real-IP", "127.0.0.1")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// Helper for JSON requests
func performJSONRequest(r http.Handler, method, path string, body []byte) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Real-IP", "127.0.0.1")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// Format error and success messages
func errorMessage(err error) string {
	return fmt.Sprintf(`{"error_message":"%s"}`, err.Error())
}

func successMessage(message string) string {
	return fmt.Sprintf(`{"success_message":"%s"}`, message)
}

// Middleware that mocks an authenticated admin user
func mockAdminMiddleware(params []uint) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set the required context values for controllers
		c.Set("userdata", user.User{ID: 2}) // Non-anonymous user
		c.Set("protected", true)
		c.Set("params", params)
		c.Next()
	}
}

// Non-protected middleware
func mockNonAdminMiddleware(params []uint) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("userdata", user.User{ID: 2})
		c.Set("protected", false)
		c.Set("params", params)
		c.Next()
	}
}