package rx

import (
	"testing"
)

func TestRequestText(t *testing.T) {
	rxhttp := NewRequest(0)
	observable, err := rxhttp.TextSubject("http://httpbin.org/get")

	if err != nil {
		t.Fatalf("Init error %v", err)
		return
	}

	subscriber := NewObserver()
	observable.Subscribe <- subscriber
	select {
	case event := <-subscriber.Next:
		log.Println(event)
		break
	case err := <-subscriber.Error:
		t.Fatalf("Error %v", err)
		return
	case <-subscriber.Complete:
		log.Println("complete")
		return
	}
}

func TestRequestLine(t *testing.T) {
	expect := 11
	actual := 0
	rxhttp := NewRequest(0)
	observable, err := rxhttp.LineSubject("http://httpbin.org/get")

	if err != nil {
		t.Fatalf("Init error %v", err)
		return
	}

	subscriber := NewObserver()
	observable.Subscribe <- subscriber
	for {
		select {
		case event := <-subscriber.Next:
			actual++
			log.Println(event)
			break
		case err := <-subscriber.Error:
			t.Fatalf("Error %v", err)
			return
		case <-subscriber.Complete:
			if actual != expect {
				t.Fatalf("Expected %v lines got %v", expect, actual)
			}
			log.Println("complete")
			return
		}
	}
}
func TestRequestJSON(t *testing.T) {
	rxhttp := NewRequest(0)
	observable, err := rxhttp.JSONSubject("http://httpbin.org/get")

	if err != nil {
		t.Fatalf("Init error %v", err)
		return
	}

	subscriber := NewObserver()
	observable.Subscribe <- subscriber
	select {
	case event := <-subscriber.Next:
		log.Println(event)
		break
	case err := <-subscriber.Error:
		t.Fatalf("Error %v", err)
		return
	case <-subscriber.Complete:
		log.Println("complete")
		return
	}
}

func TestRequestSSE(t *testing.T) {
	rxhttp := NewRequest(0)
	observable, err := rxhttp.SSESubject("http://express-eventsource.herokuapp.com/events")

	if err != nil {
		t.Fatalf("Init error %v", err)
		return
	}

	subscriber := NewObserver()
	observable.Subscribe <- subscriber
	select {
	case event := <-subscriber.Next:
		data := ToByteArrayArray(event, nil)
		if len(data) < 3 {
			t.Fatalf("Expected min 3 lines, but received %v", len(data))
		}
		break
	case err := <-subscriber.Error:
		t.Fatalf("Error %v", err)
		return
	case <-subscriber.Complete:
		log.Println("complete")
		return
	}
}
