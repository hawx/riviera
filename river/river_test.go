package river

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"hawx.me/code/riviera/river/data/memdata"
	"hawx.me/code/riviera/river/models"
)

func TestRiver(t *testing.T) {
	db := memdata.Open()

	r := New(db, DefaultOptions)

	var buf bytes.Buffer
	r.WriteTo(&buf)

	var v models.River
	json.Unmarshal(buf.Bytes(), &v)

	assert := assert.New(t)

	assert.Equal(docsPath, v.Metadata.Docs)
	assert.WithinDuration(time.Now(), v.Metadata.WhenGMT.Time, time.Second)
	assert.WithinDuration(time.Now(), v.Metadata.WhenLocal.Time, time.Second)
	assert.Equal("3", v.Metadata.Version)
	assert.Equal(float64(0), v.Metadata.Secs)

	assert.Equal(models.Feeds{[]models.Feed{}}, v.UpdatedFeeds)
}
