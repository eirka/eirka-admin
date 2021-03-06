package models

import (
	"time"

	"github.com/eirka/eirka-libs/config"
	"github.com/eirka/eirka-libs/db"
	e "github.com/eirka/eirka-libs/errors"

	u "github.com/eirka/eirka-admin/utils"
)

// BoardLogModel holds request input
type BoardLogModel struct {
	Ib     uint
	Page   uint
	Result BoardLogType
}

// BoardLogType is the container for the JSON response
type BoardLogType struct {
	Body u.PagedResponse `json:"boardlog"`
}

// Log format for audit log entries
type Log struct {
	UID    uint       `json:"user_id"`
	Name   string     `json:"user_name"`
	Group  uint       `json:"user_group"`
	Time   *time.Time `json:"log_time"`
	Action string     `json:"log_action"`
	Meta   string     `json:"log_meta"`
}

// Get will gather the information from the database and return it as JSON serialized data
func (i *BoardLogModel) Get() (err error) {

	if i.Ib == 0 || i.Page == 0 {
		return e.ErrNotFound
	}

	// Initialize response header
	response := BoardLogType{}

	// to hold log entries
	entries := []Log{}

	// Initialize struct for pagination
	paged := u.PagedResponse{}
	// Set current page to parameter
	paged.CurrentPage = i.Page
	// Set threads per index page to config setting
	paged.PerPage = config.Settings.Limits.PostsPerPage

	// Get Database handle
	dbase, err := db.GetDb()
	if err != nil {
		return
	}

	// Get total tag count and put it in pagination struct
	err = dbase.QueryRow("SELECT count(*) FROM audit WHERE ib_id = ? AND audit_type = 1", i.Ib).Scan(&paged.Total)
	if err != nil {
		return
	}

	// Calculate Limit and total Pages
	paged.Get()

	// Return 404 if page requested is larger than actual pages
	if i.Page > paged.Pages {
		return e.ErrNotFound
	}

	// get image counts from tagmap
	rows, err := dbase.Query(`SELECT audit.user_id,user_name,
    COALESCE((SELECT MAX(role_id) FROM user_ib_role_map WHERE user_ib_role_map.user_id = users.user_id AND ib_id = ?),user_role_map.role_id) as role,
    audit_time,audit_action,audit_info FROM audit
    INNER JOIN users ON audit.user_id = users.user_id
    INNER JOIN user_role_map ON (user_role_map.user_id = users.user_id)
    WHERE ib_id = ? AND audit_type = 1
    ORDER BY audit_id DESC LIMIT ?,?`, i.Ib, i.Ib, paged.Limit, paged.PerPage)
	if err != nil {
		return
	}

	for rows.Next() {
		// Initialize posts struct
		entry := Log{}
		// Scan rows and place column into struct
		err := rows.Scan(&entry.UID, &entry.Name, &entry.Group, &entry.Time, &entry.Action, &entry.Meta)
		if err != nil {
			return err
		}

		// Append rows to info struct
		entries = append(entries, entry)
	}
	if rows.Err() != nil {
		return
	}

	// Add threads slice to items interface
	paged.Items = entries

	// Add pagedresponse to the response struct
	response.Body = paged

	// This is the data we will serialize
	i.Result = response

	return

}
