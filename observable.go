package rx

import (
	"sync"
	"time"
)

type operatorType int

const (
	operatorMap operatorType = iota
	operatorFilter
	operatorTap
	operatorStartWith
)

type operator struct {
	op operatorType
	fn interface{}
}

// Observable type
type Observable struct {
	*Observer
	pipes          map[*Observable]*Observable
	pipesMutex     sync.RWMutex
	observers      map[*Observer]*Observer
	observersMutex sync.RWMutex
	publish        bool
	connect        chan bool
	connecters     map[*Observer]*Observer
	share          bool
	multicast      bool
	merge          bool
	Subscribe      chan *Observer
	subscribeOps   []operator
	Unsubscribe    chan *Observer
	completeOnce   sync.Once
	Finalize       chan bool
	buffer         *CircularBuffer
	nextOps        []operator
	repeatWhenFn   func() bool
	retryWhenFn    func() bool
	catchErrorFn   func(error)
	resubscribeFn  func(*Observable) error
	takeFn         func() bool
}

// NewObservable init
func NewObservable() *Observable {
	log.Println("Observable.NewObservable")
	id := &Observable{
		Observer:      NewObserver(),
		pipes:         map[*Observable]*Observable{},
		observers:     make(map[*Observer]*Observer, 1),
		publish:       false,
		connect:       make(chan bool, 1),
		connecters:    map[*Observer]*Observer{},
		share:         false,
		multicast:     false,
		merge:         false,
		Subscribe:     make(chan *Observer, 1),
		subscribeOps:  []operator{},
		Unsubscribe:   make(chan *Observer, 1),
		Finalize:      make(chan bool, 1),
		buffer:        nil,
		nextOps:       []operator{},
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
			id.complete(id) // cleanup the Observer member
			close(id.connect)
			close(id.Subscribe)
			close(id.Unsubscribe)
			id.observers = nil
			id.nextOps = nil
		}()
		wg.Done()
		for {
			select {
			case event := <-id.Event:
				switch event.Type {
				case EventTypeNext:
					dlog.Println(id.UID, "Observable<-Next")
					if id.onNext(event.Next) {
						break
					}
					return
				case EventTypeError:
					dlog.Println(id.UID, "Observable<-Error", event.Error)
					if id.onError(event.Error) {
						break
					}
					return
				case EventTypeComplete:
					dlog.Println(id.UID, "Observable<-Complete")
					if id.onComplete(event.Complete) {
						break
					}
					return
				}
				break
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

func (id *Observable) doComplete(err error) bool {
	id.completeOnce.Do(func() {
		id.observersMutex.Lock()
		for _, observer := range id.observers {
			delete(id.observers, observer)
			if err != nil {
				select {
				case observer.Event <- Event{Type: EventTypeError, Error: err}:
				default:
				}
			} else {
				select {
				case observer.Event <- Event{Type: EventTypeComplete, Complete: id}:
				default:
				}
			}
		}
		id.observersMutex.Unlock()

		id.clearPipes()
	})

	return true
}

func (id *Observable) addPipe(pipe *Observable) {
	id.pipesMutex.Lock()
	id.pipes[pipe] = pipe
	id.pipesMutex.Unlock()
}

func (id *Observable) delPipe(pipe *Observable, unsubscribe bool) {
	id.pipesMutex.Lock()
	delete(id.pipes, pipe)
	id.pipesMutex.Unlock()
	if unsubscribe {
		select {
		case pipe.Unsubscribe <- id.Observer:
		default:
		}
	}
}

func (id *Observable) clearPipes() {
	// remove and unsubscribe from all merged
	if len(id.pipes) != 0 {
		for pipe := range id.pipes {
			id.delPipe(pipe, true)
		}
	}
}

// onNext handler
func (id *Observable) onNext(event interface{}) bool {
	log.Println(id.UID, "Observable.onNext")

	// Operations
	for _, op := range id.nextOps {
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
		case operatorTap:
			tapFn := op.fn.(func(interface{}))
			tapFn(event)
			break
		}
	}

	// Replay / Distinct
	if id.buffer != nil {
		id.buffer.Add(event)
	}

	// multicast the event
	id.observersMutex.RLock()
	for _, observer := range id.observers {
		observer.next(event)
	}
	id.observersMutex.RUnlock()

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
	log.Println(id.UID, "Observable.onError", err)

	if id.onResubscribe(err) {
		log.Println(id.UID, "Observable<-Error blocked by resubscribe")
		return true
	}
	if id.catchErrorFn != nil {
		log.Println(id.UID, "Observable<-Error blocked by catchError")
		id.catchErrorFn(err)
		return true
	}

	id.doComplete(err)

	return false
}

// onComplete handler
func (id *Observable) onComplete(obs *Observable) bool {
	log.Println(id.UID, "Observable.onComplete")

	// if completion event came from this instance, don't interrupt it
	if id != obs {
		if id.onResubscribe(nil) {
			log.Println(id.UID, "Observable<-Complete blocked by resubscribe")
			return true
		}
		if obs != nil {
			id.delPipe(obs, false)
		}
		if id.share {
			log.Println(id.UID, "Observable<-Complete blocked by active share")
			return true
		}
		if id.merge && len(id.pipes) != 0 && len(id.observers) != 0 {
			log.Println(id.UID, "Observable<-Complete blocked by active merge")
			return true
		}
	}

	id.doComplete(nil)

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

	id.observersMutex.Lock()
	if !id.multicast {
		for _, observer := range id.observers {
			delete(id.observers, observer)
			observer.Event <- Event{Type: EventTypeComplete, Complete: id}
		}
	}
	id.observers[observer] = observer
	id.observersMutex.Unlock()

	// subscription operations
	for _, op := range id.subscribeOps {
		switch op.op {
		case operatorStartWith:
			fn := op.fn.(func() []interface{})
			events := fn()
			for _, event := range events {
				observer.next(event)
			}
			break
		}
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
	id.observersMutex.Lock()
	delete(id.observers, observer)
	id.observersMutex.Unlock()
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
		id.clearPipes()
		oldObserver := id.Observer
		id.Observer = NewObserver()
		oldObserver.complete(id)
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
	log.Println(id.UID, "Observable.Replay", bufferSize)
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
