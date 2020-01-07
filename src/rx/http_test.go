package rx

import (
	"testing"
)

func TestRequestText(t *testing.T) {
	rxhttp := NewRequest(0)
	observable, err := rxhttp.TextSubject("http://httpbin.org/get", nil)

	if err != nil {
		t.Fatalf("Init error %v", err)
		return
	}

	subscription := NewSubscription()
	observable.Subscribe <- subscription
	select {
	case event := <-subscription.Next:
		log.Println(event)
		break
	case err := <-subscription.Error:
		t.Fatalf("Error %v", err)
		return
	case <-subscription.Complete:
		log.Println("complete")
		return
	}
}

func TestRequestLine(t *testing.T) {
	expect := 11
	actual := 0
	rxhttp := NewRequest(0)
	observable, err := rxhttp.LineSubject("http://httpbin.org/get", nil)

	if err != nil {
		t.Fatalf("Init error %v", err)
		return
	}

	subscription := NewSubscription()
	observable.Subscribe <- subscription
	for {
		select {
		case event := <-subscription.Next:
			actual++
			log.Println(event)
			break
		case err := <-subscription.Error:
			t.Fatalf("Error %v", err)
			return
		case <-subscription.Complete:
			if actual != expect {
				t.Fatalf("Expected %v lines got %v", expect, actual)
				return
			}
			log.Println("complete")
			return
		}
	}
}
func TestRequestJSON(t *testing.T) {
	rxhttp := NewRequest(0)
	observable, err := rxhttp.JSONSubject("http://httpbin.org/get", nil)

	if err != nil {
		t.Fatalf("Init error %v", err)
		return
	}

	subscription := NewSubscription()
	observable.Subscribe <- subscription
	select {
	case event := <-subscription.Next:
		log.Println(event)
		break
	case err := <-subscription.Error:
		t.Fatalf("Error %v", err)
		return
	case <-subscription.Complete:
		log.Println("complete")
		return
	}
}

func TestRequestSSE(t *testing.T) {
	rxhttp := NewRequest(0)
	observable, err := rxhttp.SSESubject("http://express-eventsource.herokuapp.com/events", nil)

	if err != nil {
		t.Fatalf("Init error %v", err)
		return
	}

	subscription := NewSubscription()
	observable.Subscribe <- subscription
	select {
	case event := <-subscription.Next:
		data := ToByteArrayArray(event, nil)
		if len(data) < 3 {
			t.Fatalf("Expected min 3 lines, but received %v", len(data))
			return
		}
		break
	case err := <-subscription.Error:
		t.Fatalf("Error %v", err)
		return
	case <-subscription.Complete:
		log.Println("complete")
		return
	}
}
