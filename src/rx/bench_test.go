package rx

import (
	"testing"
)

func RxIntervalBench(events int) {
	subscription := NewSubscription()
	interval := NewInterval(50)
	interval.Take(events).Subscribe <- subscription
loop:
	for {
		select {
		case <-subscription.Next:
			break
		case <-subscription.Error:
			break loop
		case <-subscription.Complete:
			break loop
		}
	}

}

func BenchmarkRxInterval(b *testing.B) {
	for n := 0; n < b.N; n++ {
		RxIntervalBench(20)
	}
}
