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
	*Observer
	observers     map[*Observer]*Observer
	publish       bool
	connect       chan bool
	connecters    map[*Observer]*Observer
	share         bool
	multicast     bool
	Subscribe     chan *Observer
	Unsubscribe   chan *Observer
	completeOnce  sync.Once
	Finalize      chan bool
	merges        map[*Observable]*Observable
	mergesMutex   sync.RWMutex
	buffer        *CircularBuffer
	operators     []operator
	repeatWhenFn  func() bool
	retryWhenFn   func() bool
	catchErrorFn  func(error)
	resubscribeFn func(*Observable) error
	takeFn        func() bool
}

// NewObservable init
func NewObservable() *Observable {
	log.Println("Observable.NewObservable")
	id := &Observable{
		Observer:      NewObserver(),
		observers:     make(map[*Observer]*Observer, 1),
		publish:       false,
		connect:       make(chan bool, 1),
		connecters:    map[*Observer]*Observer{},
		share:         false,
		multicast:     false,
		Subscribe:     make(chan *Observer, 1),
		Unsubscribe:   make(chan *Observer, 1),
		Finalize:      make(chan bool, 1),
		merges:        map[*Observable]*Observable{},
		buffer:        nil,
		operators:     []operator{},
		repeatWhenFn:  nil,
		retryWhenFn:   nil,
		catchErrorFn:  nil,
		resubscribeFn: nil,
		takeFn:        nil,
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
			id.complete(id)
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
				if id.onNext(event) {
					break
				}
				return
			case err := <-id.Error:
				dlog.Println(id.UID, "Observable<-Error", err)
				if id.onError(err) {
					break
				}
				return
			case observable := <-id.Complete:
				dlog.Println(id.UID, "Observable<-Complete")
				if id.onComplete(observable) {
					break
				}
				return
			case observer := <-id.Subscribe:
				dlog.Println(id.UID, "Observable<-Subscribe")
				id.onSubscribe(observer)
				break
			case observer := <-id.Unsubscribe:
				dlog.Println(id.UID, "Observable<-Unsubscribe")
				if id.onUnsubscribe(observer) {
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
func (id *Observable) onNext(event interface{}) bool {
	log.Println(id.UID, "Observable.onNext")

	// Operations
	for _, op := range id.operators {
		switch op.op {
		case operatorFilter:
			filterFn := op.fn.(func(interface{}) bool)
			if filterFn(event) != true {
				return true
			}
			break
		case operatorMap:
			mapFn := op.fn.(func(interface{}) interface{})
			mappedEvent := mapFn(event)
			if mappedEvent == nil {
				return true
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

	// Take
	if id.takeFn != nil && !id.takeFn() {
		dlog.Println(id.UID, "Take complete")
		id.Yield()
		id.onComplete(id)
		return false
	}

	return true
}

// onError handler
func (id *Observable) onError(err error) bool {
	fmt.Println(id.UID, "Observable.onError", err)

	if id.onResubscribe(err) {
		fmt.Println(id.UID, "Observable<-Error blocked by resubscribe")
		return true
	}
	if id.catchErrorFn != nil {
		fmt.Println(id.UID, "Observable<-Error blocked by catchError")
		id.catchErrorFn(err)
		return true
	}

	id.completeOnce.Do(func() {
		for _, observer := range id.observers {
			delete(id.observers, observer)
			observer.error(err)
		}

		// remove and unsubscribe from all merged
		if len(id.merges) != 0 {
			id.mergesMutex.Lock()
			for merge := range id.merges {
				delete(id.merges, merge)
				select {
				case merge.Unsubscribe <- id.Observer:
				default:
				}
			}
			id.mergesMutex.Unlock()
		}
	})

	return false
}

// onComplete handler
func (id *Observable) onComplete(obs *Observable) bool {
	log.Println(id.UID, "Observable.onComplete")

	// if completion event came from this instance, don't interrupt it
	if id != obs {
		if obs != nil {
			obs.Unsubscribe <- id.Observer
		}
		if id.share {
			fmt.Println(id.UID, "Observable<-Complete blocked by active share")
			return true
		}
		if len(id.merges) != 0 && len(id.observers) != 0 {
			fmt.Println(id.UID, "Observable<-Complete blocked by active merge")
			return true
		}
		if id.onResubscribe(nil) {
			fmt.Println(id.UID, "Observable<-Complete blocked by resubscribe")
			return true
		}
	}

	id.completeOnce.Do(func() {
		for _, observer := range id.observers {
			delete(id.observers, observer)
			select {
			case observer.Complete <- id:
			default:
			}
		}

		// remove and unsubscribe from all merged
		if len(id.merges) != 0 {
			id.mergesMutex.Lock()
			for merge := range id.merges {
				delete(id.merges, merge)
				select {
				case merge.Unsubscribe <- id.Observer:
				default:
				}
			}
			id.mergesMutex.Unlock()
		}
		return
	})

	return false
}

// onSubscribe handler
func (id *Observable) onSubscribe(observer *Observer) bool {
	log.Println(id.UID, "Observable.onSubscribe")

	// if publish wait to fully subscribe
	if id.publish {
		id.connecters[observer] = observer
		return false
	}

	if !id.multicast {
		for _, observer := range id.observers {
			delete(id.observers, observer)
			observer.Complete <- id
		}
	}
	id.observers[observer] = observer

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

	return false
}

// onUnsubscribe handler
func (id *Observable) onUnsubscribe(observer *Observer) bool {
	log.Println(id.UID, "Observable.onUnsubscribe")
	delete(id.observers, observer)
	if len(id.observers) > 0 || id.share {
		return true
	}
	return false
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
		closed := id.Observer
		id.Observer = NewObserver()
		closed.complete(id)
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

// Merge operator
func (id *Observable) Merge(merge *Observable) *Observable {
	log.Println(id.UID, "Observable.Merge")
	merge.Multicast()
	id.mergesMutex.Lock()
	id.merges[merge] = merge
	id.mergesMutex.Unlock()
	merge.Pipe(id)
	return id
}
