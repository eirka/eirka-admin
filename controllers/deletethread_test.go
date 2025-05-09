package controllers

import (
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
	"github.com/eirka/eirka-libs/redis"
)

func TestDeleteThreadController(t *testing.T) {
	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(mockAdminMiddleware([]uint{1, 1}))
	router.DELETE("/deletethread", DeleteThreadController)

	// Set up fake Redis connection
	redis.NewRedisMock()

	// Set up SQL mock
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Mock the Status query
	mock.ExpectQuery(`SELECT thread_title, thread_deleted FROM threads WHERE thread_id = \? AND ib_id = \? LIMIT 1`).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"thread_title", "thread_deleted"}).
			AddRow("Test Thread", false))

	// Mock the Delete query
	mock.ExpectPrepare(`UPDATE threads SET thread_deleted = \? WHERE thread_id = \? AND ib_id = \?`).
		ExpectExec().
		WithArgs(true, 1, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Mock Redis cache deletion
	redis.Cache.Mock.Command("DEL", "index:1", "directory:1", "thread:1:1", "post:1:1", "tags:1", "image:1", "new:1", "popular:1", "favorited:1")

	// Perform the request
	response := performRequest(router, "DELETE", "/deletethread")

	// Check response code
	assert.Equal(t, http.StatusOK, response.Code, "HTTP status code should be 200")
	
	// Check response body
	assert.JSONEq(t, successMessage(audit.AuditDeleteThread), response.Body.String(), "Response should match expected success message")

	// Make sure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestDeleteThreadControllerNotProtected(t *testing.T) {
	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(mockNonAdminMiddleware([]uint{1, 1}))
	router.DELETE("/deletethread", DeleteThreadController)

	// Perform the request
	response := performRequest(router, "DELETE", "/deletethread")

	// Check response code
	assert.Equal(t, http.StatusInternalServerError, response.Code, "HTTP status code should be 500")
	
	// Check response body
	assert.JSONEq(t, errorMessage(e.ErrInternalError), response.Body.String(), "Response should match expected error message")
}

func TestDeleteThreadControllerNotFound(t *testing.T) {
	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(mockAdminMiddleware([]uint{1, 1}))
	router.DELETE("/deletethread", DeleteThreadController)

	// Set up SQL mock
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Mock the Status query - not found error
	mock.ExpectQuery(`SELECT thread_title, thread_deleted FROM threads WHERE thread_id = \? AND ib_id = \? LIMIT 1`).
		WithArgs(1, 1).
		WillReturnError(e.ErrNotFound)

	// Perform the request
	response := performRequest(router, "DELETE", "/deletethread")

	// Check response code
	assert.Equal(t, http.StatusNotFound, response.Code, "HTTP status code should be 404")
	
	// Check response body
	assert.JSONEq(t, errorMessage(e.ErrNotFound), response.Body.String(), "Response should match expected error message")

	// Make sure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestDeleteThreadControllerStatusError(t *testing.T) {
	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(mockAdminMiddleware([]uint{1, 1}))
	router.DELETE("/deletethread", DeleteThreadController)

	// Set up SQL mock
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Mock the Status query - database error
	mock.ExpectQuery(`SELECT thread_title, thread_deleted FROM threads WHERE thread_id = \? AND ib_id = \? LIMIT 1`).
		WithArgs(1, 1).
		WillReturnError(fmt.Errorf("database error"))

	// Perform the request
	response := performRequest(router, "DELETE", "/deletethread")

	// Check response code
	assert.Equal(t, http.StatusInternalServerError, response.Code, "HTTP status code should be 500")
	
	// Check response body
	assert.JSONEq(t, errorMessage(e.ErrInternalError), response.Body.String(), "Response should match expected error message")

	// Make sure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestDeleteThreadControllerDeleteError(t *testing.T) {
	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(mockAdminMiddleware([]uint{1, 1}))
	router.DELETE("/deletethread", DeleteThreadController)

	// Set up SQL mock
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Mock the Status query - successful
	mock.ExpectQuery(`SELECT thread_title, thread_deleted FROM threads WHERE thread_id = \? AND ib_id = \? LIMIT 1`).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"thread_title", "thread_deleted"}).
			AddRow("Test Thread", false))

	// Mock the Delete query with error
	mock.ExpectPrepare(`UPDATE threads SET thread_deleted = \? WHERE thread_id = \? AND ib_id = \?`).
		ExpectExec().
		WithArgs(true, 1, 1).
		WillReturnError(errors.New("database error"))

	// Perform the request
	response := performRequest(router, "DELETE", "/deletethread")

	// Check response code
	assert.Equal(t, http.StatusInternalServerError, response.Code, "HTTP status code should be 500")
	
	// Check response body
	assert.JSONEq(t, errorMessage(e.ErrInternalError), response.Body.String(), "Response should match expected error message")

	// Make sure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestDeleteThreadControllerRedisError(t *testing.T) {
	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(mockAdminMiddleware([]uint{1, 1}))
	router.DELETE("/deletethread", DeleteThreadController)

	// Set up fake Redis connection
	redis.NewRedisMock()

	// Set up SQL mock
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Mock the Status query
	mock.ExpectQuery(`SELECT thread_title, thread_deleted FROM threads WHERE thread_id = \? AND ib_id = \? LIMIT 1`).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"thread_title", "thread_deleted"}).
			AddRow("Test Thread", false))

	// Mock the Delete query - successful
	mock.ExpectPrepare(`UPDATE threads SET thread_deleted = \? WHERE thread_id = \? AND ib_id = \?`).
		ExpectExec().
		WithArgs(true, 1, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Mock Redis cache deletion - with error
	redis.Cache.Mock.Command("DEL", "index:1", "directory:1", "thread:1:1", "post:1:1", "tags:1", "image:1", "new:1", "popular:1", "favorited:1").
		ExpectError(errors.New("redis error"))

	// Perform the request
	response := performRequest(router, "DELETE", "/deletethread")

	// Check response code
	assert.Equal(t, http.StatusInternalServerError, response.Code, "HTTP status code should be 500")
	
	// Check response body
	assert.JSONEq(t, errorMessage(e.ErrInternalError), response.Body.String(), "Response should match expected error message")

	// Make sure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestDeleteThreadControllerAuditError(t *testing.T) {
	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(mockAdminMiddleware([]uint{1, 1}))
	router.DELETE("/deletethread", DeleteThreadController)

	// Set up fake Redis connection
	redis.NewRedisMock()

	// Set up SQL mock
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Mock the Status query
	mock.ExpectQuery(`SELECT thread_title, thread_deleted FROM threads WHERE thread_id = \? AND ib_id = \? LIMIT 1`).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"thread_title", "thread_deleted"}).
			AddRow("Test Thread", false))

	// Mock the Delete query
	mock.ExpectPrepare(`UPDATE threads SET thread_deleted = \? WHERE thread_id = \? AND ib_id = \?`).
		ExpectExec().
		WithArgs(true, 1, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Mock Redis cache deletion
	redis.Cache.Mock.Command("DEL", "index:1", "directory:1", "thread:1:1", "post:1:1", "tags:1", "image:1", "new:1", "popular:1", "favorited:1")

	// Perform the request
	response := performRequest(router, "DELETE", "/deletethread")

	// Check response code
	assert.Equal(t, http.StatusOK, response.Code, "HTTP status code should be 200")
	
	// Check response body - should still show success even if audit logging fails
	assert.JSONEq(t, successMessage(audit.AuditDeleteThread), response.Body.String(), "Response should match expected success message")

	// Make sure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}