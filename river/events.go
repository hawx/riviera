package river

import "time"

type Event struct {
	At   time.Time `json:"at"`
	Uri  string    `json:"uri"`
	Code int       `json:"code"`
}

type events struct {
	evs []Event
	cur int
	ln  int
	cp  int
}

func newEvents(size int) *events {
	return &events{
		evs: make([]Event, size),
		cur: -1,
		ln:  0,
		cp:  size,
	}
}

func (e *events) Prepend(ev Event) {
	e.cur = (e.cur + 1) % e.cp
	if e.ln < e.cp {
		e.ln++
	}

	e.evs[e.cp-e.cur-1] = ev
}

func (e *events) List() []Event {
	if e.ln < e.cp {
		return e.evs[e.cp-e.ln:]
	}

	idx := e.cp - e.cur - 1
	return append(e.evs[idx:], e.evs[:idx]...)
}
