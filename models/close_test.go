package models

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"github.com/eirka/eirka-libs/db"
	e "github.com/eirka/eirka-libs/errors"
)

func TestCloseIsValid(t *testing.T) {

	// Test cases for validation
	tests := []struct {
		name  string
		model *CloseModel
		valid bool
	}{
		{
			name: "valid",
			model: &CloseModel{
				ID:     1,
				Name:   "Thread 1",
				Ib:     1,
				Closed: false,
			},
			valid: true,
		},
		{
			name: "missing id",
			model: &CloseModel{
				ID:     0,
				Name:   "Thread 1",
				Ib:     1,
				Closed: false,
			},
			valid: false,
		},
		{
			name: "missing name",
			model: &CloseModel{
				ID:     1,
				Name:   "",
				Ib:     1,
				Closed: false,
			},
			valid: false,
		},
		{
			name: "missing ib",
			model: &CloseModel{
				ID:     1,
				Name:   "Thread 1",
				Ib:     0,
				Closed: false,
			},
			valid: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Check IsValid
			valid := tc.model.IsValid()
			assert.Equal(t, tc.valid, valid)
		})
	}
}

func TestCloseStatus(t *testing.T) {
	var err error

	mock, err := db.NewTestDb()
	assert.NoError(t, err)
	defer db.CloseDb()

	// Initialize model with parameters
	m := &CloseModel{
		ID: 1,
		Ib: 1,
	}

	// Status query successful
	mock.ExpectQuery(`SELECT thread_title, thread_closed FROM threads WHERE thread_id = \? AND ib_id = \? LIMIT 1`).
		WithArgs(m.ID, m.Ib).
		WillReturnRows(sqlmock.NewRows([]string{"thread_title", "thread_closed"}).
			AddRow("Thread 1", false))

	// Get the thread status
	err = m.Status()
	assert.NoError(t, err)

	// Check model integrity
	assert.Equal(t, "Thread 1", m.Name)
	assert.Equal(t, false, m.Closed)

	// Verify that all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCloseStatusNotFound(t *testing.T) {
	var err error

	mock, err := db.NewTestDb()
	assert.NoError(t, err)
	defer db.CloseDb()

	// Initialize model with parameters
	m := &CloseModel{
		ID: 1,
		Ib: 1,
	}

	// Status query not found
	mock.ExpectQuery(`SELECT thread_title, thread_closed FROM threads WHERE thread_id = \? AND ib_id = \? LIMIT 1`).
		WithArgs(m.ID, m.Ib).
		WillReturnError(e.ErrNotFound)

	// Get the thread status
	err = m.Status()
	assert.Error(t, err)
	assert.Equal(t, e.ErrNotFound, err)

	// Verify that all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCloseStatusError(t *testing.T) {
	var err error

	mock, err := db.NewTestDb()
	assert.NoError(t, err)
	defer db.CloseDb()

	// Initialize model with parameters
	m := &CloseModel{
		ID: 1,
		Ib: 1,
	}

	// Status query error
	expectedError := errors.New("database error")
	mock.ExpectQuery(`SELECT thread_title, thread_closed FROM threads WHERE thread_id = \? AND ib_id = \? LIMIT 1`).
		WithArgs(m.ID, m.Ib).
		WillReturnError(expectedError)

	// Get the thread status
	err = m.Status()
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)

	// Verify that all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCloseToggle(t *testing.T) {
	var err error

	mock, err := db.NewTestDb()
	assert.NoError(t, err)
	defer db.CloseDb()

	// Initialize model with parameters
	m := &CloseModel{
		ID:     1,
		Name:   "Thread 1",
		Ib:     1,
		Closed: false,
	}

	// Toggle prepare and exec
	mock.ExpectPrepare(`UPDATE threads SET thread_closed = \? WHERE thread_id = \? AND ib_id = \?`).
		ExpectExec().
		WithArgs(!m.Closed, m.ID, m.Ib).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Toggle the thread
	err = m.Toggle()
	assert.NoError(t, err)

	// Verify that all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCloseToggleInvalid(t *testing.T) {
	var err error

	// Initialize invalid model with parameters
	m := &CloseModel{
		ID:     0, // Invalid ID
		Name:   "Thread 1",
		Ib:     1,
		Closed: false,
	}

	// Toggle the thread
	err = m.Toggle()
	assert.Error(t, err)
	assert.Equal(t, "CloseModel is not valid", err.Error())
}

func TestCloseToggleDbError(t *testing.T) {
	var err error

	mock, err := db.NewTestDb()
	assert.NoError(t, err)
	defer db.CloseDb()

	// Initialize model with parameters
	m := &CloseModel{
		ID:     1,
		Name:   "Thread 1",
		Ib:     1,
		Closed: false,
	}

	// Toggle prepare error
	expectedError := errors.New("prepare error")
	mock.ExpectPrepare(`UPDATE threads SET thread_closed = \? WHERE thread_id = \? AND ib_id = \?`).
		WillReturnError(expectedError)

	// Toggle the thread
	err = m.Toggle()
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)

	// Verify that all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCloseToggleExecError(t *testing.T) {
	var err error

	mock, err := db.NewTestDb()
	assert.NoError(t, err)
	defer db.CloseDb()

	// Initialize model with parameters
	m := &CloseModel{
		ID:     1,
		Name:   "Thread 1",
		Ib:     1,
		Closed: false,
	}

	// Toggle prepare and exec error
	expectedError := errors.New("exec error")
	mock.ExpectPrepare(`UPDATE threads SET thread_closed = \? WHERE thread_id = \? AND ib_id = \?`).
		ExpectExec().
		WithArgs(!m.Closed, m.ID, m.Ib).
		WillReturnError(expectedError)

	// Toggle the thread
	err = m.Toggle()
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)

	// Verify that all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}