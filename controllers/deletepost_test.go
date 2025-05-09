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

func TestDeletePostController(t *testing.T) {
	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(mockAdminMiddleware([]uint{1, 1, 1}))
	router.DELETE("/deletepost", DeletePostController)

	// Set up fake Redis connection
	redis.NewRedisMock()

	// Set up SQL mock
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Mock the Status query
	mock.ExpectQuery(`SELECT thread_title, post_deleted FROM threads
		INNER JOIN posts on threads.thread_id = posts.thread_id
		WHERE threads.thread_id = \? AND ib_id = \? LIMIT 1`).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"thread_title", "post_deleted"}).
			AddRow("Test Thread", false))

	// Mock the Delete query
	mock.ExpectBegin()
	mock.ExpectPrepare(`UPDATE posts SET post_deleted = \?
		WHERE posts.thread_id = \? AND posts.post_num = \? LIMIT 1`).
		ExpectExec().
		WithArgs(true, 1, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	// Mock Redis cache deletion
	redis.Cache.Mock.Command("DEL", "index:1", "directory:1", "thread:1:1", "post:1:1", "tags:1", "image:1", "new:1", "popular:1", "favorited:1")

	// Perform the request
	response := performRequest(router, "DELETE", "/deletepost")

	// Check response code
	assert.Equal(t, http.StatusOK, response.Code, "HTTP status code should be 200")
	
	// Check response body
	assert.JSONEq(t, successMessage(audit.AuditDeletePost), response.Body.String(), "Response should match expected success message")

	// Make sure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestDeletePostControllerNotProtected(t *testing.T) {
	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(mockNonAdminMiddleware([]uint{1, 1, 1}))
	router.DELETE("/deletepost", DeletePostController)

	// Perform the request
	response := performRequest(router, "DELETE", "/deletepost")

	// Check response code
	assert.Equal(t, http.StatusInternalServerError, response.Code, "HTTP status code should be 500")
	
	// Check response body
	assert.JSONEq(t, errorMessage(e.ErrInternalError), response.Body.String(), "Response should match expected error message")
}

func TestDeletePostControllerNotFound(t *testing.T) {
	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(mockAdminMiddleware([]uint{1, 1, 1}))
	router.DELETE("/deletepost", DeletePostController)

	// Set up SQL mock
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Mock the Status query - not found error
	mock.ExpectQuery(`SELECT thread_title, post_deleted FROM threads
		INNER JOIN posts on threads.thread_id = posts.thread_id
		WHERE threads.thread_id = \? AND ib_id = \? LIMIT 1`).
		WithArgs(1, 1).
		WillReturnError(e.ErrNotFound)

	// Perform the request
	response := performRequest(router, "DELETE", "/deletepost")

	// Check response code
	assert.Equal(t, http.StatusNotFound, response.Code, "HTTP status code should be 404")
	
	// Check response body
	assert.JSONEq(t, errorMessage(e.ErrNotFound), response.Body.String(), "Response should match expected error message")

	// Make sure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestDeletePostControllerStatusError(t *testing.T) {
	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(mockAdminMiddleware([]uint{1, 1, 1}))
	router.DELETE("/deletepost", DeletePostController)

	// Set up SQL mock
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Mock the Status query - database error
	mock.ExpectQuery(`SELECT thread_title, post_deleted FROM threads
		INNER JOIN posts on threads.thread_id = posts.thread_id
		WHERE threads.thread_id = \? AND ib_id = \? LIMIT 1`).
		WithArgs(1, 1).
		WillReturnError(fmt.Errorf("database error"))

	// Perform the request
	response := performRequest(router, "DELETE", "/deletepost")

	// Check response code
	assert.Equal(t, http.StatusInternalServerError, response.Code, "HTTP status code should be 500")
	
	// Check response body
	assert.JSONEq(t, errorMessage(e.ErrInternalError), response.Body.String(), "Response should match expected error message")

	// Make sure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestDeletePostControllerDeleteError(t *testing.T) {
	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(mockAdminMiddleware([]uint{1, 1, 1}))
	router.DELETE("/deletepost", DeletePostController)

	// Set up SQL mock
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Mock the Status query - successful
	mock.ExpectQuery(`SELECT thread_title, post_deleted FROM threads
		INNER JOIN posts on threads.thread_id = posts.thread_id
		WHERE threads.thread_id = \? AND ib_id = \? LIMIT 1`).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"thread_title", "post_deleted"}).
			AddRow("Test Thread", false))

	// Mock the Delete query with error
	mock.ExpectBegin()
	mock.ExpectPrepare(`UPDATE posts SET post_deleted = \?
		WHERE posts.thread_id = \? AND posts.post_num = \? LIMIT 1`).
		ExpectExec().
		WithArgs(true, 1, 1).
		WillReturnError(errors.New("database error"))
	mock.ExpectRollback()

	// Perform the request
	response := performRequest(router, "DELETE", "/deletepost")

	// Check response code
	assert.Equal(t, http.StatusInternalServerError, response.Code, "HTTP status code should be 500")
	
	// Check response body
	assert.JSONEq(t, errorMessage(e.ErrInternalError), response.Body.String(), "Response should match expected error message")

	// Make sure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestDeletePostControllerRedisError(t *testing.T) {
	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(mockAdminMiddleware([]uint{1, 1, 1}))
	router.DELETE("/deletepost", DeletePostController)

	// Set up fake Redis connection
	redis.NewRedisMock()

	// Set up SQL mock
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Mock the Status query
	mock.ExpectQuery(`SELECT thread_title, post_deleted FROM threads
		INNER JOIN posts on threads.thread_id = posts.thread_id
		WHERE threads.thread_id = \? AND ib_id = \? LIMIT 1`).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"thread_title", "post_deleted"}).
			AddRow("Test Thread", false))

	// Mock the Delete query - successful
	mock.ExpectBegin()
	mock.ExpectPrepare(`UPDATE posts SET post_deleted = \?
		WHERE posts.thread_id = \? AND posts.post_num = \? LIMIT 1`).
		ExpectExec().
		WithArgs(true, 1, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	// Mock Redis cache deletion - with error
	redis.Cache.Mock.Command("DEL", "index:1", "directory:1", "thread:1:1", "post:1:1", "tags:1", "image:1", "new:1", "popular:1", "favorited:1").
		ExpectError(errors.New("redis error"))

	// Perform the request
	response := performRequest(router, "DELETE", "/deletepost")

	// Check response code
	assert.Equal(t, http.StatusInternalServerError, response.Code, "HTTP status code should be 500")
	
	// Check response body
	assert.JSONEq(t, errorMessage(e.ErrInternalError), response.Body.String(), "Response should match expected error message")

	// Make sure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestDeletePostControllerAuditError(t *testing.T) {
	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(mockAdminMiddleware([]uint{1, 1, 1}))
	router.DELETE("/deletepost", DeletePostController)

	// Set up fake Redis connection
	redis.NewRedisMock()

	// Set up SQL mock
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Mock the Status query
	mock.ExpectQuery(`SELECT thread_title, post_deleted FROM threads
		INNER JOIN posts on threads.thread_id = posts.thread_id
		WHERE threads.thread_id = \? AND ib_id = \? LIMIT 1`).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"thread_title", "post_deleted"}).
			AddRow("Test Thread", false))

	// Mock the Delete query
	mock.ExpectBegin()
	mock.ExpectPrepare(`UPDATE posts SET post_deleted = \?
		WHERE posts.thread_id = \? AND posts.post_num = \? LIMIT 1`).
		ExpectExec().
		WithArgs(true, 1, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	// Mock Redis cache deletion
	redis.Cache.Mock.Command("DEL", "index:1", "directory:1", "thread:1:1", "post:1:1", "tags:1", "image:1", "new:1", "popular:1", "favorited:1")

	// Perform the request
	response := performRequest(router, "DELETE", "/deletepost")

	// Check response code
	assert.Equal(t, http.StatusOK, response.Code, "HTTP status code should be 200")
	
	// Check response body - should still show success even if audit logging fails
	assert.JSONEq(t, successMessage(audit.AuditDeletePost), response.Body.String(), "Response should match expected success message")

	// Make sure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}