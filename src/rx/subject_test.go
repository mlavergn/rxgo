package rx

import (
	"testing"
)

func TestReplay(t *testing.T) {
	events := 3

	nextCnt := 0
	errorCnt := 0
	completeCnt := 0

	subscription := NewSubscription()
	subject := NewReplaySubject(events)
	subject.Next <- 1
	subject.Next <- 2
	subject.Next <- 3
	subject.Next <- 4
	subject.Next <- 5
	subject.Next <- 6
	subject.Subscribe <- subscription
loop:
	for {
		select {
		case next := <-subscription.Next:
			// t.Log("next", next.(int))
			if next == nil {
				t.Fatalf("Unexpected next nil value")
				return
			}
			nextCnt++
			if nextCnt == events {
				subscription.Complete <- subject
			}
			break
		case <-subscription.Error:
			// t.Log("error", errorCnt)
			errorCnt++
			break loop
		case <-subscription.Complete:
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
