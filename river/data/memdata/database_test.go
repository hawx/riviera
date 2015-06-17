package memdata

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"hawx.me/code/riviera/river/data"
)

func TestMemData(t *testing.T) {
	db := Open()
	assert := assert.New(t)

	var (
		bucketName = []byte("cool")
	)

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
}
