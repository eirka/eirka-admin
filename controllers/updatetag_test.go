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

func TestUpdateTagController(t *testing.T) {
	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(mockAdminMiddleware([]uint{1}))
	router.PUT("/updatetag", UpdateTagController)

	// Set up fake Redis connection
	redis.NewRedisMock()

	// Set up SQL mock
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Mock the Status query - check if tag exists
	mock.ExpectQuery(`select count\(\*\) from tags where tag_name = \? AND ib_id = \? AND NOT tag_id = \?`).
		WithArgs("Updated Tag", 1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"count(*)"}).
			AddRow(0)) // No duplicate tag

	// Mock the Update query
	mock.ExpectPrepare("UPDATE tags SET tag_name= \\?, tagtype_id= \\? WHERE tag_id = \\? AND ib_id = \\?").
		ExpectExec().
		WithArgs("Updated Tag", 1, 1, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Mock Redis cache deletion
	redis.Cache.Mock.Command("DEL", "tags:1", "tag:1:1", "image:1")

	// Create JSON request
	jsonRequest := []byte(`{"id": 1, "name": "Updated Tag", "type": 1}`)

	// Perform the request
	response := performJSONRequest(router, "PUT", "/updatetag", jsonRequest)

	// Check response code
	assert.Equal(t, http.StatusOK, response.Code, "HTTP status code should be 200")

	// Check response body
	assert.JSONEq(t, successMessage(audit.AuditUpdateTag), response.Body.String(), "Response should match expected success message")

	// Make sure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestUpdateTagControllerNotProtected(t *testing.T) {
	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(mockNonAdminMiddleware([]uint{1}))
	router.PUT("/updatetag", UpdateTagController)

	// Create JSON request
	jsonRequest := []byte(`{"id": 1, "name": "Updated Tag", "type": 1}`)

	// Perform the request
	response := performJSONRequest(router, "PUT", "/updatetag", jsonRequest)

	// Check response code
	assert.Equal(t, http.StatusInternalServerError, response.Code, "HTTP status code should be 500")

	// Check response body
	assert.JSONEq(t, errorMessage(e.ErrInternalError), response.Body.String(), "Response should match expected error message")
}

func TestUpdateTagControllerInvalidParam(t *testing.T) {
	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(mockAdminMiddleware([]uint{1}))
	router.PUT("/updatetag", UpdateTagController)

	// Create invalid JSON request (missing required fields)
	jsonRequest := []byte(`{}`)

	// Perform the request
	response := performJSONRequest(router, "PUT", "/updatetag", jsonRequest)

	// Check response code
	assert.Equal(t, http.StatusBadRequest, response.Code, "HTTP status code should be 400")

	// Check response body
	assert.JSONEq(t, errorMessage(e.ErrInvalidParam), response.Body.String(), "Response should match expected error message")
}

func TestUpdateTagControllerValidateInputError(t *testing.T) {
	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(mockAdminMiddleware([]uint{0})) // Invalid board ID
	router.PUT("/updatetag", UpdateTagController)

	// Create JSON request
	jsonRequest := []byte(`{"id": 1, "name": "Updated Tag", "type": 1}`)

	// Perform the request
	response := performJSONRequest(router, "PUT", "/updatetag", jsonRequest)

	// Check response code
	assert.Equal(t, http.StatusBadRequest, response.Code, "HTTP status code should be 400")

	// Check response body contains error message
	assert.Contains(t, response.Body.String(), "error_message", "Response should contain error message")
}

func TestUpdateTagControllerDuplicateTagError(t *testing.T) {
	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(mockAdminMiddleware([]uint{1}))
	router.PUT("/updatetag", UpdateTagController)

	// Set up SQL mock
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Mock the Status query - check if tag exists
	mock.ExpectQuery(`select count\(\*\) from tags where tag_name = \? AND ib_id = \? AND NOT tag_id = \?`).
		WithArgs("Updated Tag", 1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"count(*)"}).
			AddRow(1)) // Duplicate tag exists

	// Create JSON request
	jsonRequest := []byte(`{"id": 1, "name": "Updated Tag", "type": 1}`)

	// Perform the request
	response := performJSONRequest(router, "PUT", "/updatetag", jsonRequest)

	// Check response code
	assert.Equal(t, http.StatusBadRequest, response.Code, "HTTP status code should be 400")

	// Check response body contains error message about duplicate tag
	assert.Contains(t, response.Body.String(), "error_message", "Response should contain error message")
	assert.Contains(t, response.Body.String(), "duplicate", "Response should mention duplicate tag")

	// Make sure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestUpdateTagControllerStatusError(t *testing.T) {
	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(mockAdminMiddleware([]uint{1}))
	router.PUT("/updatetag", UpdateTagController)

	// Set up SQL mock
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Mock the Status query with error
	mock.ExpectQuery(`select count\(\*\) from tags where tag_name = \? AND ib_id = \? AND NOT tag_id = \?`).
		WithArgs("Updated Tag", 1, 1).
		WillReturnError(fmt.Errorf("database error"))

	// Create JSON request
	jsonRequest := []byte(`{"id": 1, "name": "Updated Tag", "type": 1}`)

	// Perform the request
	response := performJSONRequest(router, "PUT", "/updatetag", jsonRequest)

	// Check response code
	assert.Equal(t, http.StatusInternalServerError, response.Code, "HTTP status code should be 500")

	// Check response body
	assert.JSONEq(t, errorMessage(e.ErrInternalError), response.Body.String(), "Response should match expected error message")

	// Make sure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestUpdateTagControllerUpdateError(t *testing.T) {
	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(mockAdminMiddleware([]uint{1}))
	router.PUT("/updatetag", UpdateTagController)

	// Set up SQL mock
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Mock the Status query - check if tag exists
	mock.ExpectQuery(`select count\(\*\) from tags where tag_name = \? AND ib_id = \? AND NOT tag_id = \?`).
		WithArgs("Updated Tag", 1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"count(*)"}).
			AddRow(0)) // No duplicate tag

	// Mock the Update query with error
	mock.ExpectPrepare("UPDATE tags SET tag_name= \\?, tagtype_id= \\? WHERE tag_id = \\? AND ib_id = \\?").
		ExpectExec().
		WithArgs("Updated Tag", 1, 1, 1).
		WillReturnError(errors.New("database error"))

	// Create JSON request
	jsonRequest := []byte(`{"id": 1, "name": "Updated Tag", "type": 1}`)

	// Perform the request
	response := performJSONRequest(router, "PUT", "/updatetag", jsonRequest)

	// Check response code
	assert.Equal(t, http.StatusInternalServerError, response.Code, "HTTP status code should be 500")

	// Check response body
	assert.JSONEq(t, errorMessage(e.ErrInternalError), response.Body.String(), "Response should match expected error message")

	// Make sure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestUpdateTagControllerRedisError(t *testing.T) {
	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(mockAdminMiddleware([]uint{1}))
	router.PUT("/updatetag", UpdateTagController)

	// Set up fake Redis connection
	redis.NewRedisMock()

	// Set up SQL mock
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Mock the Status query - check if tag exists
	mock.ExpectQuery(`select count\(\*\) from tags where tag_name = \? AND ib_id = \? AND NOT tag_id = \?`).
		WithArgs("Updated Tag", 1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"count(*)"}).
			AddRow(0)) // No duplicate tag

	// Mock the Update query
	mock.ExpectPrepare("UPDATE tags SET tag_name= \\?, tagtype_id= \\? WHERE tag_id = \\? AND ib_id = \\?").
		ExpectExec().
		WithArgs("Updated Tag", 1, 1, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Mock Redis cache deletion with error
	redis.Cache.Mock.Command("DEL", "tags:1", "tag:1:1", "image:1").
		ExpectError(errors.New("redis error"))

	// Create JSON request
	jsonRequest := []byte(`{"id": 1, "name": "Updated Tag", "type": 1}`)

	// Perform the request
	response := performJSONRequest(router, "PUT", "/updatetag", jsonRequest)

	// Check response code
	assert.Equal(t, http.StatusInternalServerError, response.Code, "HTTP status code should be 500")

	// Check response body
	assert.JSONEq(t, errorMessage(e.ErrInternalError), response.Body.String(), "Response should match expected error message")

	// Make sure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestUpdateTagControllerAuditError(t *testing.T) {
	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(mockAdminMiddleware([]uint{1}))
	router.PUT("/updatetag", UpdateTagController)

	// Set up fake Redis connection
	redis.NewRedisMock()

	// Set up SQL mock
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Mock the Status query - check if tag exists
	mock.ExpectQuery(`select count\(\*\) from tags where tag_name = \? AND ib_id = \? AND NOT tag_id = \?`).
		WithArgs("Updated Tag", 1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"count(*)"}).
			AddRow(0)) // No duplicate tag

	// Mock the Update query
	mock.ExpectPrepare("UPDATE tags SET tag_name= \\?, tagtype_id= \\? WHERE tag_id = \\? AND ib_id = \\?").
		ExpectExec().
		WithArgs("Updated Tag", 1, 1, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Mock Redis cache deletion
	redis.Cache.Mock.Command("DEL", "tags:1", "tag:1:1", "image:1")

	// Create JSON request
	jsonRequest := []byte(`{"id": 1, "name": "Updated Tag", "type": 1}`)

	// Perform the request
	response := performJSONRequest(router, "PUT", "/updatetag", jsonRequest)

	// Check response code
	assert.Equal(t, http.StatusOK, response.Code, "HTTP status code should be 200")

	// Check response body - should still show success even if audit logging fails
	assert.JSONEq(t, successMessage(audit.AuditUpdateTag), response.Body.String(), "Response should match expected success message")

	// Make sure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}
