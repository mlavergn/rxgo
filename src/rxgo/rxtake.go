package rxgo

import (
	"log"
)

// Take export
func (id *RxObservable) Take(count int) {
	log.Println("RxObservable::Take")

	counter := count
	id.take = func() {
		log.Println("RxObservable::Take call")
		counter--
		if counter <= 0 {
			id.complete()
		}
	}
}

// TakeWhile export
func (id *RxObservable) TakeWhile(cond func() bool) {
	log.Println("RxObservable::TakeWhile")

	id.take = func() {
		log.Println("RxObservable::TakeWhile call")
		if cond() == false {
			id.complete()
		}
	}
}

// TakeUntil export
func (id *RxObservable) TakeUntil(event chan interface{}) {
	log.Println("RxObservable::TakeUntil")

	id.take = func() {
		log.Println("RxObservable::TakeUntil call")
		go func() {
			select {
			case <-event:
				id.complete()
				break
			}
		}()
	}
}
