package controllers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestUptimeController(t *testing.T) {
	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Set up routes
	router.GET("/uptime", UptimeController)

	// Store the original startTime and restore it after test
	originalStartTime := startTime
	defer func() { startTime = originalStartTime }()

	// Set start time to a known value (10 minutes ago)
	startTime = time.Now().Add(-10 * time.Minute)

	// Create request
	req, _ := http.NewRequest("GET", "/uptime", nil)
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Check response code
	assert.Equal(t, http.StatusOK, w.Code, "HTTP status code should be 200")

	// Check response contains uptime
	assert.Contains(t, w.Body.String(), "uptime", "Response should contain uptime field")
	assert.Contains(t, w.Body.String(), "10m", "Response should show uptime of 10m")
}