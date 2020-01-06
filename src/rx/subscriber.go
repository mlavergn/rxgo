package rx

import (
	"sync"
	"time"
)

// -----------------------------------------------------------------------------
// Observer

// Subscriber type
type Subscriber struct {
	Next      chan interface{}
	nextChan  chan interface{}
	Error     chan error
	Complete  chan bool
	closeChan chan error
	finalize  chan bool
	UID       int64
	take      func() bool
}

// NewSubscriber init
func NewSubscriber() *Subscriber {
	log.Println("Subscriber.NewSubscriber")
	id := &Subscriber{
		Next:      make(chan interface{}, 1),
		nextChan:  make(chan interface{}, 10),
		Error:     make(chan error, 1),
		Complete:  make(chan bool, 1),
		closeChan: make(chan error, 1),
		finalize:  make(chan bool, 1),
		UID:       time.Now().UnixNano(),
		take:      nil,
	}

	// block to allow the reader goroutine to spin up
	var wg sync.WaitGroup
	wg.Add(2)

	// next handler
	go func() {
		defer func() {
			recover()
		}()

		hot := true

		wg.Done()
		for {
			if !hot {
				return
			}

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
			close(id.Next)
			close(id.nextChan)
			close(id.Error)
			close(id.Complete)
			close(id.closeChan)
			close(id.finalize)
		}()

		wg.Done()
		err := <-id.closeChan
		id.Yield(1)
		if err != nil {
			id.Error <- err
		} else {
			id.Complete <- true
		}
		id.Yield(1)
		id.finalize <- true
		id.Yield(1)
	}()

	wg.Wait()

	return id
}

// next helper
func (id *Subscriber) next(event interface{}) *Subscriber {
	log.Println(id.UID, "Subscriber.next")
	defer func() {
		recover()
	}()
	id.nextChan <- event

	return id
}

// error helper
func (id *Subscriber) error(err error) *Subscriber {
	log.Println(id.UID, "Subscriber.Error")
	defer func() {
		recover()
	}()
	id.closeChan <- err

	return id
}

// complete helper
func (id *Subscriber) complete() *Subscriber {
	log.Println(id.UID, "Subscriber.Complete")
	defer func() {
		recover()
	}()
	id.closeChan <- nil

	return id
}

// Yield operator
// sleep to allow the observable to yield
func (id *Subscriber) Yield(ms time.Duration) *Subscriber {
	log.Println(id.UID, "Observable.Yield")
	<-time.After(ms * time.Millisecond)
	return id
}
