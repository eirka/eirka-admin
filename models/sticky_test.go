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

func TestStickyIsValid(t *testing.T) {
	badsticky := []StickyModel{
		{ID: 0, Name: "test", Ib: 1, Sticky: false},
		{ID: 1, Name: "", Ib: 1, Sticky: false},
		{ID: 1, Name: "test", Ib: 0, Sticky: false},
	}

	for _, sticky := range badsticky {
		assert.False(t, sticky.IsValid(), "Should be false")
	}

	goodsticky := []StickyModel{
		{ID: 1, Name: "test", Ib: 1, Sticky: false},
		{ID: 1, Name: "test", Ib: 1, Sticky: true},
	}

	for _, sticky := range goodsticky {
		assert.True(t, sticky.IsValid(), "Should be true")
	}
}

func TestStickyStatus(t *testing.T) {
	var err error

	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	rows := sqlmock.NewRows([]string{"thread_title", "thread_sticky"}).
		AddRow("test thread", 0)

	mock.ExpectQuery("SELECT thread_title, thread_sticky FROM threads WHERE thread_id = \\? AND ib_id = \\? LIMIT 1").
		WithArgs(1, 1).
		WillReturnRows(rows)

	sticky := StickyModel{
		ID: 1,
		Ib: 1,
	}

	err = sticky.Status()
	assert.NoError(t, err, "An error was not expected")

	assert.Equal(t, "test thread", sticky.Name, "Name should match")
	assert.Equal(t, false, sticky.Sticky, "Sticky status should match")

	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestStickyStatusNotFound(t *testing.T) {
	var err error

	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	mock.ExpectQuery("SELECT thread_title, thread_sticky FROM threads WHERE thread_id = \\? AND ib_id = \\? LIMIT 1").
		WithArgs(1, 1).
		WillReturnError(sql.ErrNoRows)

	sticky := StickyModel{
		ID: 1,
		Ib: 1,
	}

	err = sticky.Status()
	if assert.Error(t, err, "An error was expected") {
		assert.Equal(t, e.ErrNotFound, err, "Error should match")
	}

	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestStickyStatusError(t *testing.T) {
	var err error

	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	mock.ExpectQuery("SELECT thread_title, thread_sticky FROM threads WHERE thread_id = \\? AND ib_id = \\? LIMIT 1").
		WithArgs(1, 1).
		WillReturnError(errors.New("database error"))

	sticky := StickyModel{
		ID: 1,
		Ib: 1,
	}

	err = sticky.Status()
	if assert.Error(t, err, "An error was expected") {
		assert.Contains(t, err.Error(), "database error", "Error should contain the expected message")
	}

	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

// Skipping TestStickyStatusDatabaseConnectionError for now as it requires
// direct modification of a package function which is not permitted in Go tests

func TestStickyToggle(t *testing.T) {
	var err error

	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	mock.ExpectPrepare("UPDATE threads SET thread_sticky = \\? WHERE thread_id = \\? AND ib_id = \\?").
		ExpectExec().
		WithArgs(true, 1, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	sticky := StickyModel{
		ID:     1,
		Name:   "test",
		Ib:     1,
		Sticky: false,
	}

	err = sticky.Toggle()
	assert.NoError(t, err, "An error was not expected")

	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestStickyToggleInvalid(t *testing.T) {
	var err error

	sticky := StickyModel{
		ID:     0, // Invalid ID
		Name:   "test",
		Ib:     1,
		Sticky: false,
	}

	err = sticky.Toggle()
	if assert.Error(t, err, "An error was expected") {
		assert.Equal(t, errors.New("StickyModel is not valid"), err, "Error should match")
	}
}

func TestStickyTogglePrepareError(t *testing.T) {
	var err error

	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	mock.ExpectPrepare("UPDATE threads SET thread_sticky = \\? WHERE thread_id = \\? AND ib_id = \\?").
		WillReturnError(errors.New("prepare error"))

	sticky := StickyModel{
		ID:     1,
		Name:   "test",
		Ib:     1,
		Sticky: false,
	}

	err = sticky.Toggle()
	if assert.Error(t, err, "An error was expected") {
		assert.Contains(t, err.Error(), "prepare error", "Error should contain the expected message")
	}

	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestStickyToggleExecError(t *testing.T) {
	var err error

	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	mock.ExpectPrepare("UPDATE threads SET thread_sticky = \\? WHERE thread_id = \\? AND ib_id = \\?").
		ExpectExec().
		WithArgs(true, 1, 1).
		WillReturnError(errors.New("exec error"))

	sticky := StickyModel{
		ID:     1,
		Name:   "test",
		Ib:     1,
		Sticky: false,
	}

	err = sticky.Toggle()
	if assert.Error(t, err, "An error was expected") {
		assert.Contains(t, err.Error(), "exec error", "Error should contain the expected message")
	}

	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

// Skipping TestStickyToggleDatabaseConnectionError for now as it requires
// direct modification of a package function which is not permitted in Go tests