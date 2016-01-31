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

type PurgeThreadModel struct {
	Id   uint
	Name string
	Ib   uint
}

// check struct validity
func (p *PurgeThreadModel) IsValid() bool {

	if p.Id == 0 {
		return false
	}

	if p.Name == "" {
		return false
	}

	if p.Ib == 0 {
		return false
	}

	return true

}

type ThreadImages struct {
	Id    uint
	File  string
	Thumb string
}

// Status will return info
func (i *PurgeThreadModel) Status() (err error) {

	// Get Database handle
	dbase, err := db.GetDb()
	if err != nil {
		return
	}

	// Check if favorite is already there
	err = dbase.QueryRow("SELECT thread_title FROM threads WHERE thread_id = ? AND ib_id = ? LIMIT 1", i.Id, i.Ib).Scan(&i.Name)
	if err == sql.ErrNoRows {
		return e.ErrNotFound
	} else if err != nil {
		return
	}

	return

}

// Delete will remove the entry
func (i *PurgeThreadModel) Delete() (err error) {

	// check model validity
	if !i.IsValid() {
		return errors.New("PurgeThreadModel is not valid")
	}

	// Get Database handle
	dbase, err := db.GetDb()
	if err != nil {
		return
	}

	images := []ThreadImages{}

	// Get thread images
	rows, err := dbase.Query(`SELECT image_id,image_file,image_thumbnail FROM images
    INNER JOIN posts on images.post_id = posts.post_id
    INNER JOIN threads on threads.thread_id = posts.thread_id
    WHERE threads.thread_id = ? AND ib_id = ?`, i.Id, i.Ib)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		image := ThreadImages{}

		err := rows.Scan(&image.Id, &image.File, &image.Thumb)
		if err != nil {
			return err
		}
		// Append rows to info struct
		images = append(images, image)
	}
	err = rows.Err()
	if err != nil {
		return
	}

	// delete thread from database
	ps1, err := dbase.Prepare("DELETE FROM threads WHERE thread_id= ? AND ib_id = ? LIMIT 1")
	if err != nil {
		return
	}
	defer ps1.Close()

	_, err = ps1.Exec(i.Id, i.Ib)
	if err != nil {
		return
	}

	// delete image files
	go func() {

		for _, image := range images {

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

	}()

	return

}
