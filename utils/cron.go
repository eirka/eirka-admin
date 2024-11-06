package utils

import (
	"github.com/robfig/cron"

	"github.com/eirka/eirka-libs/db"
)

func init() {

	c := cron.New()

	// prune old analytics
	err := c.AddFunc("@midnight", PruneAnalytics)
	if err != nil {
		panic("Could not add prune analytics cron job")
	}

	c.Start()

}

// PruneAnalytics will remove old entries from the analytics table
func PruneAnalytics() {

	// Get Database handle
	dbase, err := db.GetDb()
	if err != nil {
		return
	}

	_, err = dbase.Exec("DELETE FROM analytics WHERE request_time < (now() - interval 1 month)")
	if err != nil {
		return
	}

}
