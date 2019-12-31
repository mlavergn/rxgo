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
func NewRxInterval(msec int) *RxInterval {
	log.Println("RxInterval::NewRxInterval")
	id := &RxInterval{
		RxObservable: NewRxObservable(),
	}

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
			case <-id.Close:
				return
			}
		}
	}()

	return id
}
