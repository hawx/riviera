package data

import "database/sql"

type subscriptionsDB struct {
	db   *sql.DB
	name string

	onAdd    func(uri string)
	onRemove func(uri string)
}

func (d *DB) Subscriptions(name string) *subscriptionsDB {
	return &subscriptionsDB{db: d.db, name: name}
}

func (d *subscriptionsDB) OnAdd(f func(uri string)) {
	d.onAdd = f
}

func (d *subscriptionsDB) OnRemove(f func(uri string)) {
	d.onRemove = f
}

func (d *subscriptionsDB) Add(uri string) error {
	_, err := d.db.Exec("INSERT INTO subscriptions (FeedURL, Name) VALUES (?, ?)",
		uri,
		d.name)
	if err == nil && d.onAdd != nil {
		d.onAdd(uri)
	}

	return err
}

func (d *subscriptionsDB) Remove(uri string) error {
	_, err := d.db.Exec("DELETE FROM subscriptions WHERE FeedURL = ? AND Name = ?",
		uri,
		d.name)
	if err == nil && d.onRemove != nil {
		d.onRemove(uri)
	}

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
