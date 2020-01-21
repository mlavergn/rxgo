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

// Subscription type
type Subscription struct {
	eventChan    chan Event
	Next         chan interface{}
	Error        chan error
	Complete     chan *Observable
	finalizeOnce sync.Once
	closed       bool
	UID          string
}

// NewSubscription init
func NewSubscription() *Subscription {
	log.Println("Subscription.NewSubscription")
	id := &Subscription{
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
			log.Println(id.UID, "Subscription.finalize")
			close(id.eventChan)
			close(id.Next)
			close(id.Error)
			close(id.Complete)
		}()
		wg.Done()

		for {
			dlog.Println(id.UID, "Subscription<-eventChan")
			select {
			case event := <-id.eventChan:
				switch {
				case event.evNext != nil:
					dlog.Println(id.UID, "Subscription<-Next")
					id.Next <- event.evNext
					break
				case event.evError != nil:
					dlog.Println(id.UID, "Subscription<-Error")
					id.Error <- event.evError
					return
				case event.evComplete != nil:
					dlog.Println(id.UID, "Subscription<-Complete")
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
func (id *Subscription) next(event interface{}) *Subscription {
	log.Println(id.UID, "Subscription.next")
	if id.closed {
		return nil
	}

	id.eventChan <- Event{evNext: event}

	return id
}

// error helper
func (id *Subscription) error(err error) *Subscription {
	log.Println(id.UID, "Subscription.error")

	id.finalizeOnce.Do(func() {
		id.eventChan <- Event{evError: err}
	})

	return id
}

// complete helper
func (id *Subscription) complete(obs *Observable) *Subscription {
	log.Println(id.UID, "Subscription.complete")

	id.finalizeOnce.Do(func() {
		id.eventChan <- Event{evComplete: obs}
	})

	return id
}
