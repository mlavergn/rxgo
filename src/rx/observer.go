package rx

import (
	"strconv"
	"sync"
	"time"
)

// EventType export
type EventType int

// Event types
const (
	EventTypeNext EventType = iota
	EventTypeError
	EventTypeComplete
)

// Event type
type Event struct {
	Type     EventType
	Next     interface{}
	Error    error
	Complete *Observable
}

// Observer type
type Observer struct {
	Event        chan Event
	finalizeOnce sync.Once
	closed       bool
	UID          string
}

// NewObserver init
func NewObserver() *Observer {
	log.Println("Observer.NewObserver")
	id := &Observer{
		Event:  make(chan Event, 1),
		closed: false,
		UID:    strconv.FormatInt(time.Now().UnixNano(), 10),
	}

	return id
}

// next helper
func (id *Observer) next(event interface{}) *Observer {
	log.Println(id.UID, "Observer.next")

	if !id.closed && event != nil {
		id.Event <- Event{Type: EventTypeNext, Next: event}
	}

	return id
}

// error helper
func (id *Observer) error(err error) *Observer {
	log.Println(id.UID, "Observer.error")

	id.finalizeOnce.Do(func() {
		id.closed = true
		id.Event <- Event{Type: EventTypeError, Error: err}
		close(id.Event)
	})

	return id
}

// complete helper
func (id *Observer) complete(obs *Observable) *Observer {
	log.Println(id.UID, "Observer.complete")

	id.finalizeOnce.Do(func() {
		id.closed = true
		id.Event <- Event{Type: EventTypeComplete, Complete: obs}
		close(id.Event)
	})

	return id
}
