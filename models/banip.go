package models

import (
	"database/sql"
	"errors"

	"github.com/eirka/eirka-libs/db"
	e "github.com/eirka/eirka-libs/errors"
)

type BanIpModel struct {
	Ib     uint
	User   uint
	Ip     string
	Reason string
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

	_, err = ps1.Exec("INSERT INTO banned_ips (user_id,ib_id,ban_ip,ban_reason) VALUES (?,?,?,?)",
		i.User, i.Ib, i.Ip, i.Reason)
	if err != nil {
		return
	}

	return

}
