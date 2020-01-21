package rx

import (
	"testing"
)

func TestFilter(t *testing.T) {
	events := 5

	nextCnt := 0
	errorCnt := 0
	completeCnt := 0

	observer := NewObserver()
	interval := NewInterval(10)
	interval.Filter(func(value interface{}) bool {
		return (ToInt(value, -1)%2 == 0)
	})
	interval.Take(events).Subscribe <- observer
loop:
	for {
		select {
		case next := <-observer.Next:
			if next == nil {
				t.Fatalf("Unexpected next nil value")
				return
			}
			// t.Log("next", next.(int))
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

func TestMap(t *testing.T) {
	events := 5

	nextCnt := 0
	errorCnt := 0
	completeCnt := 0

	observer := NewObserver()
	interval := NewInterval(5)
	interval.Map(func(value interface{}) interface{} {
		return ToInt(value, 0) * 10
	})
	interval.Take(events).Subscribe <- observer
loop:
	for {
		select {
		case next := <-observer.Next:
			value := ToInt(next, 0)
			if next == nil || value%10 != 0 {
				t.Fatalf("Unexpected next value %v", next)
				return
			}
			// t.Log("next", next.(int))
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
