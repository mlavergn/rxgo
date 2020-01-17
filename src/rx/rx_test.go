package rx

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	log.Println("Testing RxGo", Version)
	// setup
	code := m.Run()
	// teardown
	os.Exit(code)
}

func TestMerge(t *testing.T) {
	events := 10

	nextCnt := 0
	errorCnt := 0
	completeCnt := 0

	subscription := NewSubscription()
	subscription.Take(events)
	subscription.UID = "testSubscription"

	intervalA := NewInterval(2)
	intervalA.UID = "testIntervalA"
	intervalB := NewInterval(2)
	intervalB.UID = "testIntervalB"
	intervalC := NewInterval(2)
	intervalC.UID = "testIntervalC"
	intervalD := NewInterval(2)
	intervalD.UID = "testIntervalD"

	subject := NewSubject()
	subject.UID = "testSubject"
	subject.Publish()
	subject.Merge(intervalA)
	subject.Merge(intervalB)
	subject.Merge(intervalC)
	subject.Merge(intervalD)
	subject.Connect()

	if len(subject.merges) != 4 {
		t.Fatalf("Expected merge count of %v but got %v", 4, len(subject.merges))
	}

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

	<-subject.Finalize

	if len(subject.observers) != 0 {
		t.Fatalf("Expected observer count of %v but got %v", 0, len(subject.observers))
	}

	if len(subject.merges) != 0 {
		t.Fatalf("Expected merge count of %v but got %v", 0, len(subject.merges))
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
