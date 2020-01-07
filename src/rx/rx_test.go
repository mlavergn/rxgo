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

func TestInterval(t *testing.T) {
	events := 10

	nextCnt := 0
	errorCnt := 0
	completeCnt := 0

	subscription := NewSubscription()
	subscription.Take(events)
	interval := NewInterval(50)
	interval.Subscribe <- subscription
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
	subject.Delay(1)
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
				subscription.Complete <- true
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

func TestFilter(t *testing.T) {
	events := 5

	nextCnt := 0
	errorCnt := 0
	completeCnt := 0

	subscription := NewSubscription()
	subscription.Take(events)
	interval := NewInterval(10)
	interval.Filter(func(value interface{}) bool {
		return (ToInt(value, -1)%2 == 0)
	})
	interval.Subscribe <- subscription
loop:
	for {
		select {
		case next := <-subscription.Next:
			if next == nil {
				t.Fatalf("Unexpected next nil value")
				return
			}
			// t.Log("next", next.(int))
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

	subscription := NewSubscription()
	subscription.Take(events)
	interval := NewInterval(5)
	interval.Map(func(value interface{}) interface{} {
		return ToInt(value, 0) * 10
	})
	interval.Subscribe <- subscription
loop:
	for {
		select {
		case next := <-subscription.Next:
			value := ToInt(next, 0)
			if next == nil || value%10 != 0 {
				t.Fatalf("Unexpected next value %v", next)
				return
			}
			// t.Log("next", next.(int))
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

func TestMerge(t *testing.T) {
	events := 10

	nextCnt := 0
	errorCnt := 0
	completeCnt := 0

	subscription := NewSubscription()
	subscription.Take(events)
	subscription.UID = 123

	intervalA := NewInterval(2)
	intervalB := NewInterval(2)
	intervalC := NewInterval(2)
	intervalD := NewInterval(2)

	subject := NewSubject()
	subject.Merge(intervalA)
	subject.Merge(intervalB)
	subject.Merge(intervalC)
	subject.Merge(intervalD)

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
