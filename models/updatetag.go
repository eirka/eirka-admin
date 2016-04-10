package models

import (
	"errors"
	"html"

	"github.com/microcosm-cc/bluemonday"

	"github.com/eirka/eirka-libs/config"
	"github.com/eirka/eirka-libs/db"
	e "github.com/eirka/eirka-libs/errors"
	"github.com/eirka/eirka-libs/validate"
)

// UpdateTagModel holds request input
type UpdateTagModel struct {
	ID      uint
	Ib      uint
	Tag     string
	TagType uint
}

// IsValid will check struct validity
func (m *UpdateTagModel) IsValid() bool {

	if m.ID == 0 {
		return false
	}

	if m.Ib == 0 {
		return false
	}

	if m.Tag == "" {
		return false
	}

	if m.TagType == 0 {
		return false
	}

	return true

}

// ValidateInput checks the data input for correctness
func (m *UpdateTagModel) ValidateInput() (err error) {
	if m.Ib == 0 {
		return e.ErrInvalidParam
	}

	if m.TagType == 0 {
		return e.ErrInvalidParam
	}

	// Initialize bluemonday
	p := bluemonday.StrictPolicy()

	// sanitize for html and xss
	m.Tag = html.UnescapeString(p.Sanitize(m.Tag))

	// Validate name input
	tag := validate.Validate{Input: m.Tag, Max: config.Settings.Limits.TagMaxLength, Min: config.Settings.Limits.TagMinLength}
	if tag.IsEmpty() {
		return e.ErrNoTagName
	} else if tag.MinPartsLength() {
		return e.ErrTagShort
	} else if tag.MaxLength() {
		return e.ErrTagLong
	}

	return

}

// Status will return info
func (m *UpdateTagModel) Status() (err error) {

	// Get Database handle
	dbase, err := db.GetDb()
	if err != nil {
		return
	}

	var dupe bool

	// check if there is already a tag
	err = dbase.QueryRow("select count(*) from tags where tag_name = ? AND ib_id = ? AND NOT tag_id = ?", m.Tag, m.Ib, m.ID).Scan(&dupe)
	if err != nil {
		return
	}

	if dupe {
		return e.ErrDuplicateTag
	}

	return

}

// Update will update the entry
func (m *UpdateTagModel) Update() (err error) {

	// check model validity
	if !m.IsValid() {
		return errors.New("UpdateTagModel is not valid")
	}

	// Get Database handle
	dbase, err := db.GetDb()
	if err != nil {
		return
	}

	ps1, err := dbase.Prepare("UPDATE tags SET tag_name= ?, tagtype_id= ? WHERE tag_id = ? AND ib_id = ?")
	if err != nil {
		return
	}
	defer ps1.Close()

	_, err = ps1.Exec(m.Tag, m.TagType, m.ID, m.Ib)
	if err != nil {
		return
	}

	return

}
