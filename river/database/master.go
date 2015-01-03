package database

import (
	"github.com/hawx/riviera/data"
)

type Master interface {
	River() River
	Bucket(string) Bucket
	Close()
}

type master struct {
	db data.Database
}

func Open(path string) (Master, error) {
	db, err := data.Open(path)
	if err != nil {
		return nil, err
	}

	return &master{db}, nil
}

var riverBucket = []byte("river")

func (m *master) River() River {
	b, err := m.db.Bucket(riverBucket)
	if err != nil {
		panic(err)
	}

	return &river{b}
}

func (m *master) Bucket(name string) Bucket {
	b, err := m.db.Bucket([]byte(name))
	if err != nil {
		panic(err)
	}

	return &bucket{b}
}

func (m *master) Close() {
	m.db.Close()
}
