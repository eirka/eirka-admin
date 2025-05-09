package controllers

import (
	"errors"
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

func TestCloseThreadController(t *testing.T) {
	var err error

	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Set up fake Redis connection
	redis.NewRedisMock()

	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Use our mock middleware
	params := []uint{1, 1} // board id 1, thread id 1
	router.GET("/close", mockAdminMiddleware(params), CloseThreadController)

	// Thread status query successful - set closed to false initially
	mock.ExpectQuery(`SELECT thread_title, thread_closed FROM threads WHERE thread_id = \? AND ib_id = \? LIMIT 1`).
		WithArgs(params[1], params[0]).
		WillReturnRows(sqlmock.NewRows([]string{"thread_title", "thread_closed"}).
			AddRow("Thread 1", false))

	// Toggle prepare and exec - set to closed (!false = true)
	mock.ExpectPrepare(`UPDATE threads SET thread_closed = \? WHERE thread_id = \? AND ib_id = \?`).
		ExpectExec().
		WithArgs(true, params[1], params[0]).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Mock Redis cache deletion
	redis.Cache.Mock.Command("DEL", "index:1", "directory:1", "thread:1:1")

	// Make request
	w := performRequest(router, "GET", "/close")

	// Check response
	assert.Equal(t, http.StatusOK, w.Code, "HTTP status code should be 200")
	assert.JSONEq(t, successMessage(audit.AuditCloseThread), w.Body.String(), "Response should match expected success message")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "An error was not expected")
}

func TestCloseThreadControllerOpening(t *testing.T) {
	var err error

	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Set up fake Redis connection
	redis.NewRedisMock()

	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Use our mock middleware
	params := []uint{1, 1} // board id 1, thread id 1
	router.GET("/close", mockAdminMiddleware(params), CloseThreadController)

	// Thread status query successful - set closed to true initially
	mock.ExpectQuery(`SELECT thread_title, thread_closed FROM threads WHERE thread_id = \? AND ib_id = \? LIMIT 1`).
		WithArgs(params[1], params[0]).
		WillReturnRows(sqlmock.NewRows([]string{"thread_title", "thread_closed"}).
			AddRow("Thread 1", true))

	// Toggle prepare and exec - set to open (!true = false)
	mock.ExpectPrepare(`UPDATE threads SET thread_closed = \? WHERE thread_id = \? AND ib_id = \?`).
		ExpectExec().
		WithArgs(false, params[1], params[0]).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Mock Redis cache deletion
	redis.Cache.Mock.Command("DEL", "index:1", "directory:1", "thread:1:1")

	// Make request
	w := performRequest(router, "GET", "/close")

	// Check response
	assert.Equal(t, http.StatusOK, w.Code, "HTTP status code should be 200")
	assert.JSONEq(t, successMessage(audit.AuditOpenThread), w.Body.String(), "Response should match expected success message")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "An error was not expected")
}

func TestCloseThreadControllerNotFound(t *testing.T) {
	var err error

	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Set up fake Redis connection
	redis.NewRedisMock()

	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Use our mock middleware
	params := []uint{1, 1} // board id 1, thread id 1 (doesn't exist)
	router.GET("/close", mockAdminMiddleware(params), CloseThreadController)

	// Thread status query not found
	mock.ExpectQuery(`SELECT thread_title, thread_closed FROM threads WHERE thread_id = \? AND ib_id = \? LIMIT 1`).
		WithArgs(params[1], params[0]).
		WillReturnError(e.ErrNotFound)

	// Make request
	w := performRequest(router, "GET", "/close")

	// Check response
	assert.Equal(t, http.StatusNotFound, w.Code, "HTTP status code should be 404")
	assert.JSONEq(t, errorMessage(e.ErrNotFound), w.Body.String(), "Response should match expected error message")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "An error was not expected")
}

func TestCloseThreadControllerStatusError(t *testing.T) {
	var err error

	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Set up fake Redis connection
	redis.NewRedisMock()

	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Use our mock middleware
	params := []uint{1, 1} // board id 1, thread id 1
	router.GET("/close", mockAdminMiddleware(params), CloseThreadController)

	// Thread status query error
	expectedError := errors.New("database error")
	mock.ExpectQuery(`SELECT thread_title, thread_closed FROM threads WHERE thread_id = \? AND ib_id = \? LIMIT 1`).
		WithArgs(params[1], params[0]).
		WillReturnError(expectedError)

	// Make request
	w := performRequest(router, "GET", "/close")

	// Check response
	assert.Equal(t, http.StatusInternalServerError, w.Code, "HTTP status code should be 500")
	assert.JSONEq(t, errorMessage(e.ErrInternalError), w.Body.String(), "Response should match expected error message")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "An error was not expected")
}

func TestCloseThreadControllerToggleError(t *testing.T) {
	var err error

	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Set up fake Redis connection
	redis.NewRedisMock()

	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Use our mock middleware
	params := []uint{1, 1} // board id 1, thread id 1
	router.GET("/close", mockAdminMiddleware(params), CloseThreadController)

	// Thread status query successful
	mock.ExpectQuery(`SELECT thread_title, thread_closed FROM threads WHERE thread_id = \? AND ib_id = \? LIMIT 1`).
		WithArgs(params[1], params[0]).
		WillReturnRows(sqlmock.NewRows([]string{"thread_title", "thread_closed"}).
			AddRow("Thread 1", false))

	// Toggle error
	expectedError := errors.New("toggle error")
	mock.ExpectPrepare(`UPDATE threads SET thread_closed = \? WHERE thread_id = \? AND ib_id = \?`).
		ExpectExec().
		WithArgs(true, params[1], params[0]).
		WillReturnError(expectedError)

	// Make request
	w := performRequest(router, "GET", "/close")

	// Check response
	assert.Equal(t, http.StatusInternalServerError, w.Code, "HTTP status code should be 500")
	assert.JSONEq(t, errorMessage(e.ErrInternalError), w.Body.String(), "Response should match expected error message")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "An error was not expected")
}

func TestCloseThreadControllerRedisError(t *testing.T) {
	var err error

	gin.SetMode(gin.TestMode)

	// Setup a Redis mock
	redis.NewRedisMock()

	mock, err := db.NewTestDb()
	assert.NoError(t, err)
	defer db.CloseDb()

	// Initialize router
	router := gin.New()
	router.Use(gin.Recovery())

	// Setup routes
	params := []uint{1, 1} // board id 1, thread id 1
	router.GET("/close", mockAdminMiddleware(params), CloseThreadController)

	// Thread status query successful
	mock.ExpectQuery(`SELECT thread_title, thread_closed FROM threads WHERE thread_id = \? AND ib_id = \? LIMIT 1`).
		WithArgs(params[1], params[0]).
		WillReturnRows(sqlmock.NewRows([]string{"thread_title", "thread_closed"}).
			AddRow("Thread 1", false))

	// Toggle prepare and exec
	mock.ExpectPrepare(`UPDATE threads SET thread_closed = \? WHERE thread_id = \? AND ib_id = \?`).
		ExpectExec().
		WithArgs(true, params[1], params[0]).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Expect Redis delete operation to error
	redis.Cache.Mock.Command("DEL", "index:1", "directory:1", "thread:1:1").ExpectError(errors.New("redis error"))

	// Make request
	w := performRequest(router, "GET", "/close")

	// Check response
	assert.Equal(t, http.StatusInternalServerError, w.Code, "HTTP status code should be 500")
	assert.JSONEq(t, errorMessage(e.ErrInternalError), w.Body.String(), "Response should match expected error message")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCloseThreadControllerNotProtected(t *testing.T) {
	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Set up fake Redis connection
	redis.NewRedisMock()

	// Setup routes with non-admin middleware
	params := []uint{1, 1} // board id 1, thread id 1
	router.GET("/close", mockNonAdminMiddleware(params), CloseThreadController)

	// Make request
	w := performRequest(router, "GET", "/close")

	// Check response
	assert.Equal(t, http.StatusInternalServerError, w.Code, "HTTP status code should be 500")
	assert.JSONEq(t, errorMessage(e.ErrInternalError), w.Body.String(), "Response should match expected error message")
}

// TestCloseThreadControllerAuditError tests that the controller continues execution
// even if there's an audit submission error
func TestCloseThreadControllerAuditError(t *testing.T) {
	var err error

	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Set up fake Redis connection
	redis.NewRedisMock()

	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Use our mock middleware
	params := []uint{1, 1} // board id 1, thread id 1
	router.GET("/close", mockAdminMiddleware(params), CloseThreadController)

	// Thread status query successful
	mock.ExpectQuery(`SELECT thread_title, thread_closed FROM threads WHERE thread_id = \? AND ib_id = \? LIMIT 1`).
		WithArgs(params[1], params[0]).
		WillReturnRows(sqlmock.NewRows([]string{"thread_title", "thread_closed"}).
			AddRow("Thread 1", false))

	// Toggle prepare and exec
	mock.ExpectPrepare(`UPDATE threads SET thread_closed = \? WHERE thread_id = \? AND ib_id = \?`).
		ExpectExec().
		WithArgs(true, params[1], params[0]).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Mock Redis cache deletion
	redis.Cache.Mock.Command("DEL", "index:1", "directory:1", "thread:1:1")

	// Make request
	w := performRequest(router, "GET", "/close")

	// Check response - request should still succeed since audit errors don't affect response
	assert.Equal(t, http.StatusOK, w.Code, "HTTP status code should be 200")
	assert.JSONEq(t, successMessage(audit.AuditCloseThread), w.Body.String(), "Response should match expected success message")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "An error was not expected")
}
