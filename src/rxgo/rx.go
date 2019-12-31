package rxgo

import (
	"io/ioutil"
	"log"
	"time"
)

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
	Close    chan bool
	closed   bool
}

// NewRxObserver init
func NewRxObserver() *RxObserver {
	log.Println("RxObserver::NewRxObserver")
	id := &RxObserver{
		Next:     make(chan interface{}, 1),
		Error:    make(chan error, 1),
		Complete: make(chan bool, 1),
		Close:    make(chan bool, 1),
		closed:   false,
	}

	go func() {
		defer func() {
			id.closed = true
			close(id.Next)
			close(id.Error)
			close(id.Complete)
			close(id.Close)
		}()
		for {
			select {
			case <-id.Close:
				if len(id.Next)+len(id.Error)+len(id.Complete) == 0 {
					return
				}
				// retry after 100ms
				<-time.After(100 * time.Millisecond)
				id.Close <- true
				break
			}
		}
	}()

	return id
}

// next helper
func (id *RxObserver) next(event interface{}) {
	log.Println("RxObserver::next", event)
	if id.closed {
		return
	}

	select {
	// prevent write blocks
	case id.Next <- event:
		break
	default:
		break
	}
}

// error helper
func (id *RxObserver) error(err error) {
	log.Println("RxObserver::error")
	if id.closed {
		return
	}

	select {
	// prevent write blocks
	case id.Error <- err:
		break
	default:
		break
	}

	id.Close <- true
}

// complete helper
func (id *RxObserver) complete() {
	log.Println("RxObserver::onComplete")
	if id.closed {
		return
	}

	select {
	// prevent write blocks
	case id.Complete <- true:
		break
	default:
		break
	}

	id.Close <- true
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
		take:        nil,
	}

	go func() {
		defer func() {
			close(id.Subscribe)
			close(id.Unsubscribe)
		}()
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

	return id
}

// onNext handler
func (id *RxObservable) onNext(event interface{}) {
	log.Println("RxObservable::onNext")

	for _, observer := range id.observers {
		observer.next(event)
	}

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
}

// onUnsubscribe handler
func (id *RxObservable) onUnsubscribe(observer *RxObserver) {
	log.Println("RxObservable::onUnsubscribe")
	delete(id.observers, observer)
	observer.complete()
	observer.Close <- true
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

// Multicast helper
func (id *RxObservable) Multicast() {
	log.Println("RxObservable::multicast")
	id.multicast = true
}

// -----------------------------------------------------------------------------
// RxSubject

// RxSubject type
type RxSubject struct {
	*RxObservable
}

// NewRxSubject init
func NewRxSubject() *RxSubject {
	log.Println("RxObservable::NewRxSubject")
	id := &RxSubject{
		RxObservable: NewRxObservable(),
	}
	id.multicast = true

	return id
}
