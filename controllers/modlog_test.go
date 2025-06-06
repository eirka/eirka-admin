package controllers

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"github.com/eirka/eirka-libs/config"
	"github.com/eirka/eirka-libs/db"
	e "github.com/eirka/eirka-libs/errors"
)

func TestModLogController(t *testing.T) {
	var err error

	gin.SetMode(gin.TestMode)

	// Set config settings
	config.Settings.Limits.PostsPerPage = 10

	mock, err := db.NewTestDb()
	assert.NoError(t, err)
	defer db.CloseDb()

	// Initialize router
	router := gin.New()
	router.Use(gin.Recovery())

	// Test route with params
	params := []uint{1, 1} // board id 1, page 1
	router.GET("/modlog", mockAdminMiddleware(params), ModLogController)

	// Total count
	mock.ExpectQuery(`SELECT count\(\*\) FROM audit WHERE ib_id = \? AND audit_type = 2`).
		WithArgs(params[0]).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	// Log rows
	now := time.Now()
	logRows := sqlmock.NewRows([]string{
		"user_id", "user_name", "role", "audit_time", "audit_action", "audit_info",
	}).
		AddRow(2, "test", 3, now, "deleted thread", "Thread 1").
		AddRow(1, "admin", 4, now, "banned user", "User 2")

	mock.ExpectQuery(`SELECT audit.user_id,user_name,(.+) ORDER BY audit_id DESC LIMIT \?,\?`).
		WithArgs(params[0], params[0], 0, config.Settings.Limits.PostsPerPage).
		WillReturnRows(logRows)

	// Make request
	w := performRequest(router, "GET", "/modlog")

	// Check response
	assert.Equal(t, 200, w.Code)

	// Parse response JSON
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Check log structure
	modlog, ok := response["modlog"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, float64(1), modlog["current_page"])
	assert.Equal(t, float64(5), modlog["total"])
	assert.Equal(t, float64(1), modlog["pages"])

	// Check log items
	items, ok := modlog["items"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 2, len(items))

	// Check first item
	firstItem := items[0].(map[string]interface{})
	assert.Equal(t, float64(2), firstItem["user_id"])
	assert.Equal(t, "test", firstItem["user_name"])
	assert.Equal(t, float64(3), firstItem["user_group"])
	assert.Equal(t, "deleted thread", firstItem["log_action"])
	assert.Equal(t, "Thread 1", firstItem["log_meta"])

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestModLogControllerNotFound(t *testing.T) {
	var err error

	gin.SetMode(gin.TestMode)

	// Set config settings
	config.Settings.Limits.PostsPerPage = 10

	mock, err := db.NewTestDb()
	assert.NoError(t, err)
	defer db.CloseDb()

	// Initialize router
	router := gin.New()
	router.Use(gin.Recovery())

	// Test route with params for a non-existent page
	params := []uint{1, 2} // board id 1, page 2 (doesn't exist)
	router.GET("/modlog", mockAdminMiddleware(params), ModLogController)

	// Total count - only 5 items, so page 2 is out of range
	mock.ExpectQuery(`SELECT count\(\*\) FROM audit WHERE ib_id = \? AND audit_type = 2`).
		WithArgs(params[0]).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	// Make request
	w := performRequest(router, "GET", "/modlog")

	// Check response
	assert.Equal(t, 404, w.Code)
	assert.Equal(t, errorMessage(e.ErrNotFound), w.Body.String())

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestModLogControllerDbError(t *testing.T) {
	var err error

	gin.SetMode(gin.TestMode)

	mock, err := db.NewTestDb()
	assert.NoError(t, err)
	defer db.CloseDb()

	// Initialize router
	router := gin.New()
	router.Use(gin.Recovery())

	// Test route with params
	params := []uint{1, 1} // board id 1, page 1
	router.GET("/modlog", mockAdminMiddleware(params), ModLogController)

	// Mock a database error
	expectedError := errors.New("database error")
	mock.ExpectQuery(`SELECT count\(\*\) FROM audit WHERE ib_id = \? AND audit_type = 2`).
		WithArgs(params[0]).
		WillReturnError(expectedError)

	// Make request
	w := performRequest(router, "GET", "/modlog")

	// Check response
	assert.Equal(t, 500, w.Code)
	assert.Equal(t, errorMessage(e.ErrInternalError), w.Body.String())

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestModLogControllerNotProtected(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Initialize router
	router := gin.New()
	router.Use(gin.Recovery())

	// Test route with non-admin middleware
	params := []uint{1, 1} // board id 1, page 1
	router.GET("/modlog", mockNonAdminMiddleware(params), ModLogController)

	// Make request
	w := performRequest(router, "GET", "/modlog")

	// Check response
	assert.Equal(t, 500, w.Code)
	assert.Equal(t, errorMessage(e.ErrInternalError), w.Body.String())
}
