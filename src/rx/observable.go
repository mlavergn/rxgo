package rx

import (
	"sync"
	"time"
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
	observers     map[*Subscription]*Subscription
	Connect       chan bool
	shared        bool
	multicast     bool
	Subscribe     chan *Subscription
	Unsubscribe   chan *Subscription
	Finalize      chan bool
	merged        map[*Observable]*Observable
	buffer        *CircularBuffer
	operators     []operator
	repeatWhenFn  func() bool
	retryWhenFn   func() bool
	catchErrorFn  func(error)
	resubscribeFn func(*Observable)
}

// NewObservable init
func NewObservable() *Observable {
	log.Println("Observable.NewObservable")
	id := &Observable{
		Subscription:  NewSubscription(),
		observers:     map[*Subscription]*Subscription{},
		Connect:       make(chan bool, 1),
		shared:        false,
		multicast:     false,
		Subscribe:     make(chan *Subscription, 1),
		Unsubscribe:   make(chan *Subscription, 1),
		Finalize:      make(chan bool, 1),
		merged:        map[*Observable]*Observable{},
		buffer:        nil,
		operators:     []operator{},
		repeatWhenFn:  nil,
		retryWhenFn:   nil,
		catchErrorFn:  nil,
		resubscribeFn: nil,
	}

	// block to allow the reader goroutine to spin up
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer func() {
			log.Println(id.UID, "Observable.Finalize")
			// remove and unsubscribe from all merged
			if len(id.merged) != 0 {
				for merge := range id.merged {
					delete(id.merged, merge)
					merge.Unsubscribe <- id.Subscription
				}
			}
			id.Finalize <- true
			id.Yield()
			close(id.Finalize)
			id.complete()
			close(id.Connect)
			close(id.Subscribe)
			close(id.Unsubscribe)
			id.observers = nil
			id.operators = nil
		}()
		wg.Done()
		for {
			select {
			case event := <-id.Next:
				dlog.Println(id.UID, "Observable<-Next")
				id.onNext(event)
				break
			case err := <-id.Error:
				dlog.Println(id.UID, "Observable<-Error")
				if id.catchErrorFn != nil {
					id.catchErrorFn(err)
					id.Complete <- true
					break
				}
				if id.onResubscribe(err) {
					return
				}
				id.onError(err)
				return
			case <-id.Complete:
				dlog.Println(id.UID, "Observable<-Complete")
				if len(id.merged) != 0 {
					dlog.Println(id.UID, "Observable<-Unsubscribe error blocked by active merge")
					break
				}
				if id.onResubscribe(nil) {
					return
				}
				id.onComplete()
				return
			case observer := <-id.Subscribe:
				dlog.Println(id.UID, "Observable<-Subscribe")
				id.onSubscribe(observer)
				break
			case observer := <-id.Unsubscribe:
				dlog.Println(id.UID, "Observable<-Unsubscribe")
				id.onUnsubscribe(observer)
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
		observer.next(event)
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
	if !id.multicast {
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
	observer.complete()
	if len(id.observers) == 0 {
		dlog.Println(id.UID, "Observable.Subscription->Complete")
		id.Subscription.Complete <- true
	}
}

// onResubscribe handler
func (id *Observable) onResubscribe(err error) bool {
	log.Println(id.UID, "Observable.onResubscribe")
	if id.resubscribeFn != nil {
		if err != nil {
			if id.retryWhenFn != nil && id.retryWhenFn() {
				id.Subscription = NewSubscription()
				id.resubscribeFn(id)
				return true
			}
		} else {
			if id.repeatWhenFn != nil && id.repeatWhenFn() {
				id.Subscription = NewSubscription()
				id.resubscribeFn(id)
				return true
			}
		}
	}
	return false
}

// Yield staggers sending to a blocked select in order to avoid
// situations where multiple-cases can proceed causing the select
// to pseudo-random rather that FIFO across the channels
func (id *Observable) Yield() *Observable {
	<-time.After(1 * time.Millisecond)
	return id
}

//
// Modifiers
//

// Multicast modifier
func (id *Observable) Multicast() *Observable {
	log.Println(id.UID, "Observable.Multicast")
	id.multicast = true
	return id
}

// setBehavior modifier
func (id *Observable) setBehavior(value interface{}) *Observable {
	log.Println(id.UID, "Observable.Behavior")
	id.buffer = NewCircularBuffer(1)
	id.buffer.Add(value)
	return id
}

// setReplay modifier
func (id *Observable) setReplay(bufferSize int) *Observable {
	log.Println(id.UID, "Observable.Replay")
	id.buffer = NewCircularBuffer(bufferSize)
	return id
}

// Resubscribe modifier
func (id *Observable) Resubscribe(fn func(*Observable)) *Observable {
	log.Println(id.UID, "Observable.Resubscriber")
	id.resubscribeFn = fn
	fn(id)
	return id
}

// RetryWhen modifier
func (id *Observable) RetryWhen(fn func() bool) *Observable {
	log.Println(id.UID, "Observable.RetryWhen")
	id.retryWhenFn = fn
	return id
}

// RepeatWhen modifier
func (id *Observable) RepeatWhen(fn func() bool) *Observable {
	log.Println(id.UID, "Observable.RepeatWhen")
	id.repeatWhenFn = fn
	return id
}

// CatchError modifier
func (id *Observable) CatchError(fn func(error)) *Observable {
	log.Println(id.UID, "Observable.CatchError")
	id.catchErrorFn = fn
	return id
}

//
// Operators
//

// Merge operator
func (id *Observable) Merge(merge *Observable) *Observable {
	log.Println(id.UID, "Observable.Merge")

	merge.Multicast()
	id.merged[merge] = merge
	merge.Pipe(id)

	return id
}

// Delay operator
// sleep allows the observable to yield to the go channel
func (id *Observable) Delay(ms time.Duration) *Observable {
	// log.Println(id.UID, "Observable.Delay")
	<-time.After(ms * time.Millisecond)
	return id
}
