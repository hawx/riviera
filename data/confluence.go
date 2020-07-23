package data

import (
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"hawx.me/code/riviera/river/confluence"
	"hawx.me/code/riviera/river/riverjs"
)

func (d *DB) Confluence() confluence.Database {
	return &confluenceDB{db: d.db}
}

type confluenceDB struct {
	db *sql.DB
}

func (d *confluenceDB) Add(feed riverjs.Feed) {
	data, _ := json.Marshal(feed)

	d.db.Exec("INSERT INTO feedFetches (FeedURL, FetchedAt, Value) VALUES (?, ?, ?)",
		feed.FeedURL,
		feed.WhenLastUpdate.Add(0),
		string(data))
}

func (d *confluenceDB) Truncate(cutoff time.Duration) {
	d.db.Exec("DELETE FROM feedFetches WHERE FetchedAt < ?",
		time.Now().Add(cutoff))
}

func (d *confluenceDB) Latest(cutoff time.Duration) (feeds []riverjs.Feed) {
	rows, err := d.db.Query("SELECT Value FROM feedFetches WHERE FetchedAt > ? ORDER BY FetchedAt DESC, FeedURL DESC",
		time.Now().Add(cutoff))
	if err != nil {
		log.Println("latest:", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var data string
		rows.Scan(&data)

		var feed riverjs.Feed
		json.Unmarshal([]byte(data), &feed)

		feeds = append(feeds, feed)
	}

	return feeds
}
