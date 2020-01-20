package rx

import (
	"fmt"
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
	publish       bool
	connect       chan bool
	connecters    map[*Subscription]*Subscription
	share         bool
	multicast     bool
	Subscribe     chan *Subscription
	Unsubscribe   chan *Subscription
	Finalize      chan bool
	merges        map[*Observable]*Observable
	buffer        *CircularBuffer
	operators     []operator
	repeatWhenFn  func() bool
	retryWhenFn   func() bool
	catchErrorFn  func(error)
	resubscribeFn func(*Observable) error
}

// NewObservable init
func NewObservable() *Observable {
	log.Println("Observable.NewObservable")
	id := &Observable{
		Subscription:  NewSubscription(),
		observers:     make(map[*Subscription]*Subscription, 1),
		publish:       false,
		connect:       make(chan bool, 1),
		connecters:    map[*Subscription]*Subscription{},
		share:         false,
		multicast:     false,
		Subscribe:     make(chan *Subscription, 1),
		Unsubscribe:   make(chan *Subscription, 1),
		Finalize:      make(chan bool, 1),
		merges:        map[*Observable]*Observable{},
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
			id.Finalize <- true
			id.Yield()
			close(id.Finalize)
			id.complete()
			close(id.connect)
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
				if id.onResubscribe(err) {
					log.Println(id.UID, "Observable<-Error blocked by resubscribe")
					break
				}
				if id.catchErrorFn != nil {
					log.Println(id.UID, "Observable<-Error blocked by catchError")
					id.catchErrorFn(err)
					break
				}
				id.onError(err)
				return
			case <-id.Complete:
				dlog.Println(id.UID, "Observable<-Complete")
				if len(id.merges) != 0 && (len(id.observers) != 0 || id.share) {
					log.Println(id.UID, "Observable<-Complete blocked by active merge")
					break
				}
				if id.onResubscribe(nil) {
					log.Println(id.UID, "Observable<-Complete blocked by resubscribe")
					break
				}
				id.onComplete()
				return
			case observer := <-id.Subscribe:
				dlog.Println(id.UID, "Observable<-Subscribe")

				if id.publish {
					id.connecters[observer] = observer
					break
				}

				id.onSubscribe(observer)
				break
			case observer := <-id.Unsubscribe:
				dlog.Println(id.UID, "Observable<-Unsubscribe")
				id.onUnsubscribe(observer)
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

	// Operations
	for _, op := range id.operators {
		switch op.op {
		case operatorFilter:
			filterFn := op.fn.(func(interface{}) bool)
			if filterFn(event) != true {
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

	// remove and unsubscribe from all merged
	if len(id.merges) != 0 {
		for merge := range id.merges {
			delete(id.merges, merge)
			merge.Unsubscribe <- id.Subscription
		}
	}
}

// onComplete handler
func (id *Observable) onComplete() {
	log.Println(id.UID, "Observable.onComplete")
	for _, observer := range id.observers {
		delete(id.observers, observer)
		observer.complete()
	}

	// remove and unsubscribe from all merged
	if len(id.merges) != 0 {
		for merge := range id.merges {
			delete(id.merges, merge)
			merge.Unsubscribe <- id.Subscription
		}
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

	// if no publish hold, send Connect
	if id.publish == false {
		select {
		case id.connect <- true:
			break
		default:
			break
		}
	}
}

// onUnsubscribe handler
func (id *Observable) onUnsubscribe(observer *Subscription) {
	log.Println(id.UID, "Observable.onUnsubscribe")
	observer.observable = nil
	delete(id.observers, observer)
	observer.complete()
	if len(id.observers) == 0 && !id.share {
		id.Subscription.Complete <- true
	}
}

// onResubscribe handler
func (id *Observable) onResubscribe(err error) bool {
	log.Println(id.UID, "Observable.onResubscribe")
	// if all observers are gone, do not resubscribe
	if len(id.observers) == 0 && !id.share {
		log.Println(id.UID, "Observable.onResubscribe no observers remaining")
		return false
	}
	if id.resubscribeFn != nil {
		dlog.Println(id.UID, "Observable.onResubscribe.resubscribeFn")
		closed := id.Subscription
		id.Subscription = NewSubscription()
		closed.complete()
		if err != nil {
			dlog.Println(id.UID, "Observable.onResubscribe.retryWhen")
			if id.retryWhenFn != nil && id.retryWhenFn() {
				id.resubscribeFn(id)
				return true
			}
		} else {
			dlog.Println(id.UID, "Observable.onResubscribe.repeatWhen")
			if id.repeatWhenFn != nil && id.repeatWhenFn() {
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

// Share modifier
func (id *Observable) Share() *Observable {
	id.share = true
	return id
}

// Publish modifier
func (id *Observable) Publish() *Observable {
	id.publish = true
	return id
}

// Connect modifier
func (id *Observable) Connect() *Observable {
	for o := range id.connecters {
		id.onSubscribe(o)
	}
	id.publish = false
	select {
	case id.connect <- true:
		break
	default:
		break
	}

	return id
}

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
func (id *Observable) Resubscribe(fn func(*Observable) error) *Observable {
	log.Println(id.UID, "Observable.Resubscribe", fn != nil)
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

// Distinct operator
func (id *Observable) Distinct() *Observable {
	last := ""
	id.Filter(func(event interface{}) bool {
		eventStr := fmt.Sprint(event)
		if eventStr != last {
			last = eventStr
			return true
		}
		return false
	})
	return id
}

// Merge operator
func (id *Observable) Merge(merge *Observable) *Observable {
	log.Println(id.UID, "Observable.Merge")
	merge.Multicast()
	// TODO not thread safe
	id.merges[merge] = merge
	merge.Pipe(id)
	return id
}

// Delay operator
// sleep allows the observable to yield to the go channel
func (id *Observable) Delay(ms time.Duration) *Observable {
	log.Println(id.UID, "Observable.Delay")
	<-time.After(ms * time.Millisecond)
	return id
}
