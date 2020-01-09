package rx

import (
	"strconv"
	"sync"
	"time"
)

// Event type
type Event struct {
	value    interface{}
	err      error
	complete bool
}

// Subscription type
type Subscription struct {
	Next         chan interface{}
	Error        chan error
	Complete     chan bool
	eventChan    chan Event
	finalizeOnce sync.Once
	closed       bool
	UID          string
	takeFn       func() bool
	observable   *Observable
}

// NewSubscription init
func NewSubscription() *Subscription {
	log.Println("Subscription.NewSubscription")
	id := &Subscription{
		eventChan:  make(chan Event, 10),
		Next:       make(chan interface{}, 1),
		Error:      make(chan error, 1),
		Complete:   make(chan bool, 1),
		closed:     false,
		UID:        strconv.FormatInt(time.Now().UnixNano(), 10),
		takeFn:     nil,
		observable: nil,
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

		hot := true
		for {
			dlog.Println(id.UID, "Subscription<-eventChan")
			event := <-id.eventChan
			switch {
			case event.value != nil:
				dlog.Println(id.UID, "Subscription<-Next")
				if hot {
					id.Next <- event.value
				}

				// Take
				if id.takeFn != nil && !id.takeFn() {
					hot = false
					dlog.Println(id.UID, "Take complete")
					if id.observable != nil {
						id.observable.Unsubscribe <- id
						id.observable.yield()
					}
					id.complete()
				}
				break
			case event.err != nil:
				dlog.Println(id.UID, "Subscription<-Error")
				id.Error <- event.err
				return
			case event.complete == true:
				dlog.Println(id.UID, "Subscription<-Complete")
				id.Complete <- true
				return
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

	id.eventChan <- Event{value: event}

	return id
}

// error helper
func (id *Subscription) error(err error) *Subscription {
	log.Println(id.UID, "Subscription.error")

	id.finalizeOnce.Do(func() {
		id.eventChan <- Event{err: err}
	})

	return id
}

// complete helper
func (id *Subscription) complete() *Subscription {
	log.Println(id.UID, "Subscription.complete")

	id.finalizeOnce.Do(func() {
		id.eventChan <- Event{complete: true}
	})

	return id
}

// Default provides a default implementation which logs to the default output
func (id *Subscription) Default(handler func(event interface{}), closeCh chan bool) *Subscription {
	log.Println(id.UID, "Subscription.Default")
	defer func() { closeCh <- true }()
	for {
		select {
		case event := <-id.Next:
			log.Println(id.UID, "<-Next", event)
			if handler != nil {
				handler(event)
			}
			break
		case err := <-id.Error:
			log.Println(id.UID, "<-Error", err)
			return id
		case <-id.Complete:
			log.Println(id.UID, "<-Complete")
			return id
		}
	}
}
