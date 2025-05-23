package models

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"github.com/eirka/eirka-libs/config"
	"github.com/eirka/eirka-libs/db"
	e "github.com/eirka/eirka-libs/errors"
)

func TestBoardLogIsValid(t *testing.T) {

	// Test cases for validation
	tests := []struct {
		name      string
		model     *BoardLogModel
		expectErr error
	}{
		{
			name: "valid",
			model: &BoardLogModel{
				Ib:   1,
				Page: 1,
			},
			expectErr: nil,
		},
		{
			name: "missing ib",
			model: &BoardLogModel{
				Ib:   0,
				Page: 1,
			},
			expectErr: e.ErrNotFound,
		},
		{
			name: "missing page",
			model: &BoardLogModel{
				Ib:   1,
				Page: 0,
			},
			expectErr: e.ErrNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Call validation through Get()
			err := tc.model.Get()
			// If we want a valid case
			if tc.expectErr == nil && tc.model.Ib != 0 && tc.model.Page != 0 {
				// We expect DB error since DB is not mocked yet
				assert.Error(t, err)
			} else {
				// Otherwise we should get the expected validation error
				assert.Equal(t, tc.expectErr, err)
			}
		})
	}
}

func TestBoardLogGet(t *testing.T) {
	var err error

	// Set config settings
	config.Settings.Limits.PostsPerPage = 10

	mock, err := db.NewTestDb()
	assert.NoError(t, err)
	defer db.CloseDb()

	// Initialize model with parameters
	m := &BoardLogModel{
		Ib:   1,
		Page: 1,
	}

	// Total count
	mock.ExpectQuery(`SELECT count\(\*\) FROM audit WHERE ib_id = \? AND audit_type = 1`).
		WithArgs(m.Ib).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	// Log rows
	now := time.Now()
	logRows := sqlmock.NewRows([]string{
		"user_id", "user_name", "role", "audit_time", "audit_action", "audit_info",
	}).
		AddRow(2, "test", 3, now, "deleted", "Post 1").
		AddRow(1, "admin", 4, now, "banned", "User 2")

	mock.ExpectQuery(`SELECT audit.user_id,user_name,(.+) ORDER BY audit_id DESC LIMIT \?,\?`).
		WithArgs(m.Ib, m.Ib, 0, config.Settings.Limits.PostsPerPage).
		WillReturnRows(logRows)

	// Get the board logs
	err = m.Get()
	assert.NoError(t, err)

	// Check model integrity
	assert.NotEmpty(t, m.Result)
	assert.Equal(t, 2, len(m.Result.Body.Items.([]Log)))
	assert.Equal(t, uint(1), m.Result.Body.CurrentPage)
	assert.Equal(t, uint(5), m.Result.Body.Total)

	logs := m.Result.Body.Items.([]Log)
	assert.Equal(t, uint(2), logs[0].UID)
	assert.Equal(t, "test", logs[0].Name)
	assert.Equal(t, uint(3), logs[0].Group)
	assert.Equal(t, "deleted", logs[0].Action)
	assert.Equal(t, "Post 1", logs[0].Meta)

	assert.Equal(t, uint(1), logs[1].UID)
	assert.Equal(t, "admin", logs[1].Name)
	assert.Equal(t, uint(4), logs[1].Group)
	assert.Equal(t, "banned", logs[1].Action)
	assert.Equal(t, "User 2", logs[1].Meta)

	// Verify that all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBoardLogGetNotFound(t *testing.T) {
	var err error

	// Set config settings
	config.Settings.Limits.PostsPerPage = 10

	mock, err := db.NewTestDb()
	assert.NoError(t, err)
	defer db.CloseDb()

	// Initialize model with parameters for a page that doesn't exist
	m := &BoardLogModel{
		Ib:   1,
		Page: 2, // Page 2, but there's only 1 page of results
	}

	// Total count
	mock.ExpectQuery(`SELECT count\(\*\) FROM audit WHERE ib_id = \? AND audit_type = 1`).
		WithArgs(m.Ib).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	// Get the board logs - should return not found because page > total pages
	err = m.Get()
	assert.Equal(t, e.ErrNotFound, err)

	// Verify that all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBoardLogGetDbError(t *testing.T) {
	var err error

	mock, err := db.NewTestDb()
	assert.NoError(t, err)
	defer db.CloseDb()

	// Initialize model with parameters
	m := &BoardLogModel{
		Ib:   1,
		Page: 1,
	}

	// Total count query fails
	expectedError := errors.New("database error")
	mock.ExpectQuery(`SELECT count\(\*\) FROM audit WHERE ib_id = \? AND audit_type = 1`).
		WithArgs(m.Ib).
		WillReturnError(expectedError)

	// Get the board logs
	err = m.Get()
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)

	// Verify that all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBoardLogGetRowsError(t *testing.T) {
	var err error

	// Set config settings
	config.Settings.Limits.PostsPerPage = 10

	mock, err := db.NewTestDb()
	assert.NoError(t, err)
	defer db.CloseDb()

	// Initialize model with parameters
	m := &BoardLogModel{
		Ib:   1,
		Page: 1,
	}

	// Total count
	mock.ExpectQuery(`SELECT count\(\*\) FROM audit WHERE ib_id = \? AND audit_type = 1`).
		WithArgs(m.Ib).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	// Rows query fails
	expectedError := errors.New("row scan error")
	mock.ExpectQuery(`SELECT audit.user_id,user_name,(.+) ORDER BY audit_id DESC LIMIT \?,\?`).
		WithArgs(m.Ib, m.Ib, 0, config.Settings.Limits.PostsPerPage).
		WillReturnError(expectedError)

	// Get the board logs
	err = m.Get()
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)

	// Verify that all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBoardLogGetScanError(t *testing.T) {
	var err error

	// Set config settings
	config.Settings.Limits.PostsPerPage = 10

	mock, err := db.NewTestDb()
	assert.NoError(t, err)
	defer db.CloseDb()

	// Initialize model with parameters
	m := &BoardLogModel{
		Ib:   1,
		Page: 1,
	}

	// Total count
	mock.ExpectQuery(`SELECT count\(\*\) FROM audit WHERE ib_id = \? AND audit_type = 1`).
		WithArgs(m.Ib).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	// Log rows with scan error (type mismatch)
	logRows := sqlmock.NewRows([]string{
		"user_id", "user_name", "role", "audit_time", "audit_action", "audit_info",
	}).
		AddRow("not a number", "test", 3, time.Now(), "deleted", "Post 1") // This will cause a scan error

	mock.ExpectQuery(`SELECT audit.user_id,user_name,(.+) ORDER BY audit_id DESC LIMIT \?,\?`).
		WithArgs(m.Ib, m.Ib, 0, config.Settings.Limits.PostsPerPage).
		WillReturnRows(logRows)

	// Get the board logs
	err = m.Get()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sql: Scan error") // Check for scan error

	// Verify that all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}
