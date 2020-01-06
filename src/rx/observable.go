package rx

import (
	"sync"
)

// Observable type
type Observable struct {
	*Observer
	observers   map[*Observer]*Observer
	Connect     chan bool
	isMulticast bool
	Subscribe   chan *Observer
	Unsubscribe chan *Observer
	finally     chan bool
	merged      int8
	buffer      *CircularBuffer
	filtered    func(value interface{}) bool
	mapped      func(value interface{}) interface{}
}

// NewObservable init
func NewObservable() *Observable {
	log.Println("Observable.NewObservable")
	id := &Observable{
		Observer:    NewObserver(),
		observers:   map[*Observer]*Observer{},
		Connect:     make(chan bool, 1),
		isMulticast: false,
		Subscribe:   make(chan *Observer, 1),
		Unsubscribe: make(chan *Observer, 1),
		finally:     make(chan bool, 1),
		merged:      0,
		buffer:      nil,
		filtered:    nil,
		mapped:      nil,
	}

	// block to allow the reader goroutine to spin up
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer func() {
			id.finally <- true
			id.Delay(1)
			close(id.finally)
			id.Delay(1)
			id.complete()
			id.Delay(1)
			close(id.Subscribe)
			close(id.Unsubscribe)
		}()
		wg.Done()
		for {
			select {
			case event := <-id.Next:
				id.onNext(event)
				break
			case err := <-id.Error:
				if id.merged != 0 {
					break
				}
				id.onError(err)
				return
			case <-id.Complete:
				if id.merged != 0 {
					break
				}
				id.onComplete()
				return
			case observer := <-id.Subscribe:
				id.onSubscribe(observer)
				break
			case observer := <-id.Unsubscribe:
				id.onUnsubscribe(observer)
				// when we go cold, complete the obserable
				if id.merged != 0 {
					break
				}
				return
			}
		}
	}()

	wg.Wait()

	return id
}

// onNext handler
func (id *Observable) onNext(event interface{}) {
	log.Println(id.UID, "Observable.onNext")

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
		if observer.next(event) == nil {
			id.Unsubscribe <- observer
		}
	}
}

// onError handler
func (id *Observable) onError(err error) {
	log.Println(id.UID, "Observable.onError")
	for _, observer := range id.observers {
		delete(id.observers, observer)
		observer.error(err)
	}
}

// onComplete handler
func (id *Observable) onComplete() {
	log.Println(id.UID, "Observable.onComplete")
	for _, observer := range id.observers {
		delete(id.observers, observer)
		observer.complete()
	}
}

// onSubscribe handler
func (id *Observable) onSubscribe(observer *Observer) {
	log.Println(id.UID, "Observable.onSubscribe")
	if !id.isMulticast {
		for _, observer := range id.observers {
			delete(id.observers, observer)
			observer.complete()
		}
	}
	id.observers[observer] = observer
	observer.Subscription = id

	// connect on Subscribe
	select {
	case id.Connect <- true:
		break
	default:
		break
	}

	// replay for the new sub
	if id.buffer != nil {
		log.Println(id.UID, "Observable.onSubscribe replay")
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
	log.Println(id.UID, "Observable.onUnsubscribe")
	observer.Subscription = nil
	delete(id.observers, observer)
	if !observer.closed {
		observer.complete()
	}
}

//
// Modifiers
//

// Multicast modifier
func (id *Observable) Multicast() *Observable {
	log.Println(id.UID, "Observable.Multicast")
	id.isMulticast = true
	return id
}

// Behavior modifier
func (id *Observable) Behavior(value interface{}) *Observable {
	log.Println(id.UID, "Observable.Behavior")
	id.buffer = NewCircularBuffer(1)
	id.buffer.Add(value)
	return id
}

// Replay modifier
func (id *Observable) Replay(bufferSize int) *Observable {
	log.Println(id.UID, "Observable.Replay")
	id.buffer = NewCircularBuffer(bufferSize)
	return id
}

// Merge operator
func (id *Observable) Merge(merge *Observable) *Observable {
	log.Println(id.UID, "Observable.Merge")

	merge.Multicast()
	id.merged++
	proxy := NewObserver()
	go func(proxy *Observer) {
		defer func() {
			id.merged--
			proxy.complete()
		}()
		for {
			select {
			case <-proxy.Error:
				return
			case <-proxy.Complete:
				return
			}
		}
	}(id.Observer)
	merge.Subscribe <- proxy
	merge.Pipe(id)

	return id
}
