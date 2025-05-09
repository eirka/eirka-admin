package models

import (
	"database/sql"
	"errors"

	"github.com/eirka/eirka-libs/db"
	e "github.com/eirka/eirka-libs/errors"
)

// DeletePostModel holds request input
type DeletePostModel struct {
	Thread  uint
	ID      uint
	Ib      uint
	Name    string
	Deleted bool
}

// IsValid will check struct validity
func (m *DeletePostModel) IsValid() bool {

	if m.Thread == 0 {
		return false
	}

	if m.ID == 0 {
		return false
	}

	if m.Ib == 0 {
		return false
	}

	if m.Name == "" {
		return false
	}

	return true

}

// Status will return info
func (m *DeletePostModel) Status() (err error) {

	// Get Database handle
	dbase, err := db.GetDb()
	if err != nil {
		return
	}

	// get thread ib and title
	err = dbase.QueryRow(`SELECT thread_title, post_deleted FROM threads
	INNER JOIN posts on threads.thread_id = posts.thread_id
	WHERE threads.thread_id = ? AND ib_id = ? LIMIT 1`, m.Thread, m.Ib).Scan(&m.Name, &m.Deleted)
	if err == sql.ErrNoRows {
		return e.ErrNotFound
	} else if err != nil {
		return
	}

	return

}

// Delete will remove the entry
func (m *DeletePostModel) Delete() (err error) {

	// check model validity
	if !m.IsValid() {
		return errors.New("DeletePostModel is not valid")
	}

	// Get transaction handle
	tx, err := db.GetTransaction()
	if err != nil {
		return
	}
	defer tx.Rollback()

	// If we're deleting a post (not undeleting), check if this is the only non-deleted post in the thread
	if !m.Deleted {
		var postCount int
		err = tx.QueryRow(`SELECT COUNT(*) FROM posts
		WHERE thread_id = ? AND post_deleted = 0`, m.Thread).Scan(&postCount)
		if err != nil {
			return
		}

		// If this is the only non-deleted post in the thread, also mark the thread as deleted
		if postCount == 1 {
			ps2, err := tx.Prepare(`UPDATE threads SET thread_deleted = 1
			WHERE thread_id = ? AND ib_id = ? LIMIT 1`)
			if err != nil {
				return err
			}
			defer ps2.Close()

			_, err = ps2.Exec(m.Thread, m.Ib)
			if err != nil {
				return err
			}
		}
	}

	// set post to deleted
	ps1, err := tx.Prepare(`UPDATE posts SET post_deleted = ?
	WHERE posts.thread_id = ? AND posts.post_num = ? LIMIT 1`)
	if err != nil {
		return
	}
	defer ps1.Close()

	_, err = ps1.Exec(!m.Deleted, m.Thread, m.ID)
	if err != nil {
		return
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return
	}

	return

}
