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

func TestDeletePostIsValid(t *testing.T) {
	// Test cases for validation
	tests := []struct {
		name  string
		model *DeletePostModel
		valid bool
	}{
		{
			name: "valid",
			model: &DeletePostModel{
				Thread: 1,
				ID:     1,
				Ib:     1,
				Name:   "Test Thread",
			},
			valid: true,
		},
		{
			name: "missing thread",
			model: &DeletePostModel{
				Thread: 0, // Invalid - missing thread ID
				ID:     1,
				Ib:     1,
				Name:   "Test Thread",
			},
			valid: false,
		},
		{
			name: "missing post id",
			model: &DeletePostModel{
				Thread: 1,
				ID:     0, // Invalid - missing post ID
				Ib:     1,
				Name:   "Test Thread",
			},
			valid: false,
		},
		{
			name: "missing ib",
			model: &DeletePostModel{
				Thread: 1,
				ID:     1,
				Ib:     0, // Invalid - missing board ID
				Name:   "Test Thread",
			},
			valid: false,
		},
		{
			name: "missing name",
			model: &DeletePostModel{
				Thread: 1,
				ID:     1,
				Ib:     1,
				Name:   "", // Invalid - missing thread name
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

func TestDeletePostStatus(t *testing.T) {
	var err error

	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Initialize model with parameters
	m := &DeletePostModel{
		Thread: 1,
		Ib:     1,
	}

	// Status query successful
	mock.ExpectQuery(`SELECT thread_title, post_deleted FROM threads
		INNER JOIN posts on threads.thread_id = posts.thread_id
		WHERE threads.thread_id = \? AND ib_id = \? LIMIT 1`).
		WithArgs(m.Thread, m.Ib).
		WillReturnRows(sqlmock.NewRows([]string{"thread_title", "post_deleted"}).
			AddRow("Test Thread", false))

	// Get the status
	err = m.Status()
	assert.NoError(t, err, "No error should be returned")

	// Check model
	assert.Equal(t, "Test Thread", m.Name, "Thread name should be correctly retrieved")
	assert.Equal(t, false, m.Deleted, "Post deleted status should be correctly retrieved")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestDeletePostStatusNotFound(t *testing.T) {
	var err error

	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Initialize model with parameters
	m := &DeletePostModel{
		Thread: 1,
		Ib:     1,
	}

	// Status query not found
	mock.ExpectQuery(`SELECT thread_title, post_deleted FROM threads
		INNER JOIN posts on threads.thread_id = posts.thread_id
		WHERE threads.thread_id = \? AND ib_id = \? LIMIT 1`).
		WithArgs(m.Thread, m.Ib).
		WillReturnError(sql.ErrNoRows)

	// Get the status
	err = m.Status()
	assert.Error(t, err, "Error should be returned")
	assert.Equal(t, e.ErrNotFound, err, "Error should be ErrNotFound")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestDeletePostStatusError(t *testing.T) {
	var err error

	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Initialize model with parameters
	m := &DeletePostModel{
		Thread: 1,
		Ib:     1,
	}

	// Status query error
	expectedError := errors.New("database error")
	mock.ExpectQuery(`SELECT thread_title, post_deleted FROM threads
		INNER JOIN posts on threads.thread_id = posts.thread_id
		WHERE threads.thread_id = \? AND ib_id = \? LIMIT 1`).
		WithArgs(m.Thread, m.Ib).
		WillReturnError(expectedError)

	// Get the status
	err = m.Status()
	assert.Error(t, err, "Error should be returned")
	assert.Equal(t, expectedError, err, "Error should match the expected error")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestDeletePostDelete(t *testing.T) {
	var err error

	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Initialize model with parameters
	m := &DeletePostModel{
		Thread:  1,
		ID:      1,
		Ib:      1,
		Name:    "Test Thread",
		Deleted: false,
	}

	// Begin transaction
	mock.ExpectBegin()

	// Expect query for post count - return multiple posts
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM posts
		WHERE thread_id = \? AND post_deleted = 0`).
		WithArgs(m.Thread).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).
			AddRow(5)) // Multiple posts in thread

	// Delete prepare and exec - set post to deleted
	mock.ExpectPrepare(`UPDATE posts SET post_deleted = \?
		WHERE posts.thread_id = \? AND posts.post_num = \? LIMIT 1`).
		ExpectExec().
		WithArgs(!m.Deleted, m.Thread, m.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Commit transaction
	mock.ExpectCommit()

	// Delete the post
	err = m.Delete()
	assert.NoError(t, err, "No error should be returned")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestDeletePostDeleteLastPost(t *testing.T) {
	var err error

	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Initialize model with parameters
	m := &DeletePostModel{
		Thread:  1,
		ID:      1,
		Ib:      1,
		Name:    "Test Thread",
		Deleted: false,
	}

	// Begin transaction
	mock.ExpectBegin()

	// Expect query for post count - return only one post
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM posts
		WHERE thread_id = \? AND post_deleted = 0`).
		WithArgs(m.Thread).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).
			AddRow(1)) // Only one post in thread

	// Expect thread deletion as well
	mock.ExpectPrepare(`UPDATE threads SET thread_deleted = 1
		WHERE thread_id = \? AND ib_id = \? LIMIT 1`).
		ExpectExec().
		WithArgs(m.Thread, m.Ib).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Delete prepare and exec - set post to deleted
	mock.ExpectPrepare(`UPDATE posts SET post_deleted = \?
		WHERE posts.thread_id = \? AND posts.post_num = \? LIMIT 1`).
		ExpectExec().
		WithArgs(!m.Deleted, m.Thread, m.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Commit transaction
	mock.ExpectCommit()

	// Delete the post
	err = m.Delete()
	assert.NoError(t, err, "No error should be returned")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestDeletePostDeleteInvalid(t *testing.T) {
	// Initialize invalid model
	m := &DeletePostModel{
		Thread: 0, // Invalid thread ID
		ID:     1,
		Ib:     1,
		Name:   "Test Thread",
	}

	// Delete the post
	err := m.Delete()
	assert.Error(t, err, "Error should be returned")
	assert.Equal(t, "DeletePostModel is not valid", err.Error(), "Error message should match expected value")
}

func TestDeletePostDeleteCountError(t *testing.T) {
	var err error

	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Initialize model with parameters
	m := &DeletePostModel{
		Thread:  1,
		ID:      1,
		Ib:      1,
		Name:    "Test Thread",
		Deleted: false,
	}

	// Begin transaction
	mock.ExpectBegin()

	// Expect query for post count - return an error
	expectedError := errors.New("count query error")
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM posts
		WHERE thread_id = \? AND post_deleted = 0`).
		WithArgs(m.Thread).
		WillReturnError(expectedError)

	// Rollback transaction
	mock.ExpectRollback()

	// Delete the post
	err = m.Delete()
	assert.Error(t, err, "Error should be returned")
	assert.Equal(t, expectedError, err, "Error should match the expected error")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestDeletePostDeleteBeginError(t *testing.T) {
	var err error

	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Initialize model with parameters
	m := &DeletePostModel{
		Thread:  1,
		ID:      1,
		Ib:      1,
		Name:    "Test Thread",
		Deleted: false,
	}

	// Begin transaction with error
	expectedError := errors.New("begin transaction error")
	mock.ExpectBegin().WillReturnError(expectedError)

	// Delete the post
	err = m.Delete()
	assert.Error(t, err, "Error should be returned")
	assert.Contains(t, err.Error(), expectedError.Error(), "Error should contain the expected error message")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestDeletePostDeleteThreadPrepareError(t *testing.T) {
	var err error

	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Initialize model with parameters
	m := &DeletePostModel{
		Thread:  1,
		ID:      1,
		Ib:      1,
		Name:    "Test Thread",
		Deleted: false,
	}

	// Begin transaction
	mock.ExpectBegin()

	// Expect query for post count - return only one post
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM posts
		WHERE thread_id = \? AND post_deleted = 0`).
		WithArgs(m.Thread).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).
			AddRow(1)) // Only one post in thread

	// Expect thread deletion prepare error
	expectedError := errors.New("thread prepare error")
	mock.ExpectPrepare(`UPDATE threads SET thread_deleted = 1
		WHERE thread_id = \? AND ib_id = \? LIMIT 1`).
		WillReturnError(expectedError)

	// Rollback transaction
	mock.ExpectRollback()

	// Delete the post
	err = m.Delete()
	assert.Error(t, err, "Error should be returned")
	assert.Equal(t, expectedError, err, "Error should match the expected error")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestDeletePostDeleteThreadExecError(t *testing.T) {
	var err error

	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Initialize model with parameters
	m := &DeletePostModel{
		Thread:  1,
		ID:      1,
		Ib:      1,
		Name:    "Test Thread",
		Deleted: false,
	}

	// Begin transaction
	mock.ExpectBegin()

	// Expect query for post count - return only one post
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM posts
		WHERE thread_id = \? AND post_deleted = 0`).
		WithArgs(m.Thread).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).
			AddRow(1)) // Only one post in thread

	// Expect thread deletion exec error
	expectedError := errors.New("thread exec error")
	mock.ExpectPrepare(`UPDATE threads SET thread_deleted = 1
		WHERE thread_id = \? AND ib_id = \? LIMIT 1`).
		ExpectExec().
		WithArgs(m.Thread, m.Ib).
		WillReturnError(expectedError)

	// Rollback transaction
	mock.ExpectRollback()

	// Delete the post
	err = m.Delete()
	assert.Error(t, err, "Error should be returned")
	assert.Equal(t, expectedError, err, "Error should match the expected error")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestDeletePostDeletePrepareError(t *testing.T) {
	var err error

	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Initialize model with parameters
	m := &DeletePostModel{
		Thread:  1,
		ID:      1,
		Ib:      1,
		Name:    "Test Thread",
		Deleted: false,
	}

	// Begin transaction
	mock.ExpectBegin()

	// Expect query for post count - return multiple posts
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM posts
		WHERE thread_id = \? AND post_deleted = 0`).
		WithArgs(m.Thread).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).
			AddRow(5)) // Multiple posts in thread

	// Delete prepare error
	expectedError := errors.New("prepare error")
	mock.ExpectPrepare(`UPDATE posts SET post_deleted = \?
		WHERE posts.thread_id = \? AND posts.post_num = \? LIMIT 1`).
		WillReturnError(expectedError)

	// Rollback transaction
	mock.ExpectRollback()

	// Delete the post
	err = m.Delete()
	assert.Error(t, err, "Error should be returned")
	assert.Contains(t, err.Error(), expectedError.Error(), "Error should contain the expected error message")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestDeletePostDeleteExecError(t *testing.T) {
	var err error

	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Initialize model with parameters
	m := &DeletePostModel{
		Thread:  1,
		ID:      1,
		Ib:      1,
		Name:    "Test Thread",
		Deleted: false,
	}

	// Begin transaction
	mock.ExpectBegin()

	// Expect query for post count - return multiple posts
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM posts
		WHERE thread_id = \? AND post_deleted = 0`).
		WithArgs(m.Thread).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).
			AddRow(5)) // Multiple posts in thread

	// Delete prepare and exec error
	expectedError := errors.New("exec error")
	mock.ExpectPrepare(`UPDATE posts SET post_deleted = \?
		WHERE posts.thread_id = \? AND posts.post_num = \? LIMIT 1`).
		ExpectExec().
		WithArgs(!m.Deleted, m.Thread, m.ID).
		WillReturnError(expectedError)

	// Rollback transaction
	mock.ExpectRollback()

	// Delete the post
	err = m.Delete()
	assert.Error(t, err, "Error should be returned")
	assert.Contains(t, err.Error(), expectedError.Error(), "Error should contain the expected error message")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestDeletePostUndelete(t *testing.T) {
	var err error

	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Initialize model with parameters
	m := &DeletePostModel{
		Thread:  1,
		ID:      1,
		Ib:      1,
		Name:    "Test Thread",
		Deleted: true, // Post is already deleted and will be undeleted
	}

	// Begin transaction
	mock.ExpectBegin()

	// We should not see any COUNT query since we're undeleting, not deleting

	// Delete prepare and exec - set post to undeleted
	mock.ExpectPrepare(`UPDATE posts SET post_deleted = \?
		WHERE posts.thread_id = \? AND posts.post_num = \? LIMIT 1`).
		ExpectExec().
		WithArgs(!m.Deleted, m.Thread, m.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Commit transaction
	mock.ExpectCommit()

	// Undelete the post
	err = m.Delete()
	assert.NoError(t, err, "No error should be returned")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

func TestDeletePostDeleteCommitError(t *testing.T) {
	var err error

	mock, err := db.NewTestDb()
	assert.NoError(t, err, "An error was not expected")
	defer db.CloseDb()

	// Initialize model with parameters
	m := &DeletePostModel{
		Thread:  1,
		ID:      1,
		Ib:      1,
		Name:    "Test Thread",
		Deleted: false,
	}

	// Begin transaction
	mock.ExpectBegin()

	// Expect query for post count - return multiple posts
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM posts
		WHERE thread_id = \? AND post_deleted = 0`).
		WithArgs(m.Thread).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).
			AddRow(5)) // Multiple posts in thread

	// Delete prepare and exec - set post to deleted
	mock.ExpectPrepare(`UPDATE posts SET post_deleted = \?
		WHERE posts.thread_id = \? AND posts.post_num = \? LIMIT 1`).
		ExpectExec().
		WithArgs(!m.Deleted, m.Thread, m.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Commit transaction with error
	expectedError := errors.New("commit error")
	mock.ExpectCommit().WillReturnError(expectedError)

	// Delete the post
	err = m.Delete()
	assert.Error(t, err, "Error should be returned")
	assert.Contains(t, err.Error(), expectedError.Error(), "Error should contain the expected error message")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}