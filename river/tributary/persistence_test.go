package tributary

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"hawx.me/code/riviera/river/data/memdata"
)

func TestBucket(t *testing.T) {
	assert := assert.New(t)
	db := memdata.Open()

	bucket, err := NewBucket(db, "test")
	assert.Nil(err)

	const key = "1"
	assert.False(bucket.Contains(key))
	assert.True(bucket.Contains(key))

	bucket2, err := NewBucket(db, "test2")
	assert.Nil(err)

	assert.False(bucket2.Contains(key))
	assert.True(bucket2.Contains(key))
}
