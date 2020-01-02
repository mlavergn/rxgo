package rxgo

import (
	"log"
)

// Filter export
func (id *RxObservable) Filter(cond func(interface{}) bool) {
	log.Println("RxObservable::Filter")

	id.filter = func(event interface{}) bool {
		log.Println("RxObservable::Filter call")
		return cond(event)
	}
}
