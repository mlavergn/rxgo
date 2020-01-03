package rx

import (
	"log"
	"testing"
)

func TestRxHTTPText(t *testing.T) {
	rxhttp := NewRxHTTP(0)
	observable, err := rxhttp.Text("http://iot.reqly.com")

	if err != nil {
		t.Fatalf("TestRxHTTPText init error %v", err)
		return
	}

	subscriber := NewRxObserver()
	observable.Subscribe <- subscriber
	select {
	case event := <-subscriber.Next:
		log.Println(event)
		break
	case err := <-subscriber.Error:
		t.Fatalf("TestRxHTTPText error %v", err)
		return
	case <-subscriber.Complete:
		log.Println("TestRxHTTPText complete")
		return
	}
}

func TestRxHTTPJSON(t *testing.T) {
	rxhttp := NewRxHTTP(0)
	observable, err := rxhttp.JSON("http://iot.reqly.com")

	if err != nil {
		t.Fatalf("TestRxHTTPJSON init error %v", err)
		return
	}

	subscriber := NewRxObserver()
	observable.Subscribe <- subscriber
	select {
	case event := <-subscriber.Next:
		log.Println(event)
		break
	case err := <-subscriber.Error:
		t.Fatalf("TestRxHTTPJSON error %v", err)
		return
	case <-subscriber.Complete:
		log.Println("TestRxHTTPJSON complete")
		return
	}
}

func TestRxHTTPSSE(t *testing.T) {
	rxhttp := NewRxHTTP(0)
	observable, err := rxhttp.SSE("http://express-eventsource.herokuapp.com/events")

	if err != nil {
		t.Fatalf("TestRxHTTPSSE init error %v", err)
		return
	}

	subscriber := NewRxObserver()
	observable.Subscribe <- subscriber
	select {
	case event := <-subscriber.Next:
		data := ToByteArrayArray(event, nil)
		if len(data) < 3 {
			t.Fatalf("TestRxHTTPSSE exppected min 3 lines, but received %v", len(data))
		}
		break
	case err := <-subscriber.Error:
		t.Fatalf("TestRxHTTPSSE error %v", err)
		return
	case <-subscriber.Complete:
		log.Println("TestRxHTTPSSE complete")
		return
	}
}
