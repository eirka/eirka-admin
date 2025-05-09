package controllers

import (
	"database/sql"
	"fmt"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"github.com/eirka/eirka-libs/audit"
	"github.com/eirka/eirka-libs/db"
	e "github.com/eirka/eirka-libs/errors"
	"github.com/eirka/eirka-libs/redis"
)

func TestBanFileController(t *testing.T) {
	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(mockAdminMiddleware([]uint{1, 1, 1}))
	router.POST("/banfile", BanFileController)

	// Set up fake Redis connection
	redis.NewRedisMock()

	// Set up SQL mock
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Mock the Status query - image hash lookup
	mock.ExpectQuery(`SELECT image_hash FROM threads
	    INNER JOIN posts ON threads.thread_id = posts.thread_id
	    INNER JOIN images ON posts.post_id = images.post_id
	    WHERE ib_id = \? AND threads.thread_id = \? AND post_num = \? LIMIT 1`).
		WithArgs(1, 1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"image_hash"}).
			AddRow("abcdef1234567890"))

	// Mock the Post query - insert into banned_files
	mock.ExpectExec("INSERT IGNORE INTO banned_files \\(user_id,ib_id,ban_hash,ban_reason\\) VALUES \\(\\?,\\?,\\?,\\?\\)").
		WithArgs(2, 1, "abcdef1234567890", "test reason").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Create JSON request
	jsonRequest := []byte(`{"reason":"test reason"}`)
	
	// Perform the request
	response := performJSONRequest(router, "POST", "/banfile", jsonRequest)

	// Check response code
	assert.Equal(t, http.StatusOK, response.Code, "HTTP status code should be 200")
	
	// Check response body
	assert.JSONEq(t, successMessage(audit.AuditBanFile), response.Body.String(), "Response should match expected success message")

	// Make sure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestBanFileControllerNotProtected(t *testing.T) {
	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(mockNonAdminMiddleware([]uint{1, 1, 1}))
	router.POST("/banfile", BanFileController)

	// Create JSON request
	jsonRequest := []byte(`{"reason":"test reason"}`)
	
	// Perform the request
	response := performJSONRequest(router, "POST", "/banfile", jsonRequest)

	// Check response code
	assert.Equal(t, http.StatusInternalServerError, response.Code, "HTTP status code should be 500")
	
	// Check response body
	assert.JSONEq(t, errorMessage(e.ErrInternalError), response.Body.String(), "Response should match expected error message")
}

func TestBanFileControllerInvalidParam(t *testing.T) {
	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(mockAdminMiddleware([]uint{1, 1, 1}))
	router.POST("/banfile", BanFileController)

	// Create invalid JSON request (missing required "reason" field)
	jsonRequest := []byte(`{}`)
	
	// Perform the request
	response := performJSONRequest(router, "POST", "/banfile", jsonRequest)

	// Check response code
	assert.Equal(t, http.StatusBadRequest, response.Code, "HTTP status code should be 400")
	
	// Check response body
	assert.JSONEq(t, errorMessage(e.ErrInvalidParam), response.Body.String(), "Response should match expected error message")
}

func TestBanFileControllerNotFound(t *testing.T) {
	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(mockAdminMiddleware([]uint{1, 1, 1}))
	router.POST("/banfile", BanFileController)

	// Set up SQL mock
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Mock the Status query - not found error
	mock.ExpectQuery(`SELECT image_hash FROM threads
	    INNER JOIN posts ON threads.thread_id = posts.thread_id
	    INNER JOIN images ON posts.post_id = images.post_id
	    WHERE ib_id = \? AND threads.thread_id = \? AND post_num = \? LIMIT 1`).
		WithArgs(1, 1, 1).
		WillReturnError(sql.ErrNoRows)

	// Create JSON request
	jsonRequest := []byte(`{"reason":"test reason"}`)
	
	// Perform the request
	response := performJSONRequest(router, "POST", "/banfile", jsonRequest)

	// Check response code
	assert.Equal(t, http.StatusNotFound, response.Code, "HTTP status code should be 404")
	
	// Check response body
	assert.JSONEq(t, errorMessage(e.ErrNotFound), response.Body.String(), "Response should match expected error message")

	// Make sure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestBanFileControllerStatusError(t *testing.T) {
	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(mockAdminMiddleware([]uint{1, 1, 1}))
	router.POST("/banfile", BanFileController)

	// Set up SQL mock
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Mock the Status query - database error
	mock.ExpectQuery(`SELECT image_hash FROM threads
	    INNER JOIN posts ON threads.thread_id = posts.thread_id
	    INNER JOIN images ON posts.post_id = images.post_id
	    WHERE ib_id = \? AND threads.thread_id = \? AND post_num = \? LIMIT 1`).
		WithArgs(1, 1, 1).
		WillReturnError(fmt.Errorf("database error"))

	// Create JSON request
	jsonRequest := []byte(`{"reason":"test reason"}`)
	
	// Perform the request
	response := performJSONRequest(router, "POST", "/banfile", jsonRequest)

	// Check response code
	assert.Equal(t, http.StatusInternalServerError, response.Code, "HTTP status code should be 500")
	
	// Check response body
	assert.JSONEq(t, errorMessage(e.ErrInternalError), response.Body.String(), "Response should match expected error message")

	// Make sure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestBanFileControllerPostError(t *testing.T) {
	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(mockAdminMiddleware([]uint{1, 1, 1}))
	router.POST("/banfile", BanFileController)

	// Set up SQL mock
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Mock the Status query - successful lookup
	mock.ExpectQuery(`SELECT image_hash FROM threads
	    INNER JOIN posts ON threads.thread_id = posts.thread_id
	    INNER JOIN images ON posts.post_id = images.post_id
	    WHERE ib_id = \? AND threads.thread_id = \? AND post_num = \? LIMIT 1`).
		WithArgs(1, 1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"image_hash"}).
			AddRow("abcdef1234567890"))

	// Mock the Post query - database error
	mock.ExpectExec("INSERT IGNORE INTO banned_files \\(user_id,ib_id,ban_hash,ban_reason\\) VALUES \\(\\?,\\?,\\?,\\?\\)").
		WithArgs(2, 1, "abcdef1234567890", "test reason").
		WillReturnError(fmt.Errorf("database error"))

	// Create JSON request
	jsonRequest := []byte(`{"reason":"test reason"}`)
	
	// Perform the request
	response := performJSONRequest(router, "POST", "/banfile", jsonRequest)

	// Check response code
	assert.Equal(t, http.StatusInternalServerError, response.Code, "HTTP status code should be 500")
	
	// Check response body
	assert.JSONEq(t, errorMessage(e.ErrInternalError), response.Body.String(), "Response should match expected error message")

	// Make sure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}