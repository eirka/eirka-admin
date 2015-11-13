package models

import (
	"database/sql"

	"github.com/techjanitor/pram-libs/db"
	e "github.com/techjanitor/pram-libs/errors"
)

type StickyModel struct {
	Id     uint
	Name   string
	Ib     uint
	Sticky bool
}

// Status will return info
func (i *StickyModel) Status() (err error) {

	// Get Database handle
	dbase, err := db.GetDb()
	if err != nil {
		return
	}

	// Check if favorite is already there
	err = dbase.QueryRow("SELECT ib_id, thread_title, thread_sticky FROM threads WHERE thread_id = ? LIMIT 1", i.Id).Scan(&i.Ib, &i.Name, &i.Sticky)
	if err == sql.ErrNoRows {
		return e.ErrNotFound
	} else if err != nil {
		return
	}

	return

}

// Toggle will change the thread status
func (i *StickyModel) Toggle() (err error) {

	// Get Database handle
	dbase, err := db.GetDb()
	if err != nil {
		return
	}

	ps1, err := dbase.Prepare("UPDATE threads SET thread_sticky = ? WHERE thread_id = ?")
	if err != nil {
		return
	}
	defer ps1.Close()

	_, err = ps1.Exec(!i.Sticky, i.Id)
	if err != nil {
		return
	}

	return

}