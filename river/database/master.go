package database

import (
	"github.com/boltdb/bolt"

	"fmt"
)

type Master interface {
	River() River
	Bucket(string) Bucket
	Close()
}

type master struct {
	db *bolt.DB
}

func Open(path string) (Master, error) {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	return &master{db}, nil
}

func (m *master) River() River {
	m.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte(riverBucket))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	return &river{m.db}
}

func (m *master) Bucket(name string) Bucket {
	m.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte(name))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	return &bucket{name, m.db}
}

func (m *master) Close() {
	m.db.Close()
}
