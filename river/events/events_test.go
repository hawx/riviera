package events

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEvents(t *testing.T) {
	evs := New(3)

	ev := []Event{
		{URI: "1"},
		{URI: "2"},
		{URI: "3"},
		{URI: "4"},
		{URI: "5"},
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
