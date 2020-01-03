package rx

import (
	"time"
)

// -----------------------------------------------------------------------------
// Interval

// NewInterval init
func NewInterval(msec int) *Observable {
	log.Println("Interval.NewInterval")
	id := NewObservable()

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
