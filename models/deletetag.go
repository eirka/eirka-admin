package models

import (
	"database/sql"
	"errors"

	"github.com/eirka/eirka-libs/db"
	e "github.com/eirka/eirka-libs/errors"
)

// DeleteTagModel holds request input
type DeleteTagModel struct {
	ID   uint
	Name string
	Ib   uint
}

// IsValid will check struct validity
func (m *DeleteTagModel) IsValid() bool {

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
func (m *DeleteTagModel) Status() (err error) {

	// Get Database handle
	dbase, err := db.GetDb()
	if err != nil {
		return
	}

	// Check if favorite is already there
	err = dbase.QueryRow("SELECT tag_name FROM tags WHERE tag_id = ? AND ib_id = ? LIMIT 1", m.ID, m.Ib).Scan(&m.Name)
	if err == sql.ErrNoRows {
		return e.ErrNotFound
	} else if err != nil {
		return
	}

	return

}

// Delete will remove the entry
func (m *DeleteTagModel) Delete() (err error) {

	// check model validity
	if !m.IsValid() {
		return errors.New("DeleteTagModel is not valid")
	}

	// Get Database handle
	dbase, err := db.GetDb()
	if err != nil {
		return
	}

	ps1, err := dbase.Prepare("DELETE FROM tags WHERE tag_id= ? AND ib_id = ? LIMIT 1")
	if err != nil {
		return
	}
	defer ps1.Close()

	_, err = ps1.Exec(m.ID, m.Ib)
	if err != nil {
		return
	}

	return

}
