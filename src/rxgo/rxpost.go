package rxgo

import (
	"log"
)

// Take export
func (id *RxObservable) Take(count int) {
	log.Println("RxObservable::Take")

	counter := count
	id.take = func() {
		counter--
		if counter <= 0 {
			id.Complete()
		}
	}
}

// TakeWhile export
func (id *RxObservable) TakeWhile(cond func() bool) {
	log.Println("RxObservable::TakeWhile")

	id.take = func() {
		if cond() == false {
			id.Complete()
		}
	}
}

// TakeUntil export
func (id *RxObservable) TakeUntil(event chan interface{}) {
	log.Println("RxObservable::TakeUntil")

	id.take = func() {
		go func() {
			select {
			case <-event:
				id.Complete()
				break
			}
		}()
	}
}
