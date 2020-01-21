package rx

import (
	"testing"
	"time"
)

func TestRequestText(t *testing.T) {
	rxhttp := NewHTTPRequest(10 * time.Second)
	subject, err := rxhttp.TextSubject("http://httpbin.org/get", nil)

	if err != nil {
		t.Fatalf("Init error %v", err)
		return
	}

	completeCnt := 0

	observer := NewObserver()
	subject.Subscribe <- observer
loop:
	for {

		select {
		case event := <-observer.Next:
			data := ToString(event, "")
			if len(data) < 1 {
				t.Fatalf("Next invalid length for %v", data)
			}
			break
		case err := <-observer.Error:
			t.Fatalf("Error %v", err)
			return
		case <-observer.Complete:
			completeCnt++
			break loop
		}
	}

	if completeCnt != 1 {
		t.Fatalf("Expected complete count of %v but got %v", 1, completeCnt)
	}
}

func TestRequestLine(t *testing.T) {
	expect := 11
	actual := 0
	rxhttp := NewHTTPRequest(10 * time.Second)
	subject, err := rxhttp.LineSubject("http://httpbin.org/get", nil)

	if err != nil {
		t.Fatalf("Init error %v", err)
		return
	}

	completeCnt := 0
	lines := []string{}

	observer := NewObserver()
	subject.Subscribe <- observer
loop:
	for {
		select {
		case event := <-observer.Next:
			actual++
			data := event.([]byte)
			if len(data) < 1 {
				t.Fatalf("Next invalid length for data %v", data)
			}
			lines = append(lines, string(data))
			break
		case err := <-observer.Error:
			t.Fatalf("Error %v", err)
			break loop
		case <-observer.Complete:
			completeCnt++
			break loop
		}
	}

	if completeCnt != 1 {
		t.Fatalf("Expected complete count of %v but got %v", 1, completeCnt)
	}

	if actual != expect {
		t.Fatalf("xExpected %v lines got %v", expect, actual)
		t.Log(lines)
		return
	}
}
func TestRequestJSON(t *testing.T) {
	rxhttp := NewHTTPRequest(10 * time.Second)
	subject, err := rxhttp.JSONSubject("http://httpbin.org/get", nil)

	if err != nil {
		t.Fatalf("Init error %v", err)
		return
	}

	completeCnt := 0

	observer := NewObserver()
	subject.Subscribe <- observer
loop:
	for {
		select {
		case event := <-observer.Next:
			data := ToStringMap(event, nil)
			if len(data) < 1 {
				t.Fatalf("Next invalid length for %v", data)
			}
			break
		case err := <-observer.Error:
			t.Fatalf("Error %v", err)
			return
		case <-observer.Complete:
			completeCnt++
			break loop
		}
	}

	if completeCnt != 1 {
		t.Fatalf("Expected complete count of %v but got %v", 1, completeCnt)
	}
}

func TestRequestSSE(t *testing.T) {
	rxhttp := NewHTTPRequest(10 * time.Second)
	subject, err := rxhttp.SSESubject("http://express-eventsource.herokuapp.com/events", nil)

	if err != nil {
		t.Fatalf("Init error %v", err)
		return
	}

	completeCnt := 0

	observer := NewObserver()
	subject.Take(1).Subscribe <- observer
loop:
	for {
		select {
		case event := <-observer.Next:
			data := ToStringMap(event, nil)
			if len(data) < 3 {
				t.Fatalf("Expected min 3 lines, but received %v", len(data))
				return
			}
			break
		case err := <-observer.Error:
			t.Fatalf("Error %v", err)
			return
		case <-observer.Complete:
			completeCnt++
			break loop
		}
	}

	if completeCnt != 1 {
		t.Fatalf("Expected complete count of %v but got %v", 1, completeCnt)
	}
}
