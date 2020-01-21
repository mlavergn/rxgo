package rx

import (
	"testing"
)

func RxIntervalBench(events int) {
	observer := NewObserver()
	interval := NewInterval(50)
	interval.Take(events).Subscribe <- observer
loop:
	for {
		select {
		case <-observer.Next:
			break
		case <-observer.Error:
			break loop
		case <-observer.Complete:
			break loop
		}
	}

}

func BenchmarkRxInterval(b *testing.B) {
	for n := 0; n < b.N; n++ {
		RxIntervalBench(20)
	}
}
