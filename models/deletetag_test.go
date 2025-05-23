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

func TestDeleteTagIsValid(t *testing.T) {
	// Test cases for validation
	tests := []struct {
		name  string
		model *DeleteTagModel
		valid bool
	}{
		{
			name: "valid",
			model: &DeleteTagModel{
				ID:   1,
				Name: "test tag",
				Ib:   1,
			},
			valid: true,
		},
		{
			name: "missing id",
			model: &DeleteTagModel{
				ID:   0, // Invalid - missing tag ID
				Name: "test tag",
				Ib:   1,
			},
			valid: false,
		},
		{
			name: "missing name",
			model: &DeleteTagModel{
				ID:   1,
				Name: "", // Invalid - missing tag name
				Ib:   1,
			},
			valid: false,
		},
		{
			name: "missing ib",
			model: &DeleteTagModel{
				ID:   1,
				Name: "test tag",
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

func TestDeleteTagStatus(t *testing.T) {
	var err error

	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Initialize model with parameters
	m := &DeleteTagModel{
		ID: 1,
		Ib: 1,
	}

	// Status query successful
	mock.ExpectQuery(`SELECT tag_name FROM tags WHERE tag_id = \? AND ib_id = \? LIMIT 1`).
		WithArgs(m.ID, m.Ib).
		WillReturnRows(sqlmock.NewRows([]string{"tag_name"}).
			AddRow("test tag"))

	// Get the status
	err = m.Status()
	assert.NoError(t, err, "No error should be returned")

	// Check model
	assert.Equal(t, "test tag", m.Name, "Tag name should be correctly retrieved")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestDeleteTagStatusNotFound(t *testing.T) {
	var err error

	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Initialize model with parameters
	m := &DeleteTagModel{
		ID: 1,
		Ib: 1,
	}

	// Status query not found
	mock.ExpectQuery(`SELECT tag_name FROM tags WHERE tag_id = \? AND ib_id = \? LIMIT 1`).
		WithArgs(m.ID, m.Ib).
		WillReturnError(sql.ErrNoRows)

	// Get the status
	err = m.Status()
	assert.Error(t, err, "Error should be returned")
	assert.Equal(t, e.ErrNotFound, err, "Error should be ErrNotFound")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestDeleteTagStatusError(t *testing.T) {
	var err error

	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Initialize model with parameters
	m := &DeleteTagModel{
		ID: 1,
		Ib: 1,
	}

	// Status query error
	expectedError := errors.New("database error")
	mock.ExpectQuery(`SELECT tag_name FROM tags WHERE tag_id = \? AND ib_id = \? LIMIT 1`).
		WithArgs(m.ID, m.Ib).
		WillReturnError(expectedError)

	// Get the status
	err = m.Status()
	assert.Error(t, err, "Error should be returned")
	assert.Contains(t, err.Error(), expectedError.Error(), "Error should contain the expected error message")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestDeleteTagDelete(t *testing.T) {
	var err error

	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Initialize model with parameters
	m := &DeleteTagModel{
		ID:   1,
		Name: "test tag",
		Ib:   1,
	}

	// Delete prepare and exec
	mock.ExpectPrepare(`DELETE FROM tags WHERE tag_id= \? AND ib_id = \? LIMIT 1`).
		ExpectExec().
		WithArgs(m.ID, m.Ib).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Delete the tag
	err = m.Delete()
	assert.NoError(t, err, "No error should be returned")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestDeleteTagDeleteInvalid(t *testing.T) {
	// Initialize invalid model
	m := &DeleteTagModel{
		ID:   0, // Invalid ID
		Name: "test tag",
		Ib:   1,
	}

	// Delete the tag
	err := m.Delete()
	assert.Error(t, err, "Error should be returned")
	assert.Equal(t, "DeleteTagModel is not valid", err.Error(), "Error message should match expected value")
}

func TestDeleteTagDeleteGetDbError(t *testing.T) {
	var err error

	// Close the database connection to force a GetDb error
	db.CloseDb()

	// Initialize model with parameters
	m := &DeleteTagModel{
		ID:   1,
		Name: "test tag",
		Ib:   1,
	}

	// Delete the tag - should encounter GetDb error
	err = m.Delete()
	assert.Error(t, err, "Error should be returned")
	assert.Contains(t, err.Error(), "database is closed", "Error should indicate database is closed")
}

func TestDeleteTagDeletePrepareError(t *testing.T) {
	var err error

	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Initialize model with parameters
	m := &DeleteTagModel{
		ID:   1,
		Name: "test tag",
		Ib:   1,
	}

	// Delete prepare error
	expectedError := errors.New("prepare error")
	mock.ExpectPrepare(`DELETE FROM tags WHERE tag_id= \? AND ib_id = \? LIMIT 1`).
		WillReturnError(expectedError)

	// Delete the tag
	err = m.Delete()
	assert.Error(t, err, "Error should be returned")
	assert.Contains(t, err.Error(), expectedError.Error(), "Error should contain the expected error message")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestDeleteTagDeleteExecError(t *testing.T) {
	var err error

	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Initialize model with parameters
	m := &DeleteTagModel{
		ID:   1,
		Name: "test tag",
		Ib:   1,
	}

	// Delete prepare and exec error
	expectedError := errors.New("exec error")
	mock.ExpectPrepare(`DELETE FROM tags WHERE tag_id= \? AND ib_id = \? LIMIT 1`).
		ExpectExec().
		WithArgs(m.ID, m.Ib).
		WillReturnError(expectedError)

	// Delete the tag
	err = m.Delete()
	assert.Error(t, err, "Error should be returned")
	assert.Contains(t, err.Error(), expectedError.Error(), "Error should contain the expected error message")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}
