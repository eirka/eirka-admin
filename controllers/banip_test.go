package controllers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"github.com/eirka/eirka-libs/audit"
	"github.com/eirka/eirka-libs/db"
	e "github.com/eirka/eirka-libs/errors"
)

func TestBanIPController(t *testing.T) {
	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(mockAdminMiddleware([]uint{1, 1, 1}))
	router.POST("/banip", BanIPController)

	// Set up SQL mock
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Mock the Status query - successful IP lookup
	mock.ExpectQuery(`SELECT post_ip FROM threads
	    INNER JOIN posts ON threads.thread_id = posts.thread_id
	    WHERE ib_id = \? AND threads.thread_id = \? AND post_num = \? LIMIT 1`).
		WithArgs(1, 1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"post_ip"}).
			AddRow("10.0.0.1"))

	// Mock the insert query
	mock.ExpectExec("INSERT IGNORE INTO banned_ips \\(user_id,ib_id,ban_ip,ban_reason\\) VALUES \\(\\?,\\?,\\?,\\?\\)").
		WithArgs(2, 1, "10.0.0.1", "test reason").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Create JSON request
	jsonRequest := []byte(`{"reason":"test reason"}`)
	
	// Perform the request
	response := performJSONRequest(router, "POST", "/banip", jsonRequest)

	// Check response code
	assert.Equal(t, http.StatusOK, response.Code, "HTTP status code should be 200")
	
	// Check response body
	assert.JSONEq(t, successMessage(audit.AuditBanIP), response.Body.String(), "Response should match expected success message")

	// Make sure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestBanIPControllerNotProtected(t *testing.T) {
	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(mockNonAdminMiddleware([]uint{1, 1, 1}))
	router.POST("/banip", BanIPController)

	// Create JSON request
	jsonRequest := []byte(`{"reason":"test reason"}`)
	
	// Perform the request
	response := performJSONRequest(router, "POST", "/banip", jsonRequest)

	// Check response code
	assert.Equal(t, http.StatusInternalServerError, response.Code, "HTTP status code should be 500")
	
	// Check response body
	assert.JSONEq(t, errorMessage(e.ErrInternalError), response.Body.String(), "Response should match expected error message")
}

func TestBanIPControllerInvalidParam(t *testing.T) {
	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(mockAdminMiddleware([]uint{1, 1, 1}))
	router.POST("/banip", BanIPController)

	// Create invalid JSON request (missing required "reason" field)
	jsonRequest := []byte(`{}`)
	
	// Perform the request
	response := performJSONRequest(router, "POST", "/banip", jsonRequest)

	// Check response code
	assert.Equal(t, http.StatusBadRequest, response.Code, "HTTP status code should be 400")
	
	// Check response body
	assert.JSONEq(t, errorMessage(e.ErrInvalidParam), response.Body.String(), "Response should match expected error message")
}

func TestBanIPControllerNotFound(t *testing.T) {
	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(mockAdminMiddleware([]uint{1, 1, 1}))
	router.POST("/banip", BanIPController)

	// Set up SQL mock
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Mock the Status query - not found error
	mock.ExpectQuery(`SELECT post_ip FROM threads
	    INNER JOIN posts ON threads.thread_id = posts.thread_id
	    WHERE ib_id = \? AND threads.thread_id = \? AND post_num = \? LIMIT 1`).
		WithArgs(1, 1, 1).
		WillReturnError(sql.ErrNoRows)

	// Create JSON request
	jsonRequest := []byte(`{"reason":"test reason"}`)
	
	// Perform the request
	response := performJSONRequest(router, "POST", "/banip", jsonRequest)

	// Check response code
	assert.Equal(t, http.StatusNotFound, response.Code, "HTTP status code should be 404")
	
	// Check response body
	assert.JSONEq(t, errorMessage(e.ErrNotFound), response.Body.String(), "Response should match expected error message")

	// Make sure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestBanIPControllerStatusError(t *testing.T) {
	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(mockAdminMiddleware([]uint{1, 1, 1}))
	router.POST("/banip", BanIPController)

	// Set up SQL mock
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Mock the Status query - database error
	mock.ExpectQuery(`SELECT post_ip FROM threads
	    INNER JOIN posts ON threads.thread_id = posts.thread_id
	    WHERE ib_id = \? AND threads.thread_id = \? AND post_num = \? LIMIT 1`).
		WithArgs(1, 1, 1).
		WillReturnError(fmt.Errorf("database error"))

	// Create JSON request
	jsonRequest := []byte(`{"reason":"test reason"}`)
	
	// Perform the request
	response := performJSONRequest(router, "POST", "/banip", jsonRequest)

	// Check response code
	assert.Equal(t, http.StatusInternalServerError, response.Code, "HTTP status code should be 500")
	
	// Check response body
	assert.JSONEq(t, errorMessage(e.ErrInternalError), response.Body.String(), "Response should match expected error message")

	// Make sure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestBanIPControllerPostError(t *testing.T) {
	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(mockAdminMiddleware([]uint{1, 1, 1}))
	router.POST("/banip", BanIPController)

	// Set up SQL mock
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Mock the Status query - successful IP lookup
	mock.ExpectQuery(`SELECT post_ip FROM threads
	    INNER JOIN posts ON threads.thread_id = posts.thread_id
	    WHERE ib_id = \? AND threads.thread_id = \? AND post_num = \? LIMIT 1`).
		WithArgs(1, 1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"post_ip"}).
			AddRow("10.0.0.1"))

	// Mock the insert query - database error
	mock.ExpectExec("INSERT IGNORE INTO banned_ips \\(user_id,ib_id,ban_ip,ban_reason\\) VALUES \\(\\?,\\?,\\?,\\?\\)").
		WithArgs(2, 1, "10.0.0.1", "test reason").
		WillReturnError(errors.New("database error"))

	// Create JSON request
	jsonRequest := []byte(`{"reason":"test reason"}`)
	
	// Perform the request
	response := performJSONRequest(router, "POST", "/banip", jsonRequest)

	// Check response code
	assert.Equal(t, http.StatusInternalServerError, response.Code, "HTTP status code should be 500")
	
	// Check response body
	assert.JSONEq(t, errorMessage(e.ErrInternalError), response.Body.String(), "Response should match expected error message")

	// Make sure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}