package boltdata

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"hawx.me/code/riviera/river/data"
)

func TestBoltdata(t *testing.T) {
	dir, _ := ioutil.TempDir("", "riviera-bolt-test")
	defer os.RemoveAll(dir)

	assert := assert.New(t)

	var (
		path       = "/test.db"
		bucketName = []byte("what")
		key        = []byte("hey")
		value      = []byte("5")
	)

	db, err := Open(dir + path)
	assert.Nil(err)

	bucket, err := db.Bucket(bucketName)
	assert.Nil(err)

	bucket.Update(func(tx data.Tx) error {
		tx.Put(key, value)
		return nil
	})

	assert.Nil(db.Close())

	db, err = Open(dir + path)
	assert.Nil(err)

	bucket, err = db.Bucket(bucketName)
	assert.Nil(err)

	bucket.View(func(tx data.ReadTx) error {
		assert.Equal(tx.Get(key), value)
		return nil
	})
}
