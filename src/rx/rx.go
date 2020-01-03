package rx

import (
	"io/ioutil"
	oslog "log"
	"os"
	"sync"
	"time"
)

// -----------------------------------------------------------------------------
// Config

// Version export
const Version = "0.9.0"

// DEBUG flag for development
const DEBUG = false

// stand-in for system logger
var log *oslog.Logger

// Config export
func Config(debug bool) {
	if debug {
		log = oslog.New(os.Stderr, "RxGo ", oslog.Ltime|oslog.Lshortfile)
	} else {
		log = oslog.New(ioutil.Discard, "", 0)
	}
}

func init() {
	Config(DEBUG)
}

// -----------------------------------------------------------------------------
// Observer

// Observer type
type Observer struct {
	Next     chan interface{}
	Error    chan error
	Complete chan bool
	close    chan bool
	closed   bool
}

// NewObserver init
func NewObserver() *Observer {
	log.Println("Observer.NewObserver")
	id := &Observer{
		Next:     make(chan interface{}, 10),
		Error:    make(chan error, 1),
		Complete: make(chan bool, 1),
		close:    make(chan bool, 1),
		closed:   false,
	}

	go func() {
		defer func() {
			// the closed gate is inherently a race condition
			id.closed = true
			close(id.Next)
			close(id.Error)
			close(id.Complete)
			close(id.close)
		}()
		for {
			select {
			case <-id.close:
				if len(id.Next)+len(id.Error)+len(id.Complete) == 0 {
					return
				}
				// retry after 100ms
				<-time.After(100 * time.Millisecond)
				id.close <- true
				break
			}
		}
	}()

	return id
}

// next helper
func (id *Observer) next(event interface{}) *Observer {
	log.Println("Observer.next")
	if id.closed {
		return id
	}

	select {
	// prevent write blocks
	case id.Next <- event:
		break
	default:
		log.Println("Observer.next no channel")
		break
	}

	return id
}

// error helper
func (id *Observer) error(err error) *Observer {
	log.Println("Observer.Error")
	if id.closed {
		return id
	}

	select {
	// prevent write blocks
	case id.Error <- err:
		break
	default:
		log.Println("Observer.error no channel")
		break
	}

	id.close <- true
	return id
}

// complete helper
func (id *Observer) complete() *Observer {
	log.Println("Observer.Complete")
	if id.closed {
		return id
	}

	select {
	// prevent write blocks
	case id.Complete <- true:
		break
	default:
		log.Println("Observer.complete no channel")
		break
	}

	id.close <- true
	return id
}

// -----------------------------------------------------------------------------
// RxObservable

// Observable type
type Observable struct {
	*Observer
	observers   map[*Observer]*Observer
	multicast   bool
	Subscribe   chan *Observer
	Unsubscribe chan *Observer
	buffer      *CircularBuffer
	filtered    func(value interface{}) bool
	mapped      func(value interface{}) interface{}
	take        func()
}

// NewObservable init
func NewObservable() *Observable {
	log.Println("Observable.NewObservable")
	id := &Observable{
		Observer:    NewObserver(),
		observers:   map[*Observer]*Observer{},
		multicast:   false,
		Subscribe:   make(chan *Observer, 1),
		Unsubscribe: make(chan *Observer, 1),
		buffer:      nil,
		filtered:    nil,
		mapped:      nil,
		take:        nil,
	}

	// block to allow the reader goroutine to spin up
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer func() {
			close(id.Subscribe)
			close(id.Unsubscribe)
		}()
		log.Println("Observable.NewObservable ready")
		wg.Done()
		for {
			select {
			case event := <-id.Next:
				id.onNext(event)
				break
			case observer := <-id.Subscribe:
				id.onSubscribe(observer)
				break
			case observer := <-id.Unsubscribe:
				id.onUnsubscribe(observer)
				break
			case err := <-id.Error:
				id.onError(err)
				return
			case <-id.Complete:
				id.onComplete()
				return
			}
		}
	}()

	wg.Wait()

	return id
}

// onNext handler
func (id *Observable) onNext(event interface{}) {
	log.Println("Observable.onNext")

	// pre handlers

	// Filter
	if id.filtered != nil {
		if !id.filtered(event) {
			return
		}
	}

	// Map
	if id.mapped != nil {
		event = id.mapped(event)
	}

	// Replay / Distinct
	if id.buffer != nil {
		id.buffer.Add(event)
	}

	// multicast the event
	for _, observer := range id.observers {
		observer.next(event)
	}

	// post handlers

	// Take
	if id.take != nil {
		id.take()
	}
}

// onSubscribe handler
func (id *Observable) onSubscribe(observer *Observer) {
	log.Println("Observable.onSubscribe")
	if id.multicast {
		id.observers[observer] = observer
	} else {
		id.observers = map[*Observer]*Observer{observer: observer}
	}

	// replay for the new sub
	if id.buffer != nil {
		log.Println("Observable.onSubscribe replay")
		i := -1
		for {
			var v interface{}
			i, v = id.buffer.Next(i)
			if i == -1 {
				break
			}
			observer.next(v)
		}
	}
}

// onUnsubscribe handler
func (id *Observable) onUnsubscribe(observer *Observer) {
	log.Println("Observable.onUnsubscribe")
	delete(id.observers, observer)
	observer.complete()
	observer.close <- true
}

// onError handler
func (id *Observable) onError(err error) {
	log.Println("Observable.onError")
	for _, observer := range id.observers {
		delete(id.observers, observer)
		observer.error(err)
	}
}

// onComplete handler
func (id *Observable) onComplete() {
	log.Println("Observable.onComplete")
	for _, observer := range id.observers {
		delete(id.observers, observer)
		observer.complete()
	}
}

//
// Modifiers
//

// Multicast modifier
func (id *Observable) Multicast() *Observable {
	log.Println("Observable.Multicast")
	id.multicast = true
	return id
}

// Behavior modifier
func (id *Observable) Behavior(value interface{}) *Observable {
	log.Println("Observable.Behavior")
	id.buffer = NewCircularBuffer(1)
	id.buffer.Add(value)
	return id
}

// Replay modifier
func (id *Observable) Replay(bufferSize int) *Observable {
	log.Println("Observable.Replay")
	id.buffer = NewCircularBuffer(bufferSize)
	return id
}

//
// Operators
//

// Pipe operator
func (id *Observable) Pipe(sub *Observable) *Observable {
	id.Subscribe <- sub.Observer
	return id
}

// Yeild operator
// sleep just enough to allow the observable to yield
func (id *Observable) Yeild() *Observable {
	<-time.After(1 * time.Millisecond)
	return id
}

// -----------------------------------------------------------------------------
// Subjects

// NewSubject init
func NewSubject() *Observable {
	log.Println("Observable.NewSubject")
	id := NewObservable()
	id.multicast = true

	return id
}

// NewBehaviorSubject init
func NewBehaviorSubject(value interface{}) *Observable {
	log.Println("Observable.NewSubject")
	id := NewObservable()
	id.multicast = true
	id.Behavior(value)

	return id
}

// NewReplaySubject init
func NewReplaySubject(bufferSize int) *Observable {
	log.Println("Observable.NewSubject")
	id := NewObservable()
	id.multicast = true
	id.Replay(bufferSize)

	return id
}
