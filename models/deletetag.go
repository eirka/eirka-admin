package models

import (
	"database/sql"
	"errors"

	"github.com/eirka/eirka-libs/db"
	e "github.com/eirka/eirka-libs/errors"
)

type DeleteTagModel struct {
	Id   uint
	Name string
	Ib   uint
}

// check struct validity
func (d *DeleteTagModel) IsValid() bool {

	if d.Id == 0 {
		return false
	}

	if d.Name == "" {
		return false
	}

	if d.Ib == 0 {
		return false
	}

	return true

}

// Status will return info
func (i *DeleteTagModel) Status() (err error) {

	// Get Database handle
	dbase, err := db.GetDb()
	if err != nil {
		return
	}

	// Check if favorite is already there
	err = dbase.QueryRow("SELECT tag_name FROM tags WHERE tag_id = ? AND ib_id = ? LIMIT 1", i.Id, i.Ib).Scan(&i.Name)
	if err == sql.ErrNoRows {
		return e.ErrNotFound
	} else if err != nil {
		return
	}

	return

}

// Delete will remove the entry
func (i *DeleteTagModel) Delete() (err error) {

	// check model validity
	if !i.IsValid() {
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

	_, err = ps1.Exec(i.Id, i.Ib)
	if err != nil {
		return
	}

	return

}
