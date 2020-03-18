package rx

import (
	"strings"
	"testing"
)

func TestRequestText(t *testing.T) {
	subject, err := NewHTTPTextSubject("http://httpbin.org/get", nil)

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
		case event := <-observer.Event:
			switch event.Type {
			case EventTypeNext:
				data := ToString(event.Next, "")
				if len(data) < 1 {
					t.Fatalf("Next invalid length for %v", data)
				}
			case EventTypeError:
				t.Fatalf("Error %v", event.Error)
				return
			case EventTypeComplete:
				completeCnt++
				break loop
			}
		}
	}

	if completeCnt != 1 {
		t.Fatalf("Expected complete count of %v but got %v", 1, completeCnt)
	}
}

func TestRequestLine(t *testing.T) {
	expect := 10
	actual := 0
	subject, err := NewHTTPLineSubject("http://httpbin.org/get", nil)

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
		case event := <-observer.Event:
			switch event.Type {
			case EventTypeNext:
				actual++
				data := event.Next.([]byte)
				if len(data) < 1 {
					t.Fatalf("Next invalid length for data %v", data)
				}
				lines = append(lines, string(data))
				break
			case EventTypeError:
				t.Fatalf("Error %v", event.Error)
				break loop
			case EventTypeComplete:
				completeCnt++
				break loop

			}
		}
	}

	if completeCnt != 1 {
		t.Fatalf("Expected complete count of %v but got %v", 1, completeCnt)
	}

	if actual < expect {
		t.Fatalf("Expected at least %v lines got %v %v", expect, actual, lines)
		return
	}

	lastLine := lines[len(lines)-1]
	if !strings.HasPrefix(lastLine, "}") {
		t.Fatalf("Expected closing brace got [%v]", lastLine)
		return
	}
}
func TestRequestJSON(t *testing.T) {
	subject, err := NewHTTPJSONSubject("http://httpbin.org/get", nil)

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
		case event := <-observer.Event:
			switch event.Type {
			case EventTypeNext:
				data := ToStringMap(event.Next, nil)
				if len(data) < 1 {
					t.Fatalf("Next invalid length for %v", data)
				}
				break
			case EventTypeError:
				t.Fatalf("Error %v", event.Error)
				return
			case EventTypeComplete:
				completeCnt++
				break loop
			}
		}
	}

	if completeCnt != 1 {
		t.Fatalf("Expected complete count of %v but got %v", 1, completeCnt)
	}
}

func TestRequestSSE(t *testing.T) {
	subject, err := NewHTTPSSESubject("http://express-eventsource.herokuapp.com/events", nil)

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
		case event := <-observer.Event:
			switch event.Type {
			case EventTypeNext:
				data := ToStringMap(event.Next, nil)
				if len(data) < 3 {
					t.Fatalf("Expected min 3 lines, but received %v", len(data))
					return
				}
				break
			case EventTypeError:
				t.Fatalf("Error %v", event.Error)
				return
			case EventTypeComplete:
				completeCnt++
				break loop
			}
		}
	}

	if completeCnt != 1 {
		t.Fatalf("Expected complete count of %v but got %v", 1, completeCnt)
	}
}
