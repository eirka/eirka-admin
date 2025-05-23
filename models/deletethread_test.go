package models

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"github.com/eirka/eirka-libs/db"
	e "github.com/eirka/eirka-libs/errors"
)

func TestDeleteThreadIsValid(t *testing.T) {
	// Test cases for validation
	tests := []struct {
		name  string
		model *DeleteThreadModel
		valid bool
	}{
		{
			name: "valid",
			model: &DeleteThreadModel{
				ID:   1,
				Name: "test thread",
				Ib:   1,
			},
			valid: true,
		},
		{
			name: "missing id",
			model: &DeleteThreadModel{
				ID:   0, // Invalid - missing thread ID
				Name: "test thread",
				Ib:   1,
			},
			valid: false,
		},
		{
			name: "missing name",
			model: &DeleteThreadModel{
				ID:   1,
				Name: "", // Invalid - missing thread name
				Ib:   1,
			},
			valid: false,
		},
		{
			name: "missing ib",
			model: &DeleteThreadModel{
				ID:   1,
				Name: "test thread",
				Ib:   0, // Invalid - missing board ID
			},
			valid: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Check IsValid
			valid := tc.model.IsValid()
			assert.Equal(t, tc.valid, valid, "IsValid result should match expected value")
		})
	}
}

func TestDeleteThreadStatus(t *testing.T) {
	var err error

	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Initialize model with parameters
	m := &DeleteThreadModel{
		ID: 1,
		Ib: 1,
	}

	// Status query successful
	mock.ExpectQuery(`SELECT thread_title, thread_deleted FROM threads WHERE thread_id = \? AND ib_id = \? LIMIT 1`).
		WithArgs(m.ID, m.Ib).
		WillReturnRows(sqlmock.NewRows([]string{"thread_title", "thread_deleted"}).
			AddRow("test thread", false))

	// Get the status
	err = m.Status()
	assert.NoError(t, err, "No error should be returned")

	// Check model
	assert.Equal(t, "test thread", m.Name, "Thread name should be correctly retrieved")
	assert.Equal(t, false, m.Deleted, "Thread deleted status should be correctly retrieved")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestDeleteThreadStatusNotFound(t *testing.T) {
	var err error

	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Initialize model with parameters
	m := &DeleteThreadModel{
		ID: 1,
		Ib: 1,
	}

	// Status query not found
	mock.ExpectQuery(`SELECT thread_title, thread_deleted FROM threads WHERE thread_id = \? AND ib_id = \? LIMIT 1`).
		WithArgs(m.ID, m.Ib).
		WillReturnError(sql.ErrNoRows)

	// Get the status
	err = m.Status()
	assert.Error(t, err, "Error should be returned")
	assert.Equal(t, e.ErrNotFound, err, "Error should be ErrNotFound")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestDeleteThreadStatusError(t *testing.T) {
	var err error

	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Initialize model with parameters
	m := &DeleteThreadModel{
		ID: 1,
		Ib: 1,
	}

	// Status query error
	expectedError := errors.New("database error")
	mock.ExpectQuery(`SELECT thread_title, thread_deleted FROM threads WHERE thread_id = \? AND ib_id = \? LIMIT 1`).
		WithArgs(m.ID, m.Ib).
		WillReturnError(expectedError)

	// Get the status
	err = m.Status()
	assert.Error(t, err, "Error should be returned")
	assert.Contains(t, err.Error(), expectedError.Error(), "Error should contain the expected error message")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestDeleteThreadDelete(t *testing.T) {
	var err error

	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Initialize model with parameters
	m := &DeleteThreadModel{
		ID:      1,
		Name:    "test thread",
		Ib:      1,
		Deleted: false,
	}

	// Delete prepare and exec
	mock.ExpectPrepare(`UPDATE threads SET thread_deleted = \? WHERE thread_id = \? AND ib_id = \?`).
		ExpectExec().
		WithArgs(!m.Deleted, m.ID, m.Ib).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Delete the thread
	err = m.Delete()
	assert.NoError(t, err, "No error should be returned")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestDeleteThreadDeleteInvalid(t *testing.T) {
	// Initialize invalid model
	m := &DeleteThreadModel{
		ID:   0, // Invalid ID
		Name: "test thread",
		Ib:   1,
	}

	// Delete the thread
	err := m.Delete()
	assert.Error(t, err, "Error should be returned")
	assert.Equal(t, "DeleteThreadModel is not valid", err.Error(), "Error message should match expected value")
}

func TestDeleteThreadDeleteGetDbError(t *testing.T) {
	var err error

	// Close the database connection to force a GetDb error
	db.CloseDb()

	// Initialize model with parameters
	m := &DeleteThreadModel{
		ID:      1,
		Name:    "test thread",
		Ib:      1,
		Deleted: false,
	}

	// Delete the thread - should encounter GetDb error
	err = m.Delete()
	assert.Error(t, err, "Error should be returned")
	assert.Contains(t, err.Error(), "database is closed", "Error should indicate database is closed")
}

func TestDeleteThreadDeletePrepareError(t *testing.T) {
	var err error

	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Initialize model with parameters
	m := &DeleteThreadModel{
		ID:      1,
		Name:    "test thread",
		Ib:      1,
		Deleted: false,
	}

	// Delete prepare error
	expectedError := errors.New("prepare error")
	mock.ExpectPrepare(`UPDATE threads SET thread_deleted = \? WHERE thread_id = \? AND ib_id = \?`).
		WillReturnError(expectedError)

	// Delete the thread
	err = m.Delete()
	assert.Error(t, err, "Error should be returned")
	assert.Contains(t, err.Error(), expectedError.Error(), "Error should contain the expected error message")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestDeleteThreadDeleteExecError(t *testing.T) {
	var err error

	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Initialize model with parameters
	m := &DeleteThreadModel{
		ID:      1,
		Name:    "test thread",
		Ib:      1,
		Deleted: false,
	}

	// Delete prepare and exec error
	expectedError := errors.New("exec error")
	mock.ExpectPrepare(`UPDATE threads SET thread_deleted = \? WHERE thread_id = \? AND ib_id = \?`).
		ExpectExec().
		WithArgs(!m.Deleted, m.ID, m.Ib).
		WillReturnError(expectedError)

	// Delete the thread
	err = m.Delete()
	assert.Error(t, err, "Error should be returned")
	assert.Contains(t, err.Error(), expectedError.Error(), "Error should contain the expected error message")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}
