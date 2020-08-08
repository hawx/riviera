package data

import (
	"database/sql"
	"log"

	"hawx.me/code/riviera/feed"
)

func (d *DB) Feed(uri string) (feed.Database, error) {
	return &feedDB{db: d.db, feedURL: uri}, nil
}

type feedDB struct {
	db      *sql.DB
	feedURL string
}

func (d *feedDB) Contains(key string) bool {
	tx, err := d.db.Begin()
	if err != nil {
		log.Println("contains:", err)
		return false
	}
	defer tx.Commit()

	var v int
	err = tx.QueryRow("SELECT 1 FROM keys WHERE Bucket = ? AND Key = ?", d.feedURL, key).Scan(&v)

	if err == sql.ErrNoRows {
		_, err = tx.Exec("INSERT INTO keys (Bucket, Key) VALUES (?, ?)", d.feedURL, key)
		if err != nil {
			log.Println("contains:", err)
		}
		return false
	}

	if err != nil {
		log.Println("sql contains:", err)
		return false
	}

	return true
}
