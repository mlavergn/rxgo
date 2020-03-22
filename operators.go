package rx

import (
	"fmt"
	"time"
)

// Pipe pipes events to an observerable
func (id *Observable) Pipe(observable *Observable) *Observable {
	log.Println(id.UID, "Observable.Pipe")
	observable.addPipe(id)
	id.Subscribe <- observable.Observer
	return id
}

// Merge operator adds observable to id's event output
func (id *Observable) Merge(observable *Observable) *Observable {
	log.Println(id.UID, "Observable.Merge")
	id.merge = true
	observable.Pipe(id)
	return id
}

// Filter export
// Emit values that PASS (return true) for the filter condition
func (id *Observable) Filter(fn func(interface{}) bool) *Observable {
	log.Println(id.UID, "Observable.Filter")
	id.nextOps = append(id.nextOps, operator{operatorFilter, fn})
	return id
}

// Map modifies the event type
func (id *Observable) Map(fn func(interface{}) interface{}) *Observable {
	log.Println(id.UID, "Observable.Map")
	id.nextOps = append(id.nextOps, operator{operatorMap, fn})
	return id
}

// Tap adds a side effect to an event
func (id *Observable) Tap(fn func(interface{})) *Observable {
	log.Println(id.UID, "Observable.Tap")
	id.nextOps = append(id.nextOps, operator{operatorTap, fn})
	return id
}

// Distinct operator
func (id *Observable) Distinct() *Observable {
	log.Println(id.UID, "Observable.Distinct")
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

// Delay operator
func (id *Observable) Delay(delay time.Duration) *Observable {
	log.Println(id.UID, "Observable.Delay")
	id.Tap(func(event interface{}) {
		<-time.After(delay)
	})
	return id
}

// StartWith operator
// NOTE: Processes events on each subscription, including shared observables
func (id *Observable) StartWith(fn func() []interface{}) *Observable {
	log.Println(id.UID, "Observable.StartWith")
	id.subscribeOps = append(id.subscribeOps, operator{operatorStartWith, fn})
	return id
}
