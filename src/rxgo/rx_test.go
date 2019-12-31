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

	obvr := NewRxObserver()
	obvb := NewRxInterval(50)
	obvb.Take(events)
	obvb.Subscribe <- obvr
loop:
	for {
		select {
		case next := <-obvr.Next:
			t.Log("next", next.(int))
			nextCnt--
			break
		case <-obvr.Error:
			t.Log("error", errorCnt)
			errorCnt--
			break loop
		case <-obvr.Complete:
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
	obvr := NewRxObserver()
	obvb := NewRxInterval(50)
	obvb.Take(events)
	obvb.Subscribe <- obvr
loop:
	for {
		select {
		case <-obvr.Next:
			break
		case <-obvr.Error:
			break loop
		case <-obvr.Complete:
			break loop
		}
	}

}

func BenchmarkRxInterval(b *testing.B) {
	for n := 0; n < b.N; n++ {
		RxIntervalBench(20)
	}
}
