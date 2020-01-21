package rx

import (
	"strconv"
	"sync"
	"time"
)

// Event type
type Event struct {
	evNext     interface{}
	evError    error
	evComplete *Observable
}

// Observer type
type Observer struct {
	eventChan    chan Event
	Next         chan interface{}
	Error        chan error
	Complete     chan *Observable
	finalizeOnce sync.Once
	closed       bool
	UID          string
}

// NewObserver init
func NewObserver() *Observer {
	log.Println("Observer.NewObserver")
	id := &Observer{
		eventChan: make(chan Event, 10),
		Next:      make(chan interface{}, 1),
		Error:     make(chan error, 1),
		Complete:  make(chan *Observable, 1),
		closed:    false,
		UID:       strconv.FormatInt(time.Now().UnixNano(), 10),
	}

	// waiter to allow the goroutine to spin up
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer func() {
			id.closed = true
			log.Println(id.UID, "Observer.finalize")
			close(id.eventChan)
			close(id.Next)
			close(id.Error)
			close(id.Complete)
		}()
		wg.Done()

		for {
			dlog.Println(id.UID, "Observer<-eventChan")
			select {
			case event := <-id.eventChan:
				switch {
				case event.evNext != nil:
					dlog.Println(id.UID, "Observer<-Next")
					id.Next <- event.evNext
					break
				case event.evError != nil:
					dlog.Println(id.UID, "Observer<-Error")
					id.Error <- event.evError
					return
				case event.evComplete != nil:
					dlog.Println(id.UID, "Observer<-Complete")
					id.Complete <- event.evComplete
					return
				}
			}
		}
	}()

	wg.Wait()

	return id
}

// next helper
func (id *Observer) next(event interface{}) *Observer {
	log.Println(id.UID, "Observer.next")
	if id.closed {
		return nil
	}

	id.eventChan <- Event{evNext: event}

	return id
}

// error helper
func (id *Observer) error(err error) *Observer {
	log.Println(id.UID, "Observer.error")

	id.finalizeOnce.Do(func() {
		id.eventChan <- Event{evError: err}
	})

	return id
}

// complete helper
func (id *Observer) complete(obs *Observable) *Observer {
	log.Println(id.UID, "Observer.complete")

	id.finalizeOnce.Do(func() {
		id.eventChan <- Event{evComplete: obs}
	})

	return id
}
