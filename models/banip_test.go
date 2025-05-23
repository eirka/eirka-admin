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

func TestBanIPIsValid(t *testing.T) {
	// Test cases for validation
	tests := []struct {
		name  string
		model *BanIPModel
		valid bool
	}{
		{
			name: "valid",
			model: &BanIPModel{
				Ib:     1,
				Thread: 1,
				ID:     1,
				User:   2,
				Reason: "Spam",
				IP:     "10.0.0.1",
			},
			valid: true,
		},
		{
			name: "missing ib",
			model: &BanIPModel{
				Ib:     0, // Invalid - missing board ID
				Thread: 1,
				ID:     1,
				User:   2,
				Reason: "Spam",
				IP:     "10.0.0.1",
			},
			valid: false,
		},
		{
			name: "missing thread",
			model: &BanIPModel{
				Ib:     1,
				Thread: 0, // Invalid - missing thread ID
				ID:     1,
				User:   2,
				Reason: "Spam",
				IP:     "10.0.0.1",
			},
			valid: false,
		},
		{
			name: "missing post id",
			model: &BanIPModel{
				Ib:     1,
				Thread: 1,
				ID:     0, // Invalid - missing post ID
				User:   2,
				Reason: "Spam",
				IP:     "10.0.0.1",
			},
			valid: false,
		},
		{
			name: "invalid user - zero",
			model: &BanIPModel{
				Ib:     1,
				Thread: 1,
				ID:     1,
				User:   0, // Invalid - missing user ID
				Reason: "Spam",
				IP:     "10.0.0.1",
			},
			valid: false,
		},
		{
			name: "invalid user - anonymous",
			model: &BanIPModel{
				Ib:     1,
				Thread: 1,
				ID:     1,
				User:   1, // Invalid - anonymous user (ID 1)
				Reason: "Spam",
				IP:     "10.0.0.1",
			},
			valid: false,
		},
		{
			name: "missing reason",
			model: &BanIPModel{
				Ib:     1,
				Thread: 1,
				ID:     1,
				User:   2,
				Reason: "", // Invalid - missing reason
				IP:     "10.0.0.1",
			},
			valid: false,
		},
		{
			name: "missing IP",
			model: &BanIPModel{
				Ib:     1,
				Thread: 1,
				ID:     1,
				User:   2,
				Reason: "Spam",
				IP:     "", // Invalid - missing IP
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

func TestBanIPStatus(t *testing.T) {
	var err error

	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Initialize model with parameters
	m := &BanIPModel{
		Ib:     1,
		Thread: 1,
		ID:     1,
	}

	// Status query successful - IP lookup
	mock.ExpectQuery(`SELECT post_ip FROM threads
	    INNER JOIN posts ON threads.thread_id = posts.thread_id
	    WHERE ib_id = \? AND threads.thread_id = \? AND post_num = \? LIMIT 1`).
		WithArgs(m.Ib, m.Thread, m.ID).
		WillReturnRows(sqlmock.NewRows([]string{"post_ip"}).
			AddRow("10.0.0.1"))

	// Get the status
	err = m.Status()
	assert.NoError(t, err, "No error should be returned")

	// Check model
	assert.Equal(t, "10.0.0.1", m.IP, "IP should be correctly retrieved")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestBanIPStatusNotFound(t *testing.T) {
	var err error

	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Initialize model with parameters
	m := &BanIPModel{
		Ib:     1,
		Thread: 1,
		ID:     1,
	}

	// Status query not found
	mock.ExpectQuery(`SELECT post_ip FROM threads
	    INNER JOIN posts ON threads.thread_id = posts.thread_id
	    WHERE ib_id = \? AND threads.thread_id = \? AND post_num = \? LIMIT 1`).
		WithArgs(m.Ib, m.Thread, m.ID).
		WillReturnError(sql.ErrNoRows)

	// Get the status
	err = m.Status()
	assert.Error(t, err, "Error should be returned")
	assert.Equal(t, e.ErrNotFound, err, "Error should be ErrNotFound")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestBanIPStatusError(t *testing.T) {
	var err error

	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Initialize model with parameters
	m := &BanIPModel{
		Ib:     1,
		Thread: 1,
		ID:     1,
	}

	// Status query error
	expectedError := errors.New("database error")
	mock.ExpectQuery(`SELECT post_ip FROM threads
	    INNER JOIN posts ON threads.thread_id = posts.thread_id
	    WHERE ib_id = \? AND threads.thread_id = \? AND post_num = \? LIMIT 1`).
		WithArgs(m.Ib, m.Thread, m.ID).
		WillReturnError(expectedError)

	// Get the status
	err = m.Status()
	assert.Error(t, err, "Error should be returned")
	assert.Equal(t, expectedError, err, "Error should match the expected error")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestBanIPPost(t *testing.T) {
	var err error

	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Initialize model with parameters
	m := &BanIPModel{
		Ib:     1,
		Thread: 1,
		ID:     1,
		User:   2,
		Reason: "Spam",
		IP:     "10.0.0.1",
	}

	// Post exec
	mock.ExpectExec("INSERT IGNORE INTO banned_ips \\(user_id,ib_id,ban_ip,ban_reason\\) VALUES \\(\\?,\\?,\\?,\\?\\)").
		WithArgs(m.User, m.Ib, m.IP, m.Reason).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Post the ban
	err = m.Post()
	assert.NoError(t, err, "No error should be returned")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestBanIPPostInvalid(t *testing.T) {
	// Initialize invalid model
	m := &BanIPModel{
		Ib:     1,
		Thread: 1,
		ID:     1,
		User:   0, // Invalid user ID
		Reason: "Spam",
		IP:     "10.0.0.1",
	}

	// Post the ban
	err := m.Post()
	assert.Error(t, err, "Error should be returned")
	assert.Equal(t, "BanIPModel is not valid", err.Error(), "Error message should match expected value")
}

func TestBanIPPostError(t *testing.T) {
	var err error

	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Initialize model with parameters
	m := &BanIPModel{
		Ib:     1,
		Thread: 1,
		ID:     1,
		User:   2,
		Reason: "Spam",
		IP:     "10.0.0.1",
	}

	// Post exec error
	expectedError := errors.New("database error")
	mock.ExpectExec("INSERT IGNORE INTO banned_ips \\(user_id,ib_id,ban_ip,ban_reason\\) VALUES \\(\\?,\\?,\\?,\\?\\)").
		WithArgs(m.User, m.Ib, m.IP, m.Reason).
		WillReturnError(expectedError)

	// Post the ban
	err = m.Post()
	assert.Error(t, err, "Error should be returned")
	assert.Equal(t, expectedError, err, "Error should match the expected error")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}
