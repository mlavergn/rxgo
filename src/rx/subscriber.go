package rx

import (
	"sync"
	"time"
)

// -----------------------------------------------------------------------------
// Observer

// Subscriber type
type Subscriber struct {
	Next         chan interface{}
	nextChan     chan interface{}
	Error        chan error
	errorChan    chan error
	Complete     chan bool
	completeChan chan bool
	finalize     chan bool
	finalized    bool
	closed       bool
	UID          int64
	take         func() bool
	hot          bool
}

// NewSubscriber init
func NewSubscriber() *Subscriber {
	log.Println("Subscriber.NewSubscriber")
	id := &Subscriber{
		Next:         make(chan interface{}, 1),
		nextChan:     make(chan interface{}, 10),
		Error:        make(chan error, 1),
		errorChan:    make(chan error, 1),
		Complete:     make(chan bool, 1),
		completeChan: make(chan bool, 1),
		finalize:     make(chan bool, 1),
		finalized:    false,
		closed:       false,
		UID:          time.Now().UnixNano(),
		take:         nil,
		hot:          true,
	}

	// block to allow the reader goroutine to spin up
	var wg sync.WaitGroup
	wg.Add(2)

	// next loop
	go func() {
		defer func() {
			recover()
		}()

		wg.Done()
		for {
			if id.closed || !id.hot {
				return
			}

			id.Next <- <-id.nextChan

			// Take
			if id.take != nil {
				if !id.take() {
					id.hot = false
					id.complete()
				}
			}
		}
	}()

	// error/complete loop
	go func() {
		defer func() {
			if !id.closed {
				id.closed = true
				close(id.Next)
				close(id.nextChan)
				close(id.Error)
				close(id.errorChan)
				close(id.Complete)
				close(id.completeChan)
				close(id.finalize)
			}
		}()

		wg.Done()
		for {
			select {
			case err := <-id.errorChan:
				if !id.finalized {
					id.finalized = true
					id.Yield(1)
					id.Error <- err
					id.Yield(1)
					id.finalize <- true
					id.Yield(10)
					return
				}
				break
			case <-id.completeChan:
				if !id.finalized {
					id.finalized = true
					id.Yield(1)
					id.Complete <- true
					id.Yield(1)
					id.finalize <- true
					id.Yield(10)
					return
				}
				break
			}
		}
	}()

	wg.Wait()

	return id
}

// next helper
func (id *Subscriber) next(event interface{}) *Subscriber {
	log.Println(id.UID, "Subscriber.next")
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
func (id *Subscriber) error(err error) *Subscriber {
	log.Println(id.UID, "Subscriber.Error")
	if id.closed {
		return nil
	}

	defer func() {
		recover()
	}()
	id.errorChan <- err

	return id
}

// complete helper
func (id *Subscriber) complete() *Subscriber {
	log.Println(id.UID, "Subscriber.Complete")
	if id.closed {
		return nil
	}

	defer func() {
		recover()
	}()
	id.completeChan <- true

	return id
}

// Yield operator
// sleep to allow the observable to yield
func (id *Subscriber) Yield(ms time.Duration) *Subscriber {
	log.Println(id.UID, "Observable.Yield")
	<-time.After(ms * time.Millisecond)
	return id
}
