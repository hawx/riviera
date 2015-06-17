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

func Test2(t *testing.T) {
	dir, _ := ioutil.TempDir("", "riviera-bolt-test")
	defer os.RemoveAll(dir)

	var (
		path       = "/test2.db"
		bucketName = []byte("what")
	)

	db, _ := Open(dir + path)
	assert := assert.New(t)

	bk, err := db.Bucket(bucketName)
	assert.Nil(err)

	bk.Update(func(tx data.Tx) error {
		tx.Put([]byte("4"), []byte("d"))
		tx.Put([]byte("1"), []byte("a"))
		tx.Put([]byte("3"), []byte("c"))
		tx.Put([]byte("5"), []byte("e"))
		tx.Put([]byte("2"), []byte("b"))
		return nil
	})

	bk.View(func(tx data.ReadTx) error {
		assert.Equal([]byte("a"), tx.Get([]byte("1")))

		afters := []string{"c", "d", "e"}
		for i, v := range tx.After([]byte("3")) {
			assert.Equal([]byte(afters[i]), v)
		}
		assert.Empty(tx.After([]byte("6")))

		befores := []string{"1", "2"}
		for i, k := range tx.KeysBefore([]byte("3")) {
			assert.Equal([]byte(befores[i]), k)
		}
		assert.Empty(tx.KeysBefore([]byte("-1")))

		alls := []string{"a", "b", "c", "d", "e"}
		for i, v := range tx.All() {
			assert.Equal([]byte(alls[i]), v)
		}

		return nil
	})

	bk.Update(func(tx data.Tx) error {
		tx.Delete([]byte("a"))
		return nil
	})

	bk.View(func(tx data.ReadTx) error {
		assert.Nil(tx.Get([]byte("a")))
		return nil
	})

	assert.Nil(db.Close())

	db2, _ := Open(dir + path)
	bk2, _ := db2.Bucket(bucketName)

	bk2.View(func(tx data.ReadTx) error {
		afters := []string{"c", "d", "e"}
		for i, v := range tx.After([]byte("3")) {
			assert.Equal([]byte(afters[i]), v)
		}

		befores := []string{"1", "2"}
		for i, v := range tx.KeysBefore([]byte("3")) {
			assert.Equal([]byte(befores[i]), v)
		}

		alls := []string{"a", "b", "c", "d", "e"}
		for i, v := range tx.All() {
			assert.Equal([]byte(alls[i]), v)
		}

		return nil
	})
}
