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

func TestBanFileIsValid(t *testing.T) {
	// Test cases for invalid models
	badmodels := []BanFileModel{
		{Ib: 0, Thread: 1, ID: 1, User: 2, Reason: "test", Hash: "test"},
		{Ib: 1, Thread: 0, ID: 1, User: 2, Reason: "test", Hash: "test"},
		{Ib: 1, Thread: 1, ID: 0, User: 2, Reason: "test", Hash: "test"},
		{Ib: 1, Thread: 1, ID: 1, User: 0, Reason: "test", Hash: "test"},
		{Ib: 1, Thread: 1, ID: 1, User: 1, Reason: "test", Hash: "test"}, // User ID 1 is invalid
		{Ib: 1, Thread: 1, ID: 1, User: 2, Reason: "", Hash: "test"},
		{Ib: 1, Thread: 1, ID: 1, User: 2, Reason: "test", Hash: ""},
	}

	// Test invalid models
	for _, model := range badmodels {
		assert.False(t, model.IsValid(), "Model should be invalid")
	}

	// Test valid model
	goodmodel := BanFileModel{
		Ib:     1,
		Thread: 1,
		ID:     1,
		User:   2,
		Reason: "test reason",
		Hash:   "abcdef1234567890",
	}
	assert.True(t, goodmodel.IsValid(), "Model should be valid")
}

func TestBanFileStatus(t *testing.T) {
	var err error

	// Create a new mock database
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Create expected rows
	rows := sqlmock.NewRows([]string{"image_hash"}).
		AddRow("abcdef1234567890")

	// Set expected query with its expectations
	mock.ExpectQuery(`SELECT image_hash FROM threads
	    INNER JOIN posts ON threads.thread_id = posts.thread_id
	    INNER JOIN images ON posts.post_id = images.post_id
	    WHERE ib_id = \? AND threads.thread_id = \? AND post_num = \? LIMIT 1`).
		WithArgs(1, 1, 1).
		WillReturnRows(rows)

	// Initialize model
	model := BanFileModel{
		Ib:     1,
		Thread: 1,
		ID:     1,
	}

	// Execute the method
	err = model.Status()
	assert.NoError(t, err, "An error was not expected")

	// Verify results
	assert.Equal(t, "abcdef1234567890", model.Hash, "Hash should match expected value")

	// Make sure expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestBanFileStatusNotFound(t *testing.T) {
	var err error

	// Create a new mock database
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Set expected query with its expectations
	mock.ExpectQuery(`SELECT image_hash FROM threads
	    INNER JOIN posts ON threads.thread_id = posts.thread_id
	    INNER JOIN images ON posts.post_id = images.post_id
	    WHERE ib_id = \? AND threads.thread_id = \? AND post_num = \? LIMIT 1`).
		WithArgs(1, 1, 1).
		WillReturnError(sql.ErrNoRows)

	// Initialize model
	model := BanFileModel{
		Ib:     1,
		Thread: 1,
		ID:     1,
	}

	// Execute the method
	err = model.Status()
	
	// Should return ErrNotFound
	if assert.Error(t, err, "An error was expected") {
		assert.Equal(t, e.ErrNotFound, err, "Error should be ErrNotFound")
	}

	// Make sure expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestBanFileStatusError(t *testing.T) {
	var err error

	// Create a new mock database
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Set expected query with its expectations
	mock.ExpectQuery(`SELECT image_hash FROM threads
	    INNER JOIN posts ON threads.thread_id = posts.thread_id
	    INNER JOIN images ON posts.post_id = images.post_id
	    WHERE ib_id = \? AND threads.thread_id = \? AND post_num = \? LIMIT 1`).
		WithArgs(1, 1, 1).
		WillReturnError(errors.New("database error"))

	// Initialize model
	model := BanFileModel{
		Ib:     1,
		Thread: 1,
		ID:     1,
	}

	// Execute the method
	err = model.Status()
	
	// Should return an error
	if assert.Error(t, err, "An error was expected") {
		assert.Contains(t, err.Error(), "database error", "Error should contain the expected message")
	}

	// Make sure expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestBanFilePost(t *testing.T) {
	var err error

	// Create a new mock database
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Set expected exec with its expectations
	mock.ExpectExec("INSERT IGNORE INTO banned_files \\(user_id,ib_id,ban_hash,ban_reason\\) VALUES \\(\\?,\\?,\\?,\\?\\)").
		WithArgs(2, 1, "abcdef1234567890", "test reason").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Initialize model
	model := BanFileModel{
		Ib:     1,
		Thread: 1,
		ID:     1,
		User:   2,
		Reason: "test reason",
		Hash:   "abcdef1234567890",
	}

	// Execute the method
	err = model.Post()
	assert.NoError(t, err, "An error was not expected")

	// Make sure expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestBanFilePostInvalid(t *testing.T) {
	var err error

	// Initialize invalid model (missing Hash)
	model := BanFileModel{
		Ib:     1,
		Thread: 1,
		ID:     1,
		User:   2,
		Reason: "test reason",
		Hash:   "", // Invalid - missing hash
	}

	// Execute the method
	err = model.Post()
	
	// Should return error about invalid model
	if assert.Error(t, err, "An error was expected") {
		assert.Equal(t, errors.New("BanFileModel is not valid"), err, "Error should match expected message")
	}
}

func TestBanFilePostError(t *testing.T) {
	var err error

	// Create a new mock database
	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Set expected exec with its expectations - will return an error
	mock.ExpectExec("INSERT IGNORE INTO banned_files \\(user_id,ib_id,ban_hash,ban_reason\\) VALUES \\(\\?,\\?,\\?,\\?\\)").
		WithArgs(2, 1, "abcdef1234567890", "test reason").
		WillReturnError(errors.New("database error"))

	// Initialize model
	model := BanFileModel{
		Ib:     1,
		Thread: 1,
		ID:     1,
		User:   2,
		Reason: "test reason",
		Hash:   "abcdef1234567890",
	}

	// Execute the method
	err = model.Post()
	
	// Should return an error
	if assert.Error(t, err, "An error was expected") {
		assert.Contains(t, err.Error(), "database error", "Error should contain the expected message")
	}

	// Make sure expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}