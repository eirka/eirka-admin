package models

import (
	"database/sql"
	"errors"

	"github.com/eirka/eirka-libs/db"
	e "github.com/eirka/eirka-libs/errors"
)

type BanFileModel struct {
	Ib     uint
	Thread uint
	Id     uint
	User   uint
	Reason string
	Hash   string
}

// check struct validity
func (c *BanFileModel) IsValid() bool {

	if c.Ib == 0 {
		return false
	}

	if c.Thread == 0 {
		return false
	}

	if c.Id == 0 {
		return false
	}

	if c.User == 0 || c.User == 1 {
		return false
	}

	if c.Reason == "" {
		return false
	}

	if c.Hash == "" {
		return false
	}

	return true

}

// Status will return info
func (i *BanFileModel) Status() (err error) {

	// Get Database handle
	dbase, err := db.GetDb()
	if err != nil {
		return
	}

	// get thread ib and title
	err = dbase.QueryRow(`SELECT image_hash FROM threads
    INNER JOIN posts ON threads.thread_id = posts.thread_id
    INNER JOIN images ON posts.post_id = images.post_id
    WHERE ib_id = ? AND threads.thread_id = ? AND post_num = ? LIMIT 1`, i.Ib, i.Thread, i.Id).Scan(&i.Hash)
	if err == sql.ErrNoRows {
		return e.ErrNotFound
	} else if err != nil {
		return
	}

	return

}

// Toggle will add the ip to the ban list
func (i *BanFileModel) Post() (err error) {

	// check model validity
	if !i.IsValid() {
		return errors.New("BanFileModel is not valid")
	}

	// Get Database handle
	dbase, err := db.GetDb()
	if err != nil {
		return
	}

	_, err = dbase.Exec("INSERT IGNORE INTO banned_files (user_id,ib_id,ban_hash,ban_reason) VALUES (?,?,?,?)",
		i.User, i.Ib, i.Hash, i.Reason)
	if err != nil {
		return
	}

	return

}
