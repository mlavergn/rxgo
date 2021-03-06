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

	observer := NewObserver()
	observer.UID = "testSubscription"

	intervalA := NewInterval(2)
	intervalA.UID = "testIntervalA." + intervalA.UID
	intervalB := NewInterval(2)
	intervalB.UID = "testIntervalB." + intervalB.UID
	intervalC := NewInterval(2)
	intervalC.UID = "testIntervalC." + intervalC.UID
	intervalD := NewInterval(2)
	intervalD.UID = "testIntervalD." + intervalD.UID

	subject := NewSubject()
	subject.UID = "testSubject." + subject.UID
	subject.Publish()
	subject.Merge(intervalA)
	subject.Merge(intervalB)
	subject.Merge(intervalC)
	subject.Merge(intervalD)
	subject.Connect()

	if len(subject.pipes) != 4 {
		t.Fatalf("Expected merge count of %v but got %v", 4, len(subject.pipes))
	}

	subject.Take(events).Subscribe <- observer
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
			}
		}
	}

	<-subject.Finalize

	if len(subject.observers) != 0 {
		t.Fatalf("Expected observer count of %v but got %v", 0, len(subject.observers))
	}

	if len(subject.pipes) != 0 {
		t.Fatalf("Expected merge count of %v but got %v", 0, len(subject.pipes))
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
