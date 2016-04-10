package models

import (
	"time"

	"github.com/eirka/eirka-libs/db"
)

// StatisticsModel holds request input
type StatisticsModel struct {
	Ib     uint
	Result StatisticsType
}

// StatisticsType holds the analytics metadata
type StatisticsType struct {
	Visitors uint        `json:"visitors"`
	Hits     uint        `json:"hits"`
	Threads  uint        `json:"threads"`
	Posts    uint        `json:"posts"`
	Images   uint        `json:"images"`
	Labels   []time.Time `json:"labels"`
	Series   []Series    `json:"series"`
}

// Series holds the analytics data
type Series struct {
	Name string `json:"name"`
	Data []uint `json:"data"`
}

// Get will gather the information from the database and return it as JSON serialized data
func (m *StatisticsModel) Get() (err error) {

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
    FROM imageboards WHERE ib_id = ?`, m.Ib).Scan(&response.Threads, &response.Posts, &response.Images)
	if err != nil {
		return
	}

	// get visitor stats
	err = dbase.QueryRow(`SELECT COUNT(DISTINCT request_ip) as visitors, COUNT(request_itemkey) as hits
    FROM analytics
    WHERE request_time BETWEEN (now() - interval 1 day) AND now() AND ib_id = ?`, m.Ib).Scan(&response.Visitors, &response.Hits)
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
			var visitorCount, hitCount uint

			// period minus two hours
			previous := (hour - 4)

			err := ps1.QueryRow(hour, hour, previous, m.Ib).Scan(&label, &visitorCount, &hitCount)
			if err != nil {
				return err
			}

			response.Labels = append(response.Labels, label)
			visitors.Data = append(visitors.Data, visitorCount)
			hits.Data = append(hits.Data, hitCount)

		}
	}

	response.Series = append(response.Series, visitors, hits)

	// This is the data we will serialize
	m.Result = response

	return

}
