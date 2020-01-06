package rx

import (
	"sync"
)

// -----------------------------------------------------------------------------
// RxObservable

// Observable type
type Observable struct {
	*Subscriber
	subscribers map[*Subscriber]*Subscriber
	Connect     chan bool
	isMulticast bool
	Subscribe   chan *Subscriber
	Unsubscribe chan *Subscriber
	merged      map[*Observable]*Observable
	buffer      *CircularBuffer
	filtered    func(value interface{}) bool
	mapped      func(value interface{}) interface{}
}

// NewObservable init
func NewObservable() *Observable {
	log.Println("Observable.NewObservable")
	id := &Observable{
		Subscriber:  NewSubscriber(),
		subscribers: map[*Subscriber]*Subscriber{},
		Connect:     make(chan bool, 1),
		isMulticast: false,
		Subscribe:   make(chan *Subscriber, 1),
		Unsubscribe: make(chan *Subscriber, 1),
		merged:      map[*Observable]*Observable{},
		buffer:      nil,
		filtered:    nil,
		mapped:      nil,
	}

	// block to allow the reader goroutine to spin up
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer func() {
			close(id.Subscribe)
			close(id.Unsubscribe)
			id.complete()
		}()
		wg.Done()
		for {
			select {
			case event := <-id.Next:
				id.onNext(event)
				break
			case err := <-id.Error:
				if len(id.merged) != 0 {
					break
				}
				id.onError(err)
				return
			case <-id.Complete:
				if len(id.merged) != 0 {
					break
				}
				id.onComplete()
				return
			case subscriber := <-id.Subscribe:
				id.onSubscribe(subscriber)
				break
			case subscriber := <-id.Unsubscribe:
				id.onUnsubscribe(subscriber)
				// when we go cold, complete the obserable
				if len(id.subscribers) == 0 {
					return
				}
				break
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
	for _, subscriber := range id.subscribers {
		if subscriber.next(event) == nil {
			id.Unsubscribe <- subscriber
		}
	}
}

// onError handler
func (id *Observable) onError(err error) {
	log.Println(id.UID, "Observable.onError")
	for key, subscriber := range id.subscribers {
		delete(id.subscribers, key)
		subscriber.error(err)
	}
}

// onComplete handler
func (id *Observable) onComplete() {
	log.Println(id.UID, "Observable.onComplete")
	for key, subscriber := range id.subscribers {
		delete(id.subscribers, key)
		subscriber.complete()
	}
}

// onSubscribe handler
func (id *Observable) onSubscribe(subscriber *Subscriber) {
	log.Println(id.UID, "Observable.onSubscribe")
	if !id.isMulticast {
		for key, subscriber := range id.subscribers {
			delete(id.subscribers, key)
			subscriber.complete()
		}
	}
	id.subscribers[subscriber] = subscriber

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
			subscriber.next(v)
		}
	}
}

// onUnsubscribe handler
func (id *Observable) onUnsubscribe(subscriber *Subscriber) {
	log.Println(id.UID, "Observable.onUnsubscribe")
	delete(id.subscribers, subscriber)
	subscriber.complete()
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
	id.merged[merge] = merge
	proxy := NewSubscriber()
	go func(proxy *Subscriber) {
		for {
			select {
			case event := <-proxy.Next:
				proxy.next(event)
				break
			case err := <-proxy.Error:
				delete(id.merged, merge)
				proxy.error(err)
				return
			case <-proxy.Complete:
				delete(id.merged, merge)
				proxy.complete()
				return
			}
		}
	}(id.Subscriber)
	merge.Subscribe <- proxy

	return id
}
