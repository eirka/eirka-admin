package models

import (
	"database/sql"
	"errors"

	"github.com/eirka/eirka-libs/db"
	e "github.com/eirka/eirka-libs/errors"
)

// DeleteThreadModel holds request input
type DeleteThreadModel struct {
	ID      uint
	Name    string
	Ib      uint
	Deleted bool
}

// IsValid will check struct validity
func (m *DeleteThreadModel) IsValid() bool {

	if m.ID == 0 {
		return false
	}

	if m.Name == "" {
		return false
	}

	if m.Ib == 0 {
		return false
	}

	return true

}

// Status will return info
func (m *DeleteThreadModel) Status() (err error) {

	// Get Database handle
	dbase, err := db.GetDb()
	if err != nil {
		return
	}

	// Check if favorite is already there
	err = dbase.QueryRow("SELECT thread_title, thread_deleted FROM threads WHERE thread_id = ? AND ib_id = ? LIMIT 1", m.ID, m.Ib).Scan(&m.Name, &m.Deleted)
	if err == sql.ErrNoRows {
		return e.ErrNotFound
	} else if err != nil {
		return
	}

	return

}

// Delete will remove the entry
func (m *DeleteThreadModel) Delete() (err error) {

	// check model validity
	if !m.IsValid() {
		return errors.New("DeleteThreadModel is not valid")
	}

	// Get Database handle
	dbase, err := db.GetDb()
	if err != nil {
		return
	}

	ps1, err := dbase.Prepare("UPDATE threads SET thread_deleted = ? WHERE thread_id = ? AND ib_id = ?")
	if err != nil {
		return
	}
	defer ps1.Close()

	_, err = ps1.Exec(!m.Deleted, m.ID, m.Ib)
	if err != nil {
		return
	}

	return

}
