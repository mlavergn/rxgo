package rx

import (
	"fmt"
	"strconv"
	"sync"
	"time"
)

// Subscription type
type Subscription struct {
	Next       chan interface{}
	nextChan   chan interface{}
	Error      chan error
	Complete   chan bool
	closeChan  chan error
	closed     bool
	UID        string
	take       func() bool
	observable *Observable
}

// NewSubscription init
func NewSubscription() *Subscription {
	log.Println("Subscription.NewSubscription")
	id := &Subscription{
		Next:       make(chan interface{}, 1),
		nextChan:   make(chan interface{}, 10),
		Error:      make(chan error, 1),
		Complete:   make(chan bool, 1),
		closeChan:  make(chan error, 1),
		closed:     false,
		UID:        strconv.FormatInt(time.Now().UnixNano(), 10),
		take:       nil,
		observable: nil,
	}

	// block to allow the reader goroutine to spin up
	var wg sync.WaitGroup
	wg.Add(2)

	// next handler
	go func() {
		defer func() {
			recover()
		}()
		wg.Done()

		hot := true
		for hot {
			id.Next <- <-id.nextChan

			// Take
			if id.take != nil {
				if !id.take() {
					hot = false
					id.complete()
				}
			}
		}
	}()

	// error/complete handler
	go func() {
		defer func() {
			recover()
			id.closed = true
			log.Println(id.UID, "Subscription.finalize")
			close(id.Next)
			close(id.nextChan)
			close(id.Error)
			close(id.Complete)
			close(id.closeChan)
		}()
		wg.Done()

		err := <-id.closeChan
		id.Delay(1)
		if err != nil {
			id.Error <- err
		} else {
			id.Complete <- true
		}
		id.Delay(1)
		if id.observable != nil {
			id.observable.Unsubscribe <- id
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

	defer func() {
		recover()
	}()
	id.nextChan <- event

	return id
}

// error helper
func (id *Subscription) error(err error) *Subscription {
	log.Println(id.UID, "Subscription.Error")
	if id.closed {
		return nil
	}

	defer func() {
		recover()
	}()
	id.closeChan <- err

	return id
}

// complete helper
func (id *Subscription) complete() *Subscription {
	log.Println(id.UID, "Subscription.Complete")
	if id.closed {
		return nil
	}

	defer func() {
		recover()
	}()
	id.closeChan <- nil

	return id
}

// Default provides a default implementation which logs to the default output
func (id *Subscription) Default(handler func(event interface{}), closeCh chan bool) *Subscription {
	defer func() { closeCh <- true }()
	for {
		select {
		case event := <-id.Next:
			fmt.Println(id.UID, "Next", event)
			if handler != nil {
				handler(event)
			}
			break
		case err := <-id.Error:
			fmt.Println(id.UID, "Error", err)
			return id
		case <-id.Complete:
			fmt.Println(id.UID, "Complete")
			return id
		}
	}
}

//
// Operators
//

// Delay operator
// sleep allows the observable to yield to the go channel
func (id *Subscription) Delay(ms time.Duration) *Subscription {
	// log.Println(id.UID, "Observable.Delay")
	<-time.After(ms * time.Millisecond)
	return id
}
