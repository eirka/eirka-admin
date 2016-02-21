package models

import (
	"database/sql"
	"errors"

	"github.com/eirka/eirka-libs/db"
)

type BanIpModel struct {
	Ib     uint
	Thread uint
	Id     uint
	User   uint
	Reason string
	Ip     string
}

// check struct validity
func (c *BanIpModel) IsValid() bool {

	if c.Ib == 0 {
		return false
	}

	if c.User == 0 || c.User == 1 {
		return false
	}

	if c.Ip == "" {
		return false
	}

	return true

}

// Status will return info
func (i *BanIpModel) Status() (err error) {

	// Get Database handle
	dbase, err := db.GetDb()
	if err != nil {
		return
	}

	// get thread ib and title
	err = dbase.QueryRow(`SELECT post_ip FROM threads
    INNER JOIN posts ON threads.thread_id = posts.thread_id
    WHERE ib_id = ? AND threads.thread_id = ? AND post_num = ? LIMIT 1`, i.Ib, i.Thread, i.Id).Scan(&i.Ip)
	if err == sql.ErrNoRows {
		return e.ErrNotFound
	} else if err != nil {
		return
	}

	return

}

// Toggle will add the ip to the ban list
func (i *BanIpModel) Post() (err error) {

	// check model validity
	if !i.IsValid() {
		return errors.New("BanIpModel is not valid")
	}

	// Get Database handle
	dbase, err := db.GetDb()
	if err != nil {
		return
	}

	_, err = dbase.Exec("INSERT INTO banned_ips (user_id,ib_id,ban_ip,ban_reason) VALUES (?,?,?,?)",
		i.User, i.Ib, i.Ip, i.Reason)
	if err != nil {
		return
	}

	return

}
