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
	Visitors uint        `json:"visitors"`
	Hits     uint        `json:"hits"`
	Threads  uint        `json:"threads"`
	Posts    uint        `json:"posts"`
	Images   uint        `json:"images"`
	Labels   []time.Time `json:"labels"`
	Series   []Series    `json:"series"`
}

type Series struct {
	Name string `json:"name"`
	Data []uint `json:"data"`
}

// Get will gather the information from the database and return it as JSON serialized data
func (i *StatisticsModel) Get() (err error) {

	// Initialize response header
	response := StatisticsType{}

	// holds visitors info
	visitors := Series{
		Name: "Visitors",
	}

	// holds count of hits
	hits := Series{
		Name: "Hits",
	}

	// Get Database handle
	dbase, err := db.GetDb()
	if err != nil {
		return
	}

	// get board stats
	err = dbase.QueryRow(`SELECT (SELECT COUNT(thread_id)
    FROM threads
    WHERE threads.ib_id=imageboards.ib_id AND thread_deleted != 1) AS thread_count,
    (SELECT COUNT(post_id)
    FROM threads
    LEFT JOIN posts ON posts.thread_id = threads.thread_id
    WHERE threads.ib_id=imageboards.ib_id AND post_deleted != 1) AS post_count,
    (SELECT COUNT(image_id)
    FROM threads
    LEFT JOIN posts ON posts.thread_id = threads.thread_id
    LEFT JOIN images ON images.post_id = posts.post_id
    WHERE threads.ib_id=imageboards.ib_id AND post_deleted != 1) AS image_count
    FROM imageboards WHERE ib_id = ?`, i.Ib).Scan(&response.Threads, &response.Posts, &response.Images)
	if err != nil {
		return
	}

	// get visitor stats
	err = dbase.QueryRow(`SELECT COUNT(DISTINCT request_ip) as visitors, COUNT(request_itemkey) as hits 
    FROM analytics 
    WHERE request_time BETWEEN (now() - interval 1 day) AND now() AND ib_id = ?`, i.Ib).Scan(&response.Visitors, &response.Hits)
	if err != nil {
		return
	}

	// get visitor period stats for chart
	ps1, err := dbase.Prepare(`SELECT (now() - interval ? hour) as time, 
    COUNT(DISTINCT request_ip) as visitors, COUNT(request_itemkey) as hits 
    FROM analytics 
    WHERE request_time BETWEEN (now() - interval ? hour) AND (now() - interval ? hour) AND ib_id = ?`)
	if err != nil {
		return
	}
	defer ps1.Close()

	// loop through every four hours
	for hour := 24; hour >= 4; hour-- {
		if hour%4 == 0 {

			var label time.Time
			var visitor_count, hit_count uint

			// period minus two hours
			previous := (hour - 4)

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
