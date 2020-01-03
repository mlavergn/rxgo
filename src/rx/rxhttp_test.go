package rx

import (
	"testing"
)

func TestRequestText(t *testing.T) {
	rxhttp := NewRequest(0)
	observable, err := rxhttp.TextSubject("http://iot.reqly.com")

	if err != nil {
		t.Fatalf("TestHTTPText init error %v", err)
		return
	}

	subscriber := NewObserver()
	observable.Subscribe <- subscriber
	select {
	case event := <-subscriber.Next:
		log.Println(event)
		break
	case err := <-subscriber.Error:
		t.Fatalf("TestHTTPText error %v", err)
		return
	case <-subscriber.Complete:
		log.Println("TestHTTPText complete")
		return
	}
}

func TestRequestJSON(t *testing.T) {
	rxhttp := NewRequest(0)
	observable, err := rxhttp.JSONSubject("http://iot.reqly.com")

	if err != nil {
		t.Fatalf("TestHTTPJSON init error %v", err)
		return
	}

	subscriber := NewObserver()
	observable.Subscribe <- subscriber
	select {
	case event := <-subscriber.Next:
		log.Println(event)
		break
	case err := <-subscriber.Error:
		t.Fatalf("TestHTTPJSON error %v", err)
		return
	case <-subscriber.Complete:
		log.Println("TestHTTPJSON complete")
		return
	}
}

func TestRequestSSE(t *testing.T) {
	rxhttp := NewRequest(0)
	observable, err := rxhttp.SSESubject("http://express-eventsource.herokuapp.com/events")

	if err != nil {
		t.Fatalf("TestHTTPSSE init error %v", err)
		return
	}

	subscriber := NewObserver()
	observable.Subscribe <- subscriber
	select {
	case event := <-subscriber.Next:
		data := ToByteArrayArray(event, nil)
		if len(data) < 3 {
			t.Fatalf("TestHTTPSSE exppected min 3 lines, but received %v", len(data))
		}
		break
	case err := <-subscriber.Error:
		t.Fatalf("TestHTTPSSE error %v", err)
		return
	case <-subscriber.Complete:
		log.Println("TestHTTPSSE complete")
		return
	}
}
