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
	id.operators = append(id.operators, operator{operatorFilter, fn})
	return id
}

// Map modifies the event type
func (id *Observable) Map(fn func(interface{}) interface{}) *Observable {
	log.Println(id.UID, "Observable.Map")
	id.operators = append(id.operators, operator{operatorMap, fn})
	return id
}

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

// Delay operator
// sleep allows the observable to yield to the go channel
func (id *Observable) Delay(ms time.Duration) *Observable {
	log.Println(id.UID, "Observable.Delay")
	<-time.After(ms * time.Millisecond)
	return id
}
