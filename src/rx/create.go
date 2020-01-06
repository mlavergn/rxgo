package rx

import (
	"time"
)

// NewInterval init
func NewInterval(msec int) *Observable {
	log.Println("Interval.NewInterval")
	id := NewObservable()

	go func() {
		ticker := time.NewTicker(time.Duration(msec) * time.Millisecond)
		defer func() {
			ticker.Stop()
		}()

		// wait for connect
		<-id.Connect

		i := 0
		for {
			select {
			case <-ticker.C:
				id.next(i)
				i++
				break
			case <-id.finally:
				return
			}
		}
	}()

	return id
}

// NewFrom init
func NewFrom(values []interface{}) *Observable {
	log.Println("Interval.NewFrom")
	id := NewObservable()

	go func() {
		// wait for connect
		<-id.Connect

		for value := range values {
			id.Next <- value
		}
		id.Complete <- true
	}()

	return id
}
