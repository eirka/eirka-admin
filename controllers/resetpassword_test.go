package controllers

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	e "github.com/eirka/eirka-libs/errors"
)

// Mock functions
// Since user.RandomPassword and user.UpdatePassword are in an external library,
// we would need to mock them using a testing framework with function patching
// or use dependency injection. Since that's not available in this codebase,
// these tests will focus on the aspects we can test without mocking these functions.

func TestResetPasswordControllerNotProtected(t *testing.T) {
	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(mockNonAdminMiddleware([]uint{1}))
	router.POST("/resetpassword", ResetPasswordController)

	// Create JSON request
	jsonRequest := []byte(`{"uid":1}`)
	
	// Perform the request
	response := performJSONRequest(router, "POST", "/resetpassword", jsonRequest)

	// Check response code
	assert.Equal(t, http.StatusInternalServerError, response.Code, "HTTP status code should be 500")
	
	// Check response body
	assert.JSONEq(t, errorMessage(e.ErrInternalError), response.Body.String(), "Response should match expected error message")
}

func TestResetPasswordControllerInvalidParam(t *testing.T) {
	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(mockAdminMiddleware([]uint{1}))
	router.POST("/resetpassword", ResetPasswordController)

	// Create invalid JSON request (missing required "uid" field)
	jsonRequest := []byte(`{}`)
	
	// Perform the request
	response := performJSONRequest(router, "POST", "/resetpassword", jsonRequest)

	// Check response code
	assert.Equal(t, http.StatusBadRequest, response.Code, "HTTP status code should be 400")
	
	// Check response body
	assert.JSONEq(t, errorMessage(e.ErrInvalidParam), response.Body.String(), "Response should match expected error message")
}

// Note: Testing the actual password reset functionality would require being able to mock
// user.RandomPassword() and user.UpdatePassword() functions, which would require
// refactoring the controller to use interfaces or dependency injection.
// The tests above verify the middleware checks and parameter validation aspects of the controller.