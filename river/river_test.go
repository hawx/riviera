package river

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"hawx.me/code/riviera/data"
	"hawx.me/code/riviera/river/riverjs"
)

func TestRiver(t *testing.T) {
	db, _ := data.Open("file:TestRiver?cache=shared&mode=memory")

	r := New(db, Options{})

	var buf bytes.Buffer
	r.Encode(&buf)

	var v riverjs.River
	json.Unmarshal(buf.Bytes(), &v)

	assert := assert.New(t)

	assert.Equal(docsPath, v.Metadata.Docs)
	assert.WithinDuration(time.Now(), v.Metadata.WhenGMT.Time, time.Second)
	assert.WithinDuration(time.Now(), v.Metadata.WhenLocal.Time, time.Second)
	assert.Equal("3", v.Metadata.Version)
	assert.Equal(float64(0), v.Metadata.Secs)

	assert.Equal(riverjs.Feeds{}, v.UpdatedFeeds)
}
