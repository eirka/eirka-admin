package models

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/eirka/eirka-libs/db"
	e "github.com/eirka/eirka-libs/errors"

	local "github.com/eirka/eirka-admin/config"
	u "github.com/eirka/eirka-admin/utils"
)

type PurgePostModel struct {
	Thread uint
	Id     uint
	Ib     uint
	Name   string
}

// check struct validity
func (p *PurgePostModel) IsValid() bool {

	if p.Thread == 0 {
		return false
	}

	if p.Id == 0 {
		return false
	}

	if p.Id == 0 {
		return false
	}

	if p.Name == "" {
		return false
	}

	return true

}

type PostImage struct {
	Id    uint
	File  string
	Thumb string
}

// Status will return info
func (i *PurgePostModel) Status() (err error) {

	// Get Database handle
	dbase, err := db.GetDb()
	if err != nil {
		return
	}

	// get thread ib and title
	err = dbase.QueryRow("SELECT ib_id, thread_title FROM threads WHERE thread_id = ? LIMIT 1", i.Thread).Scan(&i.Ib, &i.Name)
	if err == sql.ErrNoRows {
		return e.ErrNotFound
	} else if err != nil {
		return
	}

	return

}

// Delete will remove the entry
func (i *PurgePostModel) Delete() (err error) {

	// check model validity
	if !i.IsValid() {
		return errors.New("PurgePostModel is not valid")
	}

	// Get transaction handle
	tx, err := db.GetTransaction()
	if err != nil {
		return
	}
	defer tx.Rollback()

	image := PostImage{}

	img := true

	// check if post has an image
	err = tx.QueryRow(`SELECT image_id,image_file,image_thumbnail FROM posts 
    INNER JOIN images on posts.post_id = images.post_id 
    WHERE posts.thread_id = ? AND posts.post_num = ? LIMIT 1`, i.Thread, i.Id).Scan(&image.Id, &image.File, &image.Thumb)
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

	_, err = ps1.Exec(i.Thread, i.Id)
	if err != nil {
		return
	}

	var lasttime *time.Time

	// get last post time
	err = tx.QueryRow("SELECT post_time FROM posts WHERE thread_id = ? ORDER BY post_id DESC LIMIT 1", i.Thread).Scan(&lasttime)
	if err != nil {
		return
	}

	// update last post time in thread
	ps2, err := tx.Prepare("UPDATE threads SET thread_last_post= ? WHERE thread_id= ?")
	if err != nil {
		return
	}
	defer ps2.Close()

	_, err = ps2.Exec(lasttime, i.Thread)
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
		u.DeleteS3(fmt.Sprintf("src/%s", image.File))
		if err != nil {
			return
		}

		u.DeleteS3(fmt.Sprintf("thumb/%s", image.Thumb))
		if err != nil {
			return
		}

		os.RemoveAll(filepath.Join(local.Settings.Directories.ImageDir, image.File))
		os.RemoveAll(filepath.Join(local.Settings.Directories.ThumbnailDir, image.Thumb))

	}

	return

}
