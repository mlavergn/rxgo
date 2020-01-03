package rxgo

import (
	"log"
	"time"
)

// -----------------------------------------------------------------------------
// RxInterval

// NewRxInterval init
func NewRxInterval(msec int) *RxObservable {
	log.Println("RxInterval::NewRxInterval")
	id := NewRxObservable()

	go func() {
		ticker := time.NewTicker(time.Duration(msec) * time.Millisecond)
		defer func() {
			ticker.Stop()
		}()
		i := 0
		for {
			select {
			case <-ticker.C:
				id.next(i)
				i++
				break
			case <-id.close:
				return
			}
		}
	}()

	return id
}
