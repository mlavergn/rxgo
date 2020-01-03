package main

import (
	"log"
	"rx"
)

func demoInterval(subscriber *rx.Subscriber) *rx.Observable {
	observable := rx.NewInterval(200)
	observable.Subscribe <- subscriber
	return observable
}

func demoSubject(subscriber *rx.Subscriber) *rx.Observable {
	interval := rx.NewInterval(200)
	observable := rx.NewSubject()
	interval.Pipe(observable)
	observable.Subscribe <- subscriber
	observable.Next <- 99
	return observable
}

func demoBehavior(subscriber *rx.Subscriber) *rx.Observable {
	observable := rx.NewBehaviorSubject(99)
	observable.Subscribe <- subscriber
	return observable
}

func demoReplay(subscriber *rx.Subscriber) *rx.Observable {
	observable := rx.NewReplaySubject(5)
	observable.Next <- 11
	observable.Next <- 22
	observable.Next <- 33
	observable.Next <- 44
	observable.Next <- 55
	observable.Next <- 66
	observable.Next <- 77
	observable.Next <- 88
	observable.Next <- 99
	// we get here too quickly, so yield
	observable.Yield(1)
	observable.Subscribe <- subscriber
	return observable
}

func main() {
	subscriber := rx.NewSubscriber()
	subscriber.Take(10)

	// observable := demoInterval(subscriber)
	// observable := demoSubject(subscriber)
	// observable := demoBehavior(subscriber)
	observable := demoReplay(subscriber)

	for {
		select {
		case next := <-subscriber.Next:
			v := rx.ToInt(next, -1)
			log.Println("next", v)
			if v == 99 || v == -1 {
				observable.Complete <- true
				break
			}
			break
		case <-subscriber.Error:
			log.Println("error")
			return
		case <-subscriber.Complete:
			log.Println("complete")
			return
		}
	}
}
