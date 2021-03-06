package rx

import (
	"testing"
)

func TestReplay(t *testing.T) {
	events := 3

	nextCnt := 0
	errorCnt := 0
	completeCnt := 0

	observer := NewObserver()
	subject := NewReplaySubject(events)
	subject.Event <- Event{Type: EventTypeNext, Next: 1}
	subject.Event <- Event{Type: EventTypeNext, Next: 2}
	subject.Event <- Event{Type: EventTypeNext, Next: 3}
	subject.Event <- Event{Type: EventTypeNext, Next: 4}
	subject.Event <- Event{Type: EventTypeNext, Next: 5}
	subject.Event <- Event{Type: EventTypeNext, Next: 6}
	subject.Subscribe <- observer
loop:
	for {
		select {
		case event := <-observer.Event:
			switch event.Type {
			case EventTypeNext:
				// t.Log("next", event.Next.(int))
				if event.Next == nil {
					t.Fatalf("Unexpected next nil value")
					return
				}
				nextCnt++
				if nextCnt == events {
					observer.Event <- Event{Type: EventTypeComplete, Complete: subject}
				}
				break
			case EventTypeError:
				// t.Log("error", errorCnt)
				errorCnt++
				break loop
			case EventTypeComplete:
				// t.Log("complete", completeCnt)
				completeCnt++
				break loop
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
