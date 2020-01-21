package rx

import (
	"testing"
)

func TestInterval(t *testing.T) {
	events := 10

	nextCnt := 0
	errorCnt := 0
	completeCnt := 0

	observer := NewObserver()
	interval := NewInterval(50)
	interval.Take(events).Subscribe <- observer
loop:
	for {
		select {
		case next := <-observer.Next:
			// t.Log("next", next.(int))
			if next == nil {
				t.Fatalf("Unexpected next nil value")
				return
			}
			nextCnt++
			break
		case <-observer.Error:
			// t.Log("error", errorCnt)
			errorCnt++
			break loop
		case <-observer.Complete:
			// t.Log("complete", completeCnt)
			completeCnt++
			break loop
		}
	}

	if nextCnt != events {
		t.Fatalf("Expected next count of %v but got %v", events, nextCnt)
	}
	if errorCnt != 0 {
		t.Fatalf("Expected error count of %v but got %v", 0, errorCnt)
	}
	if completeCnt != 1 {
		t.Fatalf("Expected complete count of %v but got %v", 1, completeCnt)
	}
}
