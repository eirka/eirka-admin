package models

import (
	"database/sql"
	"errors"

	"github.com/eirka/eirka-libs/db"
	e "github.com/eirka/eirka-libs/errors"
)

// BanFileModel holds request input
type BanFileModel struct {
	Ib     uint
	Thread uint
	ID     uint
	User   uint
	Reason string
	Hash   string
}

// IsValid will check struct validity
func (m *BanFileModel) IsValid() bool {

	if m.Ib == 0 {
		return false
	}

	if m.Thread == 0 {
		return false
	}

	if m.ID == 0 {
		return false
	}

	if m.User == 0 || m.User == 1 {
		return false
	}

	if m.Reason == "" {
		return false
	}

	if m.Hash == "" {
		return false
	}

	return true

}

// Status will return info
func (m *BanFileModel) Status() (err error) {

	// Get Database handle
	dbase, err := db.GetDb()
	if err != nil {
		return
	}

	// get thread ib and title
	err = dbase.QueryRow(`SELECT image_hash FROM threads
    INNER JOIN posts ON threads.thread_id = posts.thread_id
    INNER JOIN images ON posts.post_id = images.post_id
    WHERE ib_id = ? AND threads.thread_id = ? AND post_num = ? LIMIT 1`, m.Ib, m.Thread, m.ID).Scan(&m.Hash)
	if err == sql.ErrNoRows {
		return e.ErrNotFound
	} else if err != nil {
		return
	}

	return

}

// Post will add the file to the table
func (m *BanFileModel) Post() (err error) {

	// check model validity
	if !m.IsValid() {
		return errors.New("BanFileModel is not valid")
	}

	// Get Database handle
	dbase, err := db.GetDb()
	if err != nil {
		return
	}

	_, err = dbase.Exec("INSERT IGNORE INTO banned_files (user_id,ib_id,ban_hash,ban_reason) VALUES (?,?,?,?)",
		m.User, m.Ib, m.Hash, m.Reason)
	if err != nil {
		return
	}

	return

}
