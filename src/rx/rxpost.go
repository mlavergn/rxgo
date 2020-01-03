package rx

import (
	"log"
)

// Take export
func (id *Observable) Take(count int) {
	log.Println("rx.Observable.Take")

	counter := count
	id.take = func() {
		counter--
		if counter <= 0 {
			id.complete()
		}
	}
}

// TakeWhile export
func (id *Observable) TakeWhile(cond func() bool) {
	log.Println("rx.Observable.TakeWhile")

	id.take = func() {
		if cond() == false {
			id.complete()
		}
	}
}

// TakeUntil export
func (id *Observable) TakeUntil(event chan interface{}) {
	log.Println("rx.Observable.TakeUntil")

	id.take = func() {
		go func() {
			select {
			case <-event:
				id.complete()
				break
			}
		}()
	}
}
