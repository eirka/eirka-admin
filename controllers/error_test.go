package controllers

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	e "github.com/eirka/eirka-libs/errors"
)

func TestErrorController(t *testing.T) {
	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Set up routes
	router.GET("/error", ErrorController)

	// Perform the request
	response := performRequest(router, "GET", "/error")

	// Check response code
	assert.Equal(t, http.StatusNotFound, response.Code, "HTTP status code should be 404")

	// Check response body
	assert.JSONEq(t, errorMessage(e.ErrNotFound), response.Body.String(), "Response should match expected error message")
}
