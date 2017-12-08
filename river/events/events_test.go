package events

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEvents(t *testing.T) {
	evs := newEvents(3)

	ev := []Event{
		{Uri: "1"},
		{Uri: "2"},
		{Uri: "3"},
		{Uri: "4"},
		{Uri: "5"},
	}

	assert.Equal(t, []Event{}, evs.List())

	evs.Prepend(ev[0])
	assert.Equal(t, []Event{ev[0]}, evs.List())

	evs.Prepend(ev[1])
	assert.Equal(t, []Event{ev[1], ev[0]}, evs.List())

	evs.Prepend(ev[2])
	assert.Equal(t, []Event{ev[2], ev[1], ev[0]}, evs.List())

	evs.Prepend(ev[3])
	assert.Equal(t, []Event{ev[3], ev[2], ev[1]}, evs.List())
}
