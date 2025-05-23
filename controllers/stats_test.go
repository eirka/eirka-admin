package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"github.com/eirka/eirka-admin/models"
	"github.com/eirka/eirka-libs/db"
	e "github.com/eirka/eirka-libs/errors"
)

func TestStatisticsController(t *testing.T) {
	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(mockAdminMiddleware([]uint{1}))
	router.GET("/stats", StatisticsController)

	// Set up SQL mock
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Mock the board stats query
	mock.ExpectQuery(`SELECT \(SELECT COUNT\(thread_id\).*FROM imageboards WHERE ib_id = \?`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"thread_count", "post_count", "image_count"}).
			AddRow(10, 100, 50))

	// Mock the visitor stats query
	mock.ExpectQuery(`SELECT COUNT\(DISTINCT request_ip\) as visitors, COUNT\(request_itemkey\) as hits.*WHERE request_time BETWEEN \(CURDATE\(\) - INTERVAL 1 DAY\) AND now\(\) AND ib_id = \?`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"visitors", "hits"}).
			AddRow(200, 500))

	// Mock the chart data prepare
	mock.ExpectPrepare(`SELECT \(now\(\) - interval \? hour\) as time.*AND ib_id = \?`)

	// Mock period data (6 periods at every 4 hours)
	timestamp := time.Now()
	for hour := 24; hour >= 4; hour -= 4 {
		mock.ExpectQuery(`SELECT \(now\(\) - interval \? hour\) as time.*AND ib_id = \?`).
			WithArgs(hour, hour, hour-4, 1).
			WillReturnRows(sqlmock.NewRows([]string{"time", "visitors", "hits"}).
				AddRow(timestamp.Add(-time.Hour*time.Duration(hour)), 20, 50))
	}

	// Perform the request
	response := performRequest(router, "GET", "/stats")

	// Check response code
	assert.Equal(t, http.StatusOK, response.Code, "HTTP status code should be 200")

	// Parse response JSON to verify structure
	var result models.StatisticsType
	err = json.Unmarshal(response.Body.Bytes(), &result)
	assert.NoError(t, err, "Response should be valid JSON")

	// Verify data
	assert.Equal(t, uint(10), result.Threads, "Threads count should match")
	assert.Equal(t, uint(100), result.Posts, "Posts count should match")
	assert.Equal(t, uint(50), result.Images, "Images count should match")
	assert.Equal(t, uint(200), result.Visitors, "Visitors count should match")
	assert.Equal(t, uint(500), result.Hits, "Hits count should match")

	// Check if we have the right number of data points
	assert.Equal(t, 6, len(result.Labels), "Should have 6 data points")
	assert.Equal(t, 2, len(result.Series), "Should have 2 series")
	assert.Equal(t, "Visitors", result.Series[0].Name, "First series should be visitors")
	assert.Equal(t, "Hits", result.Series[1].Name, "Second series should be hits")

	// Make sure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestStatisticsControllerNotProtected(t *testing.T) {
	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(mockNonAdminMiddleware([]uint{1}))
	router.GET("/stats", StatisticsController)

	// Perform the request
	response := performRequest(router, "GET", "/stats")

	// Check response code
	assert.Equal(t, http.StatusInternalServerError, response.Code, "HTTP status code should be 500")

	// Check response body
	assert.JSONEq(t, errorMessage(e.ErrInternalError), response.Body.String(), "Response should match expected error message")
}

func TestStatisticsControllerGetError(t *testing.T) {
	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(mockAdminMiddleware([]uint{1}))
	router.GET("/stats", StatisticsController)

	// Set up SQL mock
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Mock the board stats query with error
	mock.ExpectQuery(`SELECT \(SELECT COUNT\(thread_id\).*FROM imageboards WHERE ib_id = \?`).
		WithArgs(1).
		WillReturnError(errors.New("database error"))

	// Perform the request
	response := performRequest(router, "GET", "/stats")

	// Check response code
	assert.Equal(t, http.StatusInternalServerError, response.Code, "HTTP status code should be 500")

	// Check response body
	assert.JSONEq(t, errorMessage(e.ErrInternalError), response.Body.String(), "Response should match expected error message")

	// Make sure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestStatisticsControllerVisitorStatsError(t *testing.T) {
	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(mockAdminMiddleware([]uint{1}))
	router.GET("/stats", StatisticsController)

	// Set up SQL mock
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Mock the board stats query
	mock.ExpectQuery(`SELECT \(SELECT COUNT\(thread_id\).*FROM imageboards WHERE ib_id = \?`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"thread_count", "post_count", "image_count"}).
			AddRow(10, 100, 50))

	// Mock the visitor stats query with error
	mock.ExpectQuery(`SELECT COUNT\(DISTINCT request_ip\) as visitors, COUNT\(request_itemkey\) as hits.*WHERE request_time BETWEEN \(CURDATE\(\) - INTERVAL 1 DAY\) AND now\(\) AND ib_id = \?`).
		WithArgs(1).
		WillReturnError(errors.New("database error"))

	// Perform the request
	response := performRequest(router, "GET", "/stats")

	// Check response code
	assert.Equal(t, http.StatusInternalServerError, response.Code, "HTTP status code should be 500")

	// Check response body
	assert.JSONEq(t, errorMessage(e.ErrInternalError), response.Body.String(), "Response should match expected error message")

	// Make sure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestStatisticsControllerPrepareError(t *testing.T) {
	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(mockAdminMiddleware([]uint{1}))
	router.GET("/stats", StatisticsController)

	// Set up SQL mock
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Mock the board stats query
	mock.ExpectQuery(`SELECT \(SELECT COUNT\(thread_id\).*FROM imageboards WHERE ib_id = \?`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"thread_count", "post_count", "image_count"}).
			AddRow(10, 100, 50))

	// Mock the visitor stats query
	mock.ExpectQuery(`SELECT COUNT\(DISTINCT request_ip\) as visitors, COUNT\(request_itemkey\) as hits.*WHERE request_time BETWEEN \(CURDATE\(\) - INTERVAL 1 DAY\) AND now\(\) AND ib_id = \?`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"visitors", "hits"}).
			AddRow(200, 500))

	// Mock the chart data prepare with error
	mock.ExpectPrepare(`SELECT \(now\(\) - interval \? hour\) as time.*AND ib_id = \?`).
		WillReturnError(errors.New("prepare error"))

	// Perform the request
	response := performRequest(router, "GET", "/stats")

	// Check response code
	assert.Equal(t, http.StatusInternalServerError, response.Code, "HTTP status code should be 500")

	// Check response body
	assert.JSONEq(t, errorMessage(e.ErrInternalError), response.Body.String(), "Response should match expected error message")

	// Make sure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestStatisticsControllerPeriodDataError(t *testing.T) {
	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(mockAdminMiddleware([]uint{1}))
	router.GET("/stats", StatisticsController)

	// Set up SQL mock
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Mock the board stats query
	mock.ExpectQuery(`SELECT \(SELECT COUNT\(thread_id\).*FROM imageboards WHERE ib_id = \?`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"thread_count", "post_count", "image_count"}).
			AddRow(10, 100, 50))

	// Mock the visitor stats query
	mock.ExpectQuery(`SELECT COUNT\(DISTINCT request_ip\) as visitors, COUNT\(request_itemkey\) as hits.*WHERE request_time BETWEEN \(CURDATE\(\) - INTERVAL 1 DAY\) AND now\(\) AND ib_id = \?`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"visitors", "hits"}).
			AddRow(200, 500))

	// Mock the chart data prepare
	mock.ExpectPrepare(`SELECT \(now\(\) - interval \? hour\) as time.*AND ib_id = \?`)

	// Mock first period data with error
	mock.ExpectQuery(`SELECT \(now\(\) - interval \? hour\) as time.*AND ib_id = \?`).
		WithArgs(24, 24, 20, 1).
		WillReturnError(errors.New("query error"))

	// Perform the request
	response := performRequest(router, "GET", "/stats")

	// Check response code
	assert.Equal(t, http.StatusInternalServerError, response.Code, "HTTP status code should be 500")

	// Check response body
	assert.JSONEq(t, errorMessage(e.ErrInternalError), response.Body.String(), "Response should match expected error message")

	// Make sure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestStatisticsControllerJsonError(t *testing.T) {
	// This test is to verify handling of JSON marshaling errors
	// Since it's difficult to force a JSON marshal error in a way that's testable,
	// we'll focus on other test cases that provide better value and coverage.
	t.Skip("Skipping JSON marshal error test - difficult to reproduce in a test environment")
}
