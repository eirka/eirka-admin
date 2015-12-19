package models

import (
	"github.com/eirka/eirka-libs/config"
	"github.com/eirka/eirka-libs/db"
	e "github.com/eirka/eirka-libs/errors"
	"github.com/eirka/eirka-libs/validate"
)

type UpdateTagModel struct {
	Id      uint
	Ib      uint
	Tag     string
	TagType uint
}

func (i *UpdateTagModel) ValidateInput() (err error) {
	if i.Ib == 0 {
		return e.ErrInvalidParam
	}

	if i.TagType == 0 {
		return e.ErrInvalidParam
	}

	// Validate name input
	tag := validate.Validate{Input: i.Tag, Max: config.Settings.Limits.TagMaxLength, Min: config.Settings.Limits.TagMinLength}
	if tag.IsEmpty() {
		return e.ErrNoTagName
	} else if tag.MinLength() {
		return e.ErrTagShort
	} else if tag.MaxLength() {
		return e.ErrTagLong
	}

	return

}

// Status will return info
func (i *UpdateTagModel) Status() (err error) {

	// Get Database handle
	dbase, err := db.GetDb()
	if err != nil {
		return
	}

	var dupe bool

	// check if there is already a tag
	err = dbase.QueryRow("select count(*) from tags where tag_name = ? AND ib_id = ? AND NOT tag_id = ?", i.Tag, i.Ib, i.Id).Scan(&dupe)
	if err != nil {
		return
	}

	if dupe {
		return e.ErrDuplicateTag
	}

	return

}

// Update will update the entry
func (i *UpdateTagModel) Update() (err error) {

	// Get Database handle
	dbase, err := db.GetDb()
	if err != nil {
		return
	}

	ps1, err := dbase.Prepare("UPDATE tags SET tag_name= ?, tagtype_id= ? WHERE tag_id = ?")
	if err != nil {
		return
	}
	defer ps1.Close()

	_, err = ps1.Exec(i.Tag, i.TagType, i.Id)
	if err != nil {
		return
	}

	return

}
