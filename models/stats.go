package models

import (
	"github.com/eirka/eirka-libs/db"
	"time"
)

// NewModel holds the parameters from the request and also the key for the cache
type StatisticsModel struct {
	Ib     uint
	Result StatisticsType
}

type StatisticsType struct {
	Labels []time.Time `json:"labels"`
	Series []Series    `json:"series"`
}

type Series struct {
	Name string `json:"name"`
	Data []uint `json:"data"`
}

// Get will gather the information from the database and return it as JSON serialized data
func (i *StatisticsModel) Get() (err error) {

	// Initialize response header
	response := StatisticsType{}

	visitors := Series{
		Name: "Visitors",
	}

	hits := Series{
		Name: "Hits",
	}

	// Get Database handle
	dbase, err := db.GetDb()
	if err != nil {
		return
	}

	ps1, err := dbase.Prepare(`SELECT (now() - interval ? hour) as time, 
    COUNT(DISTINCT request_ip) as visitors, COUNT(request_itemkey) as hits 
    FROM analytics 
    WHERE request_time BETWEEN (now() - interval ? hour) AND (now() - interval ? hour) AND ib_id = ?`)
	if err != nil {
		return
	}
	defer ps1.Close()

	// loop through every two hours
	for hour := 24; hour >= 2; hour-- {
		if hour%2 == 0 {

			var label time.Time
			var visitor_count, hit_count uint

			// period minus two hours
			previous := (hour - 2)

			err := ps1.QueryRow(hour, hour, previous, i.Ib).Scan(&label, &visitor_count, &hit_count)
			if err != nil {
				return err
			}

			response.Labels = append(response.Labels, label)
			visitors.Data = append(visitors.Data, visitor_count)
			hits.Data = append(hits.Data, hit_count)

		}
	}

	response.Series = append(response.Series, visitors, hits)

	// This is the data we will serialize
	i.Result = response

	return

}
