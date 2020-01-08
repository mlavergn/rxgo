package rx

import (
	"sync"
)

type operatorType int

const (
	operatorMap operatorType = iota
	operatorFilter
)

type operator struct {
	op operatorType
	fn interface{}
}

// Observable type
type Observable struct {
	*Subscription
	observers    map[*Subscription]*Subscription
	Connect      chan bool
	isMulticast  bool
	Subscribe    chan *Subscription
	Unsubscribe  chan *Subscription
	finally      chan bool
	merged       int8
	buffer       *CircularBuffer
	operators    []operator
	retryWhen    func() bool
	retryFn      func(*Observable)
	catchErrorFn func(error)
}

// NewObservable init
func NewObservable() *Observable {
	log.Println("Observable.NewObservable")
	id := &Observable{
		Subscription: NewSubscription(),
		observers:    map[*Subscription]*Subscription{},
		Connect:      make(chan bool, 1),
		isMulticast:  false,
		Subscribe:    make(chan *Subscription, 1),
		Unsubscribe:  make(chan *Subscription, 1),
		finally:      make(chan bool, 1),
		merged:       0,
		buffer:       nil,
		operators:    []operator{},
		retryWhen:    nil,
		retryFn:      nil,
	}

	// block to allow the reader goroutine to spin up
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer func() {
			log.Println(id.UID, "Observable.finalize")
			id.finally <- true
			id.Delay(1)
			close(id.finally)
			id.Delay(1)
			id.complete()
			id.Delay(1)
			close(id.Subscribe)
			close(id.Unsubscribe)
			id.observers = nil
			id.operators = nil
		}()
		wg.Done()
		for {
			select {
			case event := <-id.Next:
				id.onNext(event)
				break
			case err := <-id.Error:
				if id.catchErrorFn != nil {
					id.catchErrorFn(err)
					id.Complete <- true
					break
				}
				id.onError(err)
				return
			case <-id.Complete:
				if id.merged != 0 {
					break
				}
				if id.doRetry() {
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

	// Operations
	for _, op := range id.operators {
		switch op.op {
		case operatorFilter:
			filterFn := op.fn.(func(interface{}) bool)
			if !filterFn(event) {
				return
			}
			break
		case operatorMap:
			mapFn := op.fn.(func(interface{}) interface{})
			mappedEvent := mapFn(event)
			if mappedEvent == nil {
				return
			}
			event = mappedEvent
			break
		}
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
func (id *Observable) onSubscribe(observer *Subscription) {
	log.Println(id.UID, "Observable.onSubscribe")
	if !id.isMulticast {
		for _, observer := range id.observers {
			delete(id.observers, observer)
			observer.complete()
		}
	}
	id.observers[observer] = observer
	observer.observable = id

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
func (id *Observable) onUnsubscribe(observer *Subscription) {
	log.Println(id.UID, "Observable.onUnsubscribe")
	observer.observable = nil
	delete(id.observers, observer)
	if !observer.closed {
		observer.complete()
	}
}

// return handler
func (id *Observable) doRetry() bool {
	if id.retryWhen != nil && id.retryFn != nil {
		if id.retryWhen() {
			id.Subscription = NewSubscription()
			id.retryFn(id)
			return true
		}
	}
	return false
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
	proxy := NewSubscription()
	go func(subscription *Subscription) {
		defer func() {
			recover()
			id.merged--
		}()
		for {
			select {
			case err := <-subscription.Error:
				id.Error <- err
				return
			case <-subscription.Complete:
				id.Complete <- true
				return
			}
		}
	}(id.Subscription)
	merge.Subscribe <- proxy
	merge.Pipe(id)

	return id
}

// RetryWhen modifier
func (id *Observable) RetryWhen(fn func() bool) *Observable {
	log.Println(id.UID, "Observable.Retry")
	id.retryWhen = fn
	return id
}

// CatchError modifier
func (id *Observable) CatchError(fn func(error)) *Observable {
	log.Println(id.UID, "Observable.CatchError")
	id.catchErrorFn = fn
	return id
}
