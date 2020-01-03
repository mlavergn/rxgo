package rx

import (
	"log"
)

// Filter export
func (id *Observable) Filter(fn func(interface{}) bool) *Observable {
	log.Println("rx.Observable.Filter")
	id.filtered = fn
	return id
}

// Map export
func (id *Observable) Map(fn func(interface{}) interface{}) *Observable {
	log.Println("rx.Observable.Map")
	id.mapped = fn
	return id
}
