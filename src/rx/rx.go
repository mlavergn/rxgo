package rx

import (
	"io/ioutil"
	"log"
	"sync"
	"time"
)

// Version export
const Version = "0.7.0"

// RxSetup export
func RxSetup(prod bool) {
	if prod {
		log.SetFlags(0)
		log.SetOutput(ioutil.Discard)
	} else {
		log.SetFlags(log.Ltime | log.Lshortfile)
	}
}

// -----------------------------------------------------------------------------
// RxObserver

// RxObserver type
type RxObserver struct {
	Next     chan interface{}
	Error    chan error
	Complete chan bool
	close    chan bool
	closed   bool
}

// NewRxObserver init
func NewRxObserver() *RxObserver {
	log.Println("RxObserver::NewRxObserver")
	id := &RxObserver{
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
func (id *RxObserver) next(event interface{}) *RxObserver {
	log.Println("RxObserver::next")
	if id.closed {
		return id
	}

	select {
	// prevent write blocks
	case id.Next <- event:
		break
	default:
		log.Println("RxObserver::next no channel")
		break
	}

	return id
}

// error helper
func (id *RxObserver) error(err error) *RxObserver {
	log.Println("RxObserver::Error")
	if id.closed {
		return id
	}

	select {
	// prevent write blocks
	case id.Error <- err:
		break
	default:
		log.Println("RxObserver::error no channel")
		break
	}

	id.close <- true
	return id
}

// complete helper
func (id *RxObserver) complete() *RxObserver {
	log.Println("RxObserver::Complete")
	if id.closed {
		return id
	}

	select {
	// prevent write blocks
	case id.Complete <- true:
		break
	default:
		log.Println("RxObserver::complete no channel")
		break
	}

	id.close <- true
	return id
}

// -----------------------------------------------------------------------------
// RxObservable

// RxObservable type
type RxObservable struct {
	*RxObserver
	observers   map[*RxObserver]*RxObserver
	multicast   bool
	Subscribe   chan *RxObserver
	Unsubscribe chan *RxObserver
	buffer      *CircularBuffer
	filter      func(value interface{}) bool
	take        func()
}

// NewRxObservable init
func NewRxObservable() *RxObservable {
	log.Println("RxObservable::NewRxObservable")
	id := &RxObservable{
		RxObserver:  NewRxObserver(),
		observers:   map[*RxObserver]*RxObserver{},
		multicast:   false,
		Subscribe:   make(chan *RxObserver, 1),
		Unsubscribe: make(chan *RxObserver, 1),
		buffer:      nil,
		filter:      nil,
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
		log.Println("RxObservable::NewRxObservable ready")
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
func (id *RxObservable) onNext(event interface{}) {
	log.Println("RxObservable::onNext")

	// pre handlers
	if id.filter != nil {
		if !id.filter(event) {
			return
		}
	}

	// buffer for replay / distinct etc
	if id.buffer != nil {
		id.buffer.Add(event)
	}

	for _, observer := range id.observers {
		observer.next(event)
	}

	// post handlers
	if id.take != nil {
		id.take()
	}
}

// onSubscribe handler
func (id *RxObservable) onSubscribe(observer *RxObserver) {
	log.Println("RxObservable::onSubscribe")
	if id.multicast {
		id.observers[observer] = observer
	} else {
		id.observers = map[*RxObserver]*RxObserver{observer: observer}
	}

	// replay for the new sub
	if id.buffer != nil {
		log.Println("RxObservable::onSubscribe replay")
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
func (id *RxObservable) onUnsubscribe(observer *RxObserver) {
	log.Println("RxObservable::onUnsubscribe")
	delete(id.observers, observer)
	observer.complete()
	observer.close <- true
}

// onError handler
func (id *RxObservable) onError(err error) {
	log.Println("RxObservable::onError")
	for _, observer := range id.observers {
		delete(id.observers, observer)
		observer.error(err)
	}
}

// onComplete handler
func (id *RxObservable) onComplete() {
	log.Println("RxObservable::onComplete")
	for _, observer := range id.observers {
		delete(id.observers, observer)
		observer.complete()
	}
}

//
// Modifiers
//

// Multicast modifier
func (id *RxObservable) Multicast() *RxObservable {
	log.Println("RxObservable::Multicast")
	id.multicast = true
	return id
}

// Behavior modifier
func (id *RxObservable) Behavior(value interface{}) *RxObservable {
	log.Println("RxObservable::Behavior")
	id.buffer = NewCircularBuffer(1)
	id.buffer.Add(value)
	return id
}

// Replay modifier
func (id *RxObservable) Replay(bufferSize int) *RxObservable {
	log.Println("RxObservable::Replay")
	id.buffer = NewCircularBuffer(bufferSize)
	return id
}

//
// Operators
//

// Pipe operator
func (id *RxObservable) Pipe(sub *RxObservable) *RxObservable {
	id.Subscribe <- sub.RxObserver
	return id
}

// Warmup operator
// sleep just enough to allow the observable to warm
func (id *RxObservable) Warmup() *RxObservable {
	<-time.After(1 * time.Millisecond)
	return id
}

// -----------------------------------------------------------------------------
// RxSubjects

// NewRxSubject init
func NewRxSubject() *RxObservable {
	log.Println("RxObservable::NewRxSubject")
	id := NewRxObservable()
	id.multicast = true

	return id
}

// NewRxBehaviorSubject init
func NewRxBehaviorSubject(value interface{}) *RxObservable {
	log.Println("RxObservable::NewRxSubject")
	id := NewRxObservable()
	id.multicast = true
	id.Behavior(value)

	return id
}

// NewRxReplaySubject init
func NewRxReplaySubject(bufferSize int) *RxObservable {
	log.Println("RxObservable::NewRxSubject")
	id := NewRxObservable()
	id.multicast = true
	id.Replay(bufferSize)

	return id
}
