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
		case event := <-observer.Event:
			switch event.Type {
			case EventTypeNext:
				break
			case EventTypeError:
				break loop
			case EventTypeComplete:
				break loop
			}
		}
	}

}

func BenchmarkRxInterval(b *testing.B) {
	for n := 0; n < b.N; n++ {
		RxIntervalBench(20)
	}
}
