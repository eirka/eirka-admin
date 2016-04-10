package models

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/eirka/eirka-libs/amazon"
	"github.com/eirka/eirka-libs/db"
	e "github.com/eirka/eirka-libs/errors"

	local "github.com/eirka/eirka-admin/config"
)

// PurgePostModel holds request input
type PurgePostModel struct {
	Thread uint
	ID     uint
	Ib     uint
	Name   string
}

// IsValid will check struct validity
func (m *PurgePostModel) IsValid() bool {

	if m.Thread == 0 {
		return false
	}

	if m.ID == 0 {
		return false
	}

	if m.ID == 0 {
		return false
	}

	if m.Name == "" {
		return false
	}

	return true

}

type imageInfo struct {
	ID    uint
	File  string
	Thumb string
}

// Status will return info
func (m *PurgePostModel) Status() (err error) {

	// Get Database handle
	dbase, err := db.GetDb()
	if err != nil {
		return
	}

	// get thread ib and title
	err = dbase.QueryRow("SELECT thread_title FROM threads WHERE thread_id = ? AND ib_id = ? LIMIT 1", m.Thread, m.Ib).Scan(&m.Name)
	if err == sql.ErrNoRows {
		return e.ErrNotFound
	} else if err != nil {
		return
	}

	return

}

// Delete will remove the entry
func (m *PurgePostModel) Delete() (err error) {

	// check model validity
	if !m.IsValid() {
		return errors.New("PurgePostModel is not valid")
	}

	// Get transaction handle
	tx, err := db.GetTransaction()
	if err != nil {
		return
	}
	defer tx.Rollback()

	image := imageInfo{}

	img := true

	// check if post has an image
	err = tx.QueryRow(`SELECT image_id,image_file,image_thumbnail FROM posts
    INNER JOIN images on posts.post_id = images.post_id
    WHERE posts.thread_id = ? AND posts.post_num = ? LIMIT 1`, m.Thread, m.ID).Scan(&image.ID, &image.File, &image.Thumb)
	if err == sql.ErrNoRows {
		img = false
	} else if err != nil {
		return
	}

	// delete thread from database
	ps1, err := tx.Prepare("DELETE FROM posts WHERE thread_id= ? AND post_num = ? LIMIT 1")
	if err != nil {
		return
	}
	defer ps1.Close()

	_, err = ps1.Exec(m.Thread, m.ID)
	if err != nil {
		return
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return
	}

	// delete image file
	if img {

		// filename must exist to prevent deleting the directory ;D
		if image.Thumb == "" {
			return
		}

		if image.File == "" {
			return
		}

		// delete from amazon s3
		s3 := amazon.New()

		s3.Delete(fmt.Sprintf("src/%s", image.File))
		if err != nil {
			return
		}

		s3.Delete(fmt.Sprintf("thumb/%s", image.Thumb))
		if err != nil {
			return
		}

		os.RemoveAll(filepath.Join(local.Settings.Directories.ImageDir, image.File))
		os.RemoveAll(filepath.Join(local.Settings.Directories.ThumbnailDir, image.Thumb))

	}

	return

}
