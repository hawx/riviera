// Package data implements the data access for riviera.
package data

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	// register sqlite3 for database/sql
	_ "github.com/mattn/go-sqlite3"
	"hawx.me/code/riviera/feed"
	"hawx.me/code/riviera/river/confluence"
	"hawx.me/code/riviera/river/riverjs"
)

type DB struct {
	db *sql.DB
}

func Open(path string) (*DB, error) {
	sqlite, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	db := &DB{db: sqlite}

	return db, db.migrate()
}

func (d *DB) migrate() error {
	_, err := d.db.Exec(`
    CREATE TABLE IF NOT EXISTS keys (
      Bucket TEXT,
      Key TEXT,
      PRIMARY KEY (Bucket, Key)
    );

    CREATE TABLE IF NOT EXISTS gardenFeeds (
      FeedURL    TEXT PRIMARY KEY
    );

    CREATE TABLE IF NOT EXISTS feeds (
      FeedURL     TEXT PRIMARY KEY,
      WebsiteURL  TEXT,
      Title       TEXT,
      Description TEXT,
      UpdatedAt   DATETIME
    );
    CREATE UNIQUE INDEX IF NOT EXISTS idx_feeds_feedurl ON feeds (FeedURL);

    CREATE TABLE IF NOT EXISTS feedItems (
      Key       TEXT PRIMARY KEY,
      FeedURL   TEXT,
      PermaLink TEXT,
      PubDate   DATETIME,
      Title     TEXT,
      Link      TEXT,
      Body      TEXT,
      ID        TEXT,
      Comments  TEXT
    );

    CREATE TABLE IF NOT EXISTS enclosures (
      ID      INTEGER PRIMARY KEY AUTOINCREMENT,
      ItemKey TEXT,
      URL     TEXT,
      Type    TEXT,
      Length  INTEGER
    );

    CREATE TABLE IF NOT EXISTS thumbnails (
      ID      INTEGER PRIMARY KEY AUTOINCREMENT,
      ItemKey TEXT,
      URL     TEXT,
      Height  INTEGER,
      Width   INTEGER
    );

    CREATE TABLE IF NOT EXISTS riverFeeds (
      FeedURL     TEXT PRIMARY KEY
    );

    CREATE TABLE IF NOT EXISTS feedFetches (
      FeedURL   TEXT NOT NULL,
      FetchedAt DATETIME NOT NULL,
      Value     TEXT,
      PRIMARY KEY (FeedURL, FetchedAt)
    );
`)

	log.Println("migrated")
	return err
}

func (d *DB) Close() error {
	return d.db.Close()
}

func (d *DB) Read(uri string) (feed Feed, err error) {
	row := d.db.QueryRow("SELECT WebsiteURL, Title, Description, UpdatedAt FROM feeds WHERE FeedURL = ?",
		uri)

	feed.FeedURL = uri
	if err = row.Scan(&feed.WebsiteURL, &feed.Title, &feed.Description, &feed.UpdatedAt); err != nil {
		return feed, fmt.Errorf("scanning feed row: %w", err)
	}

	rows, err := d.db.Query("SELECT Key, PermaLink, PubDate, Title, Link, Body, ID, Comments FROM feedItems WHERE FeedURL = ?",
		uri)
	if err != nil {
		return feed, fmt.Errorf("selecting feedItems: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item FeedItem
		if err = rows.Scan(&item.Key, &item.PermaLink, &item.PubDate, &item.Title, &item.Link, &item.Body, &item.ID, &item.Comments); err != nil {
			return feed, fmt.Errorf("scanning feedItems row: %w", err)
		}

		// and enclosures
		// and thumbnails

		feed.Items = append(feed.Items, item)
	}

	if err = rows.Err(); err != nil {
		return feed, fmt.Errorf("rows err: %w", err)
	}

	return
}

func (d *DB) UpdateFeed(feed Feed) (err error) {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	_, err = tx.Exec(`REPLACE INTO feeds (WebsiteURL, Title, Description, UpdatedAt, FeedURL)
                                VALUES (?,          ?,     ?,           ?,         ?)`,
		feed.WebsiteURL,
		feed.Title,
		feed.Description,
		feed.UpdatedAt,
		feed.FeedURL)
	if err != nil {
		return err
	}

	for _, item := range feed.Items {
		_, err = tx.Exec(`INSERT INTO feedItems (Key, FeedURL, PermaLink, PubDate, Title, Link, Body, ID, Comments)
                                     VALUES (?,   ?,       ?,         ?,       ?,     ?,    ?,    ?,  ?)`,
			item.Key,
			feed.FeedURL,
			item.PermaLink,
			item.PubDate,
			item.Title,
			item.Link,
			item.Body,
			item.ID,
			item.Comments)
		if err != nil {
			return err
		}

		for _, enclosure := range item.Enclosures {
			_, err = tx.Exec(`INSERT INTO enclosures (ItemKey, URL, Type, Length)
                                        VALUES (?,       ?,   ?,    ?)`,
				item.Key,
				enclosure.URL,
				enclosure.Type,
				enclosure.Length)
			if err != nil {
				return err
			}
		}

		for _, thumbnail := range item.Thumbnails {
			_, err = tx.Exec(`INSERT INTO thumbnails (ItemKey, URL, Height, Width)
                                        VALUES (?,       ?,   ?,      ?)`,
				item.Key,
				thumbnail.URL,
				thumbnail.Height,
				thumbnail.Width)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (d *DB) Contains(key string) bool {
	var v int
	err := d.db.QueryRow("SELECT 1 FROM feedItems WHERE Key = ?", key).Scan(&v)

	if err != nil {
		if err != sql.ErrNoRows {
			log.Println("sql contains:", err)
		}
		return false
	}

	return true
}

func (d *DB) Feed(uri string) (feed.Database, error) {
	return &feedDB{db: d.db, feedURL: uri}, nil
}

type feedDB struct {
	db      *sql.DB
	feedURL string
}

func (d *feedDB) Contains(key string) bool {
	log.Println("contains", key)
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
