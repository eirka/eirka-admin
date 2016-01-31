package models

import (
	"database/sql"
	"errors"

	"github.com/eirka/eirka-libs/db"
	e "github.com/eirka/eirka-libs/errors"
)

type CloseModel struct {
	Id     uint
	Name   string
	Ib     uint
	Closed bool
}

// check struct validity
func (c *CloseModel) IsValid() bool {

	if c.Id == 0 {
		return false
	}

	if c.Name == "" {
		return false
	}

	if c.Ib == 0 {
		return false
	}

	return true

}

// Status will return info
func (i *CloseModel) Status() (err error) {

	// Get Database handle
	dbase, err := db.GetDb()
	if err != nil {
		return
	}

	// Check if favorite is already there
	err = dbase.QueryRow("SELECT thread_title, thread_closed FROM threads WHERE thread_id = ? AND ib_id = ? LIMIT 1", i.Id, i.Ib).Scan(&i.Name, &i.Closed)
	if err == sql.ErrNoRows {
		return e.ErrNotFound
	} else if err != nil {
		return
	}

	return

}

// Toggle will change the thread status
func (i *CloseModel) Toggle() (err error) {

	// check model validity
	if !i.IsValid() {
		return errors.New("CloseModel is not valid")
	}

	// Get Database handle
	dbase, err := db.GetDb()
	if err != nil {
		return
	}

	ps1, err := dbase.Prepare("UPDATE threads SET thread_closed = ? WHERE thread_id = ? AND ib_id = ?")
	if err != nil {
		return
	}
	defer ps1.Close()

	_, err = ps1.Exec(!i.Closed, i.Id, i.Ib)
	if err != nil {
		return
	}

	return

}
