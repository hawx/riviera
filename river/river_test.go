package river

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"hawx.me/code/riviera/data/memdata"
	"hawx.me/code/riviera/river/models"
	"hawx.me/code/riviera/subscriptions"
)

type testSubs struct{}

func (s testSubs) List() []subscriptions.Subscription {
	return []subscriptions.Subscription{}
}

func (s testSubs) Refresh(sub subscriptions.Subscription) {

}

func (s testSubs) OnAdd(f func(subscriptions.Subscription)) {

}

func (s testSubs) OnRemove(f func(string)) {

}

func TestRiver(t *testing.T) {
	db := memdata.Open()
	subs := testSubs{}

	r := New(db, subs, DefaultOptions)

	var buf bytes.Buffer
	r.WriteTo(&buf)

	var v models.River
	json.Unmarshal(buf.Bytes(), &v)

	assert := assert.New(t)

	assert.Equal(DOCS, v.Metadata.Docs)
	assert.WithinDuration(time.Now(), v.Metadata.WhenGMT.Time, time.Second)
	assert.WithinDuration(time.Now(), v.Metadata.WhenLocal.Time, time.Second)
	assert.Equal("3", v.Metadata.Version)
	assert.Equal(0, v.Metadata.Secs)

	assert.Equal(models.Feeds{[]models.Feed{}}, v.UpdatedFeeds)
}
