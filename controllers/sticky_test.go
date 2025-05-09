package controllers

import (
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

func TestStickyThreadController(t *testing.T) {
	var err error

	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Use our mock middleware 
	router.Use(mockAdminMiddleware([]uint{1, 1}))

	// Set up routes
	router.POST("/sticky/:ib/:id", StickyThreadController)

	// Set up fake Redis connection
	redis.NewRedisMock()

	// Set up SQL mock
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Mock the Status query - thread exists and is currently not sticky
	mock.ExpectQuery(`SELECT thread_title, thread_sticky FROM threads WHERE thread_id = \? AND ib_id = \? LIMIT 1`).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"thread_title", "thread_sticky"}).
			AddRow("Test Thread", 0))

	// Mock the Toggle query - toggle the sticky state
	mock.ExpectPrepare("UPDATE threads SET thread_sticky").
		ExpectExec().
		WithArgs(true, 1, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Mock Redis cache deletion
	redis.Cache.Mock.Command("DEL", "index:1", "directory:1", "thread:1:1")

	// Perform the request
	first := performRequest(router, "POST", "/sticky/1/1")

	// Check assertions
	assert.Equal(t, http.StatusOK, first.Code, "HTTP request code should match")
	assert.JSONEq(t, successMessage(audit.AuditStickyThread), first.Body.String(), "HTTP response should match")

	// Verify all SQL expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "An error was not expected")
}

func TestStickyThreadController_UnstickThread(t *testing.T) {
	var err error

	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Use our mock middleware
	router.Use(mockAdminMiddleware([]uint{1, 1}))

	// Set up routes
	router.POST("/sticky/:ib/:id", StickyThreadController)

	// Set up fake Redis connection
	redis.NewRedisMock()

	// Set up SQL mock
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Mock the Status query - thread exists and is currently sticky
	mock.ExpectQuery(`SELECT thread_title, thread_sticky FROM threads WHERE thread_id = \? AND ib_id = \? LIMIT 1`).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"thread_title", "thread_sticky"}).
			AddRow("Test Thread", 1)) // Thread is already sticky (1)

	// Mock the Toggle query - toggle the sticky state (to false now)
	mock.ExpectPrepare("UPDATE threads SET thread_sticky").
		ExpectExec().
		WithArgs(false, 1, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Mock Redis cache deletion
	redis.Cache.Mock.Command("DEL", "index:1", "directory:1", "thread:1:1")

	// Perform the request
	first := performRequest(router, "POST", "/sticky/1/1")

	// Check assertions - the success message should now be AuditUnstickyThread
	assert.Equal(t, http.StatusOK, first.Code, "HTTP request code should match")
	assert.JSONEq(t, successMessage(audit.AuditUnstickyThread), first.Body.String(), "HTTP response should match")

	// Verify all SQL expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "An error was not expected")
}

func TestStickyThreadController_NotProtected(t *testing.T) {
	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Use non-protected middleware
	router.Use(mockNonAdminMiddleware([]uint{1, 1}))
	
	// Set up routes
	router.POST("/sticky/:ib/:id", StickyThreadController)

	// Perform the request
	first := performRequest(router, "POST", "/sticky/1/1")

	// Check assertions
	assert.Equal(t, http.StatusInternalServerError, first.Code, "HTTP request code should match")
	assert.JSONEq(t, errorMessage(e.ErrInternalError), first.Body.String(), "HTTP response should match")
}

func TestStickyThreadController_DatabaseConnectionError(t *testing.T) {
	// This test is difficult to implement without access to the database package's internals
	// We'll simulate a database error by using the error path from a different test
	
	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Use our mock middleware
	router.Use(mockAdminMiddleware([]uint{1, 1}))

	// Set up routes
	router.POST("/sticky/:ib/:id", StickyThreadController)

	// Set up fake Redis connection
	redis.NewRedisMock()

	// Set up SQL mock
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Mock the Status query - will return a database error
	mock.ExpectQuery(`SELECT thread_title, thread_sticky FROM threads WHERE thread_id = \? AND ib_id = \? LIMIT 1`).
		WithArgs(1, 1).
		WillReturnError(fmt.Errorf("database connection error"))

	// Perform the request
	first := performRequest(router, "POST", "/sticky/1/1")

	// Check assertions
	assert.Equal(t, http.StatusInternalServerError, first.Code, "HTTP request code should match")
	assert.JSONEq(t, errorMessage(e.ErrInternalError), first.Body.String(), "HTTP response should match")
	
	// Verify all SQL expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "An error was not expected")
}

func TestStickyThreadController_RedisError(t *testing.T) {
	var err error

	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	
	// Use our mock middleware
	router.Use(mockAdminMiddleware([]uint{1, 1}))

	// Set up routes
	router.POST("/sticky/:ib/:id", StickyThreadController)

	// Set up fake Redis connection
	redis.NewRedisMock()

	// Set up SQL mock
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Mock the Status query
	mock.ExpectQuery(`SELECT thread_title, thread_sticky FROM threads WHERE thread_id = \? AND ib_id = \? LIMIT 1`).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"thread_title", "thread_sticky"}).
			AddRow("Test Thread", 0))

	// Mock the Toggle query
	mock.ExpectPrepare("UPDATE threads SET thread_sticky").
		ExpectExec().
		WithArgs(true, 1, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Mock Redis cache deletion - with error
	redis.Cache.Mock.Command("DEL", "index:1", "directory:1", "thread:1:1").ExpectError(fmt.Errorf("redis error"))

	// Perform the request
	first := performRequest(router, "POST", "/sticky/1/1")

	// Check assertions
	assert.Equal(t, http.StatusInternalServerError, first.Code, "HTTP request code should match")
	assert.JSONEq(t, errorMessage(e.ErrInternalError), first.Body.String(), "HTTP response should match")

	// Verify all SQL expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "An error was not expected")
}

// Skipping the audit test for now as it requires patching a method that's not directly accessible
// This would require a refactor of the audit package or a different approach to testing