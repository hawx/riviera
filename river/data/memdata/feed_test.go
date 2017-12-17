package memdata

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBucket(t *testing.T) {
	assert := assert.New(t)
	db := Open()

	bucket, err := db.(*database).Feed("test")
	assert.Nil(err)

	const key = "1"
	assert.False(bucket.Contains(key))
	assert.True(bucket.Contains(key))

	bucket2, err := db.(*database).Feed("test2")
	assert.Nil(err)

	assert.False(bucket2.Contains(key))
	assert.True(bucket2.Contains(key))
}
