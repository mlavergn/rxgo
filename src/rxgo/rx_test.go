package rxgo

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	RxSetup(true)
	code := m.Run()
	// teardown
	os.Exit(code)
}

func TestRxInterval(t *testing.T) {
	events := 10

	nextCnt := events
	errorCnt := 0
	completeCnt := 1

	observer := NewRxObserver()
	interval := NewRxInterval(50)
	interval.Take(events)
	interval.Subscribe <- observer
loop:
	for {
		select {
		case next := <-observer.OnNext:
			t.Log("next", next.(int))
			nextCnt--
			break
		case <-observer.OnError:
			t.Log("error", errorCnt)
			errorCnt--
			break loop
		case <-observer.OnComplete:
			t.Log("complete", completeCnt)
			completeCnt--
			break loop
		}
	}

	if nextCnt != 0 {
		t.Fatalf("Expected next count of %v but got %v", 0, nextCnt)
	}
	if errorCnt != 0 {
		t.Fatalf("Expected error count of %v but got %v", 0, errorCnt)
	}
	if completeCnt != 0 {
		t.Fatalf("Expected complete count of %v but got %v", 0, completeCnt)
	}
}

func TestRxFilter(t *testing.T) {
	events := 5

	nextCnt := events
	errorCnt := 0
	completeCnt := 1

	observer := NewRxObserver()
	interval := NewRxInterval(10)
	interval.Take(events)
	interval.Filter(func(value interface{}) bool {
		return (ToInt(value, -1)%2 == 0)
	})
	interval.Subscribe <- observer
loop:
	for {
		select {
		case next := <-observer.OnNext:
			t.Log("next", next.(int))
			nextCnt--
			break
		case <-observer.OnError:
			t.Log("error", errorCnt)
			errorCnt--
			break loop
		case <-observer.OnComplete:
			t.Log("complete", completeCnt)
			completeCnt--
			break loop
		}
	}

	if nextCnt != 0 {
		t.Fatalf("Expected next count of %v but got %v", 0, nextCnt)
	}
	if errorCnt != 0 {
		t.Fatalf("Expected error count of %v but got %v", 0, errorCnt)
	}
	if completeCnt != 0 {
		t.Fatalf("Expected complete count of %v but got %v", 0, completeCnt)
	}
}

func RxIntervalBench(events int) {
	observer := NewRxObserver()
	interval := NewRxInterval(50)
	interval.Take(events)
	interval.Subscribe <- observer
loop:
	for {
		select {
		case <-observer.OnNext:
			break
		case <-observer.OnError:
			break loop
		case <-observer.OnComplete:
			break loop
		}
	}

}

func BenchmarkRxInterval(b *testing.B) {
	for n := 0; n < b.N; n++ {
		RxIntervalBench(20)
	}
}
