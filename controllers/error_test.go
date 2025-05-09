package controllers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	e "github.com/eirka/eirka-libs/errors"
)

// Format error messages like the API
func errorMessageJSON(err error) string {
	return fmt.Sprintf(`{"error_message":"%s"}`, err)
}

func TestErrorController(t *testing.T) {
	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Set up routes
	router.GET("/error", ErrorController)

	// Create request
	req, _ := http.NewRequest("GET", "/error", nil)
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Check assertions
	assert.Equal(t, http.StatusNotFound, w.Code, "HTTP request code should match")
	assert.JSONEq(t, w.Body.String(), errorMessageJSON(e.ErrNotFound), "HTTP response should match")
}