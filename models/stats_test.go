package models

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"github.com/eirka/eirka-libs/db"
)

func TestStatisticsModelGet(t *testing.T) {
	var err error

	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Initialize model with parameters
	m := &StatisticsModel{
		Ib: 1,
	}

	// Mock the board stats query
	mock.ExpectQuery(`SELECT \(SELECT COUNT\(thread_id\).*FROM imageboards WHERE ib_id = \?`).
		WithArgs(m.Ib).
		WillReturnRows(sqlmock.NewRows([]string{"thread_count", "post_count", "image_count"}).
			AddRow(10, 100, 50))

	// Mock the visitor stats query
	mock.ExpectQuery(`SELECT COUNT\(DISTINCT request_ip\) as visitors, COUNT\(request_itemkey\) as hits.*WHERE request_time BETWEEN \(CURDATE\(\) - INTERVAL 1 DAY\) AND now\(\) AND ib_id = \?`).
		WithArgs(m.Ib).
		WillReturnRows(sqlmock.NewRows([]string{"visitors", "hits"}).
			AddRow(200, 500))

	// Mock the chart data prepare
	mock.ExpectPrepare(`SELECT \(now\(\) - interval \? hour\) as time.*AND ib_id = \?`)

	// Mock period data (6 periods at every 4 hours)
	timestamp := time.Now()
	for hour := 24; hour >= 4; hour -= 4 {
		mock.ExpectQuery(`SELECT \(now\(\) - interval \? hour\) as time.*AND ib_id = \?`).
			WithArgs(hour, hour, hour-4, m.Ib).
			WillReturnRows(sqlmock.NewRows([]string{"time", "visitors", "hits"}).
				AddRow(timestamp.Add(-time.Hour*time.Duration(hour)), 20, 50))
	}

	// Get the statistics
	err = m.Get()
	assert.NoError(t, err, "No error should be returned")

	// Verify result data
	assert.Equal(t, uint(10), m.Result.Threads, "Threads count should match")
	assert.Equal(t, uint(100), m.Result.Posts, "Posts count should match")
	assert.Equal(t, uint(50), m.Result.Images, "Images count should match")
	assert.Equal(t, uint(200), m.Result.Visitors, "Visitors count should match")
	assert.Equal(t, uint(500), m.Result.Hits, "Hits count should match")
	assert.Equal(t, 6, len(m.Result.Labels), "Should have 6 data points")
	assert.Equal(t, 2, len(m.Result.Series), "Should have 2 series")
	assert.Equal(t, "Visitors", m.Result.Series[0].Name, "First series should be visitors")
	assert.Equal(t, "Hits", m.Result.Series[1].Name, "Second series should be hits")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestStatisticsModelGetDbError(t *testing.T) {
	// Close the database connection to force a GetDb error
	db.CloseDb()

	// Initialize model with parameters
	m := &StatisticsModel{
		Ib: 1,
	}

	// Get the statistics - should encounter GetDb error
	err := m.Get()
	assert.Error(t, err, "Error should be returned")
	assert.Contains(t, err.Error(), "database is closed", "Error should indicate database is closed")
}

func TestStatisticsModelBoardStatsError(t *testing.T) {
	var err error

	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Initialize model with parameters
	m := &StatisticsModel{
		Ib: 1,
	}

	// Mock the board stats query with error
	mock.ExpectQuery(`SELECT \(SELECT COUNT\(thread_id\).*FROM imageboards WHERE ib_id = \?`).
		WithArgs(m.Ib).
		WillReturnError(errors.New("database error"))

	// Get the statistics
	err = m.Get()
	assert.Error(t, err, "Error should be returned")
	assert.Contains(t, err.Error(), "database error", "Error should contain the expected error message")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestStatisticsModelVisitorStatsError(t *testing.T) {
	var err error

	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Initialize model with parameters
	m := &StatisticsModel{
		Ib: 1,
	}

	// Mock the board stats query
	mock.ExpectQuery(`SELECT \(SELECT COUNT\(thread_id\).*FROM imageboards WHERE ib_id = \?`).
		WithArgs(m.Ib).
		WillReturnRows(sqlmock.NewRows([]string{"thread_count", "post_count", "image_count"}).
			AddRow(10, 100, 50))

	// Mock the visitor stats query with error
	mock.ExpectQuery(`SELECT COUNT\(DISTINCT request_ip\) as visitors, COUNT\(request_itemkey\) as hits.*WHERE request_time BETWEEN \(CURDATE\(\) - INTERVAL 1 DAY\) AND now\(\) AND ib_id = \?`).
		WithArgs(m.Ib).
		WillReturnError(errors.New("database error"))

	// Get the statistics
	err = m.Get()
	assert.Error(t, err, "Error should be returned")
	assert.Contains(t, err.Error(), "database error", "Error should contain the expected error message")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestStatisticsModelPrepareError(t *testing.T) {
	var err error

	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Initialize model with parameters
	m := &StatisticsModel{
		Ib: 1,
	}

	// Mock the board stats query
	mock.ExpectQuery(`SELECT \(SELECT COUNT\(thread_id\).*FROM imageboards WHERE ib_id = \?`).
		WithArgs(m.Ib).
		WillReturnRows(sqlmock.NewRows([]string{"thread_count", "post_count", "image_count"}).
			AddRow(10, 100, 50))

	// Mock the visitor stats query
	mock.ExpectQuery(`SELECT COUNT\(DISTINCT request_ip\) as visitors, COUNT\(request_itemkey\) as hits.*WHERE request_time BETWEEN \(CURDATE\(\) - INTERVAL 1 DAY\) AND now\(\) AND ib_id = \?`).
		WithArgs(m.Ib).
		WillReturnRows(sqlmock.NewRows([]string{"visitors", "hits"}).
			AddRow(200, 500))

	// Mock the chart data prepare with error
	mock.ExpectPrepare(`SELECT \(now\(\) - interval \? hour\) as time.*AND ib_id = \?`).
		WillReturnError(errors.New("prepare error"))

	// Get the statistics
	err = m.Get()
	assert.Error(t, err, "Error should be returned")
	assert.Contains(t, err.Error(), "prepare error", "Error should contain the expected error message")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestStatisticsModelPeriodDataError(t *testing.T) {
	var err error

	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Initialize model with parameters
	m := &StatisticsModel{
		Ib: 1,
	}

	// Mock the board stats query
	mock.ExpectQuery(`SELECT \(SELECT COUNT\(thread_id\).*FROM imageboards WHERE ib_id = \?`).
		WithArgs(m.Ib).
		WillReturnRows(sqlmock.NewRows([]string{"thread_count", "post_count", "image_count"}).
			AddRow(10, 100, 50))

	// Mock the visitor stats query
	mock.ExpectQuery(`SELECT COUNT\(DISTINCT request_ip\) as visitors, COUNT\(request_itemkey\) as hits.*WHERE request_time BETWEEN \(CURDATE\(\) - INTERVAL 1 DAY\) AND now\(\) AND ib_id = \?`).
		WithArgs(m.Ib).
		WillReturnRows(sqlmock.NewRows([]string{"visitors", "hits"}).
			AddRow(200, 500))

	// Mock the chart data prepare
	mock.ExpectPrepare(`SELECT \(now\(\) - interval \? hour\) as time.*AND ib_id = \?`)

	// Mock first period data with error
	mock.ExpectQuery(`SELECT \(now\(\) - interval \? hour\) as time.*AND ib_id = \?`).
		WithArgs(24, 24, 20, m.Ib).
		WillReturnError(errors.New("query error"))

	// Get the statistics
	err = m.Get()
	assert.Error(t, err, "Error should be returned")
	assert.Contains(t, err.Error(), "query error", "Error should contain the expected error message")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}
