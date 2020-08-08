package data

import "database/sql"

type subscriptionsDB struct {
	db   *sql.DB
	name string
}

func (d *DB) Subscriptions(name string) *subscriptionsDB {
	return &subscriptionsDB{db: d.db, name: name}
}

func (d *subscriptionsDB) Add(uri string) error {
	_, err := d.db.Exec("INSERT OR IGNORE INTO subscriptions (FeedURL, Name) VALUES (?, ?)",
		uri,
		d.name)

	return err
}

func (d *subscriptionsDB) Remove(uri string) error {
	_, err := d.db.Exec("DELETE FROM subscriptions WHERE FeedURL = ? AND Name = ?",
		uri,
		d.name)

	return err
}

func (d *subscriptionsDB) List() (list []string, err error) {
	rows, err := d.db.Query("SELECT FeedURL FROM subscriptions WHERE Name = ?",
		d.name)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var uri string
		if err = rows.Scan(&uri); err != nil {
			return
		}
		list = append(list, uri)
	}

	err = rows.Err()
	return
}
