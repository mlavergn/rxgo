package rx

import (
	"sync"
	"time"
)

// NewInterval init
func NewInterval(msec int) *Observable {
	log.Println("Interval.NewInterval")
	id := NewObservable()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		wg.Done()
		ticker := time.NewTicker(time.Duration(msec) * time.Millisecond)
		defer func() {
			ticker.Stop()
		}()

		// wait for connect
		<-id.connect

		i := 0
		for {
			select {
			case <-ticker.C:
				id.next(i)
				i++
				break
			case <-id.Finalize:
				return
			}
		}
	}()

	wg.Wait()
	return id
}

// NewFrom init
func NewFrom(values []interface{}) *Observable {
	log.Println("Interval.NewFrom")
	id := NewObservable()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		wg.Done()
		// wait for connect
		<-id.connect

		for value := range values {
			id.Next <- value
		}
		id.Complete <- true
	}()

	wg.Wait()
	return id
}
