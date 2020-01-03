package rx

import (
	"log"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	log.Println("Testing RxGo", Version)
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
		case next := <-observer.Next:
			// t.Log("next", next.(int))
			if next == nil {
				t.Fatalf("Unexpected next nil value")
			}
			nextCnt--
			break
		case <-observer.Error:
			// t.Log("error", errorCnt)
			errorCnt--
			break loop
		case <-observer.Complete:
			// t.Log("complete", completeCnt)
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

func TestRxReplay(t *testing.T) {
	events := 3

	nextCnt := events
	errorCnt := 0
	completeCnt := 1

	observer := NewRxObserver()
	subject := NewRxReplaySubject(events)
	subject.Next <- 1
	subject.Next <- 2
	subject.Next <- 3
	subject.Next <- 4
	subject.Next <- 5
	subject.Next <- 6
	subject.Warmup()
	subject.Subscribe <- observer
loop:
	for {
		select {
		case next := <-observer.Next:
			// t.Log("next", next.(int))
			if next == nil {
				t.Fatalf("Unexpected next nil value")
			}
			nextCnt--
			if nextCnt == 0 {
				observer.Complete <- true
			}
			break
		case <-observer.Error:
			// t.Log("error", errorCnt)
			errorCnt--
			break loop
		case <-observer.Complete:
			// t.Log("complete", completeCnt)
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
		case next := <-observer.Next:
			if next == nil {
				t.Fatalf("Unexpected next nil value")
			}
			// t.Log("next", next.(int))
			nextCnt--
			break
		case <-observer.Error:
			// t.Log("error", errorCnt)
			errorCnt--
			break loop
		case <-observer.Complete:
			// t.Log("complete", completeCnt)
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
