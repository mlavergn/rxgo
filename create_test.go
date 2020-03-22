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
		case event := <-observer.Event:
			switch event.Type {
			case EventTypeNext:
				// t.Log("next", next.(int))
				if event.Next == nil {
					t.Fatalf("Unexpected next nil value")
					return
				}
				nextCnt++
				break
			case EventTypeError:
				// t.Log("error", errorCnt)
				errorCnt++
				break loop
			case EventTypeComplete:
				// t.Log("complete", completeCnt)
				completeCnt++
				break loop
			default:
				break
			}
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
