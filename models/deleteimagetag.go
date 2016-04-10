package models

import (
	"database/sql"
	"errors"

	"github.com/eirka/eirka-libs/db"
	e "github.com/eirka/eirka-libs/errors"
)

// DeleteImageTagModel holds request input
type DeleteImageTagModel struct {
	Image uint
	Tag   uint
	Name  string
	Ib    uint
}

// IsValid will check struct validity
func (m *DeleteImageTagModel) IsValid() bool {

	if m.Image == 0 {
		return false
	}

	if m.Tag == 0 {
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
func (m *DeleteImageTagModel) Status() (err error) {

	// Get Database handle
	dbase, err := db.GetDb()
	if err != nil {
		return
	}

	// Check if the tag is there
	err = dbase.QueryRow("SELECT tag_name FROM tags WHERE tag_id = ? AND ib_id = ? LIMIT 1", m.Tag, m.Ib).Scan(&m.Name)
	if err == sql.ErrNoRows {
		return e.ErrNotFound
	} else if err != nil {
		return
	}

	return

}

// Delete will remove the entry
func (m *DeleteImageTagModel) Delete() (err error) {

	// check model validity
	if !m.IsValid() {
		return errors.New("DeleteImageTagModel is not valid")
	}

	// Get Database handle
	dbase, err := db.GetDb()
	if err != nil {
		return
	}

	ps1, err := dbase.Prepare(`DELETE tm FROM tagmap AS tm
    INNER JOIN tags ON tm.tag_id = tags.tag_id
    WHERE image_id = ? AND tm.tag_id = ? AND ib_id = ?`)
	if err != nil {
		return
	}
	defer ps1.Close()

	_, err = ps1.Exec(m.Image, m.Tag, m.Ib)
	if err != nil {
		return
	}

	return

}
