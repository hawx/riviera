package events

import "time"

// An Event keeps track of the results of fetching a feed.
type Event struct {
	At   time.Time `json:"at"`
	Uri  string    `json:"uri"`
	Code int       `json:"code"`
}

type Events struct {
	evs []Event
	cur int
	ln  int
	cp  int
}

func New(size int) *Events {
	return &Events{
		evs: make([]Event, size),
		cur: -1,
		ln:  0,
		cp:  size,
	}
}

func (e *Events) Prepend(ev Event) {
	e.cur = (e.cur + 1) % e.cp
	if e.ln < e.cp {
		e.ln++
	}

	e.evs[e.cp-e.cur-1] = ev
}

func (e *Events) List() []Event {
	if e.ln < e.cp {
		return e.evs[e.cp-e.ln:]
	}

	idx := e.cp - e.cur - 1
	return append(e.evs[idx:], e.evs[:idx]...)
}
