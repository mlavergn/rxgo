package rx

import (
	"fmt"
	"time"
)

// Pipe forwards events to the observer
func (id *Observable) Pipe(observer *Observable) *Observable {
	log.Println(id.UID, "Observable.Pipe")
	id.Subscribe <- observer.Observer
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
