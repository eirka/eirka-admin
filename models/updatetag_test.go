package models

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"github.com/eirka/eirka-libs/config"
	"github.com/eirka/eirka-libs/db"
	e "github.com/eirka/eirka-libs/errors"
)

func TestUpdateTagIsValid(t *testing.T) {
	// Test cases for validation
	tests := []struct {
		name  string
		model *UpdateTagModel
		valid bool
	}{
		{
			name: "valid",
			model: &UpdateTagModel{
				ID:      1,
				Ib:      1,
				Tag:     "test tag",
				TagType: 1,
			},
			valid: true,
		},
		{
			name: "missing id",
			model: &UpdateTagModel{
				ID:      0, // Invalid - missing tag ID
				Ib:      1,
				Tag:     "test tag",
				TagType: 1,
			},
			valid: false,
		},
		{
			name: "missing ib",
			model: &UpdateTagModel{
				ID:      1,
				Ib:      0, // Invalid - missing board ID
				Tag:     "test tag",
				TagType: 1,
			},
			valid: false,
		},
		{
			name: "missing tag",
			model: &UpdateTagModel{
				ID:      1,
				Ib:      1,
				Tag:     "", // Invalid - missing tag name
				TagType: 1,
			},
			valid: false,
		},
		{
			name: "missing tag type",
			model: &UpdateTagModel{
				ID:      1,
				Ib:      1,
				Tag:     "test tag",
				TagType: 0, // Invalid - missing tag type
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

func TestUpdateTagValidateInput(t *testing.T) {
	// Save original config values and restore after test
	originalTagMinLength := config.Settings.Limits.TagMinLength
	originalTagMaxLength := config.Settings.Limits.TagMaxLength
	config.Settings.Limits.TagMinLength = 3
	config.Settings.Limits.TagMaxLength = 20
	defer func() {
		config.Settings.Limits.TagMinLength = originalTagMinLength
		config.Settings.Limits.TagMaxLength = originalTagMaxLength
	}()

	// Test cases for validation
	tests := []struct {
		name  string
		model *UpdateTagModel
		err   error
	}{
		{
			name: "valid",
			model: &UpdateTagModel{
				ID:      1,
				Ib:      1,
				Tag:     "test tag",
				TagType: 1,
			},
			err: nil,
		},
		{
			name: "missing ib",
			model: &UpdateTagModel{
				ID:      1,
				Ib:      0, // Invalid - missing board ID
				Tag:     "test tag",
				TagType: 1,
			},
			err: e.ErrInvalidParam,
		},
		{
			name: "missing tag type",
			model: &UpdateTagModel{
				ID:      1,
				Ib:      1,
				Tag:     "test tag",
				TagType: 0, // Invalid - missing tag type
			},
			err: e.ErrInvalidParam,
		},
		{
			name: "empty tag",
			model: &UpdateTagModel{
				ID:      1,
				Ib:      1,
				Tag:     "", // Invalid - empty tag name
				TagType: 1,
			},
			err: e.ErrNoTagName,
		},
		{
			name: "tag too short",
			model: &UpdateTagModel{
				ID:      1,
				Ib:      1,
				Tag:     "ab", // Invalid - tag too short
				TagType: 1,
			},
			err: e.ErrTagShort,
		},
		{
			name: "tag too long",
			model: &UpdateTagModel{
				ID:      1,
				Ib:      1,
				Tag:     "this is a very long tag name that exceeds the maximum allowed length", // Invalid - tag too long
				TagType: 1,
			},
			err: e.ErrTagLong,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Check ValidateInput
			err := tc.model.ValidateInput()
			if tc.err == nil {
				assert.NoError(t, err, "No error should be returned")
			} else {
				assert.Equal(t, tc.err, err, "Error should match expected error")
			}
		})
	}
}

func TestUpdateTagStatus(t *testing.T) {
	var err error

	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Initialize model with parameters
	m := &UpdateTagModel{
		ID:      1,
		Ib:      1,
		Tag:     "test tag",
		TagType: 1,
	}

	// Status query successful - no duplicate tag
	mock.ExpectQuery(`select count\(\*\) from tags where tag_name = \? AND ib_id = \? AND NOT tag_id = \?`).
		WithArgs(m.Tag, m.Ib, m.ID).
		WillReturnRows(sqlmock.NewRows([]string{"count(*)"}).
			AddRow(0))

	// Get the status
	err = m.Status()
	assert.NoError(t, err, "No error should be returned")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestUpdateTagStatusDuplicate(t *testing.T) {
	var err error

	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Initialize model with parameters
	m := &UpdateTagModel{
		ID:      1,
		Ib:      1,
		Tag:     "test tag",
		TagType: 1,
	}

	// Status query - duplicate tag exists
	mock.ExpectQuery(`select count\(\*\) from tags where tag_name = \? AND ib_id = \? AND NOT tag_id = \?`).
		WithArgs(m.Tag, m.Ib, m.ID).
		WillReturnRows(sqlmock.NewRows([]string{"count(*)"}).
			AddRow(1))

	// Get the status
	err = m.Status()
	assert.Error(t, err, "Error should be returned")
	assert.Equal(t, e.ErrDuplicateTag, err, "Error should be ErrDuplicateTag")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestUpdateTagStatusError(t *testing.T) {
	var err error

	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Initialize model with parameters
	m := &UpdateTagModel{
		ID:      1,
		Ib:      1,
		Tag:     "test tag",
		TagType: 1,
	}

	// Status query error
	expectedError := errors.New("database error")
	mock.ExpectQuery(`select count\(\*\) from tags where tag_name = \? AND ib_id = \? AND NOT tag_id = \?`).
		WithArgs(m.Tag, m.Ib, m.ID).
		WillReturnError(expectedError)

	// Get the status
	err = m.Status()
	assert.Error(t, err, "Error should be returned")
	assert.Contains(t, err.Error(), expectedError.Error(), "Error should contain the expected error message")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestUpdateTagUpdate(t *testing.T) {
	var err error

	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Initialize model with parameters
	m := &UpdateTagModel{
		ID:      1,
		Ib:      1,
		Tag:     "test tag",
		TagType: 1,
	}

	// Update prepare and exec
	mock.ExpectPrepare("UPDATE tags SET tag_name= \\?, tagtype_id= \\? WHERE tag_id = \\? AND ib_id = \\?").
		ExpectExec().
		WithArgs(m.Tag, m.TagType, m.ID, m.Ib).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Update the tag
	err = m.Update()
	assert.NoError(t, err, "No error should be returned")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestUpdateTagUpdateInvalid(t *testing.T) {
	// Initialize invalid model
	m := &UpdateTagModel{
		ID:      0, // Invalid ID
		Ib:      1,
		Tag:     "test tag",
		TagType: 1,
	}

	// Update the tag
	err := m.Update()
	assert.Error(t, err, "Error should be returned")
	assert.Equal(t, "UpdateTagModel is not valid", err.Error(), "Error message should match expected value")
}

func TestUpdateTagUpdateGetDbError(t *testing.T) {
	var err error

	// Close the database connection to force a GetDb error
	db.CloseDb()

	// Initialize model with parameters
	m := &UpdateTagModel{
		ID:      1,
		Ib:      1,
		Tag:     "test tag",
		TagType: 1,
	}

	// Update the tag - should encounter GetDb error
	err = m.Update()
	assert.Error(t, err, "Error should be returned")
	assert.Contains(t, err.Error(), "database is closed", "Error should indicate database is closed")
}

func TestUpdateTagUpdatePrepareError(t *testing.T) {
	var err error

	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Initialize model with parameters
	m := &UpdateTagModel{
		ID:      1,
		Ib:      1,
		Tag:     "test tag",
		TagType: 1,
	}

	// Update prepare error
	expectedError := errors.New("prepare error")
	mock.ExpectPrepare("UPDATE tags SET tag_name= \\?, tagtype_id= \\? WHERE tag_id = \\? AND ib_id = \\?").
		WillReturnError(expectedError)

	// Update the tag
	err = m.Update()
	assert.Error(t, err, "Error should be returned")
	assert.Contains(t, err.Error(), expectedError.Error(), "Error should contain the expected error message")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestUpdateTagUpdateExecError(t *testing.T) {
	var err error

	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Initialize model with parameters
	m := &UpdateTagModel{
		ID:      1,
		Ib:      1,
		Tag:     "test tag",
		TagType: 1,
	}

	// Update prepare and exec error
	expectedError := errors.New("exec error")
	mock.ExpectPrepare("UPDATE tags SET tag_name= \\?, tagtype_id= \\? WHERE tag_id = \\? AND ib_id = \\?").
		ExpectExec().
		WithArgs(m.Tag, m.TagType, m.ID, m.Ib).
		WillReturnError(expectedError)

	// Update the tag
	err = m.Update()
	assert.Error(t, err, "Error should be returned")
	assert.Contains(t, err.Error(), expectedError.Error(), "Error should contain the expected error message")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}
