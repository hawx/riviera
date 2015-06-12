package river

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetaEvents(t *testing.T) {
	evs := newMetaStore(3)

	evs.Set("thing", feeddata{Uri: "thinguri"})
	evs.Set("and", feeddata{Uri: "anduri"})

	assert.Equal(t, []Metadata{
		{Uri: "anduri", Log: []Event{}},
		{Uri: "thinguri", Log: []Event{}},
	}, evs.List())

	evs.Set("thing", feeddata{Uri: "no"})

	assert.Equal(t, []Metadata{
		{Uri: "anduri", Log: []Event{}},
		{Uri: "no", Log: []Event{}},
	}, evs.List())

	assert.Nil(t, evs.Log("and", Event{Code: 300}))
	assert.Nil(t, evs.Log("and", Event{Code: 200}))

	assert.Equal(t, []Metadata{
		{Uri: "anduri", Log: []Event{{Code: 200}, {Code: 300}}},
		{Uri: "no", Log: []Event{}},
	}, evs.List())

	assert.Nil(t, evs.Log("thing", Event{Code: 404}))

	assert.Equal(t, []Metadata{
		{Uri: "anduri", Log: []Event{{Code: 200}, {Code: 300}}},
		{Uri: "no", Log: []Event{{Code: 404}}},
	}, evs.List())

	evs.Delete("and")

	assert.Equal(t, []Metadata{
		{Uri: "no", Log: []Event{{Code: 404}}},
	}, evs.List())
}
