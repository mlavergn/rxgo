package rxgo

import (
	"log"
	"time"
)

// -----------------------------------------------------------------------------
// RxInterval

// RxInterval type
type RxInterval struct {
	*RxObservable
}

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
				id.Next(i)
				i++
				break
			case <-id.close:
				return
			}
		}
	}()

	return id
}
