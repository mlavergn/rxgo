package main

import (
	"rx"
)

func demoInterval(observer *rx.Observer) *rx.Observable {
	observable := rx.NewInterval(200)
	observable.Subscribe <- observer
	return observable
}

func demoSubject(observer *rx.Observer) *rx.Observable {
	interval := rx.NewInterval(200)
	observable := rx.NewSubject()
	interval.Pipe(observable)
	observable.Subscribe <- observer
	observable.Next <- 99
	return observable
}

func demoBehavior(observer *rx.Observer) *rx.Observable {
	observable := rx.NewBehaviorSubject(99)
	observable.Subscribe <- observer
	return observable
}

func demoReplay(observer *rx.Observer) *rx.Observable {
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
	observable.Delay(1)
	observable.Subscribe <- observer
	return observable
}

func main() {
	observer := rx.NewObserver()
	observer.Take(10)

	// observable := demoInterval(observer)
	// observable := demoSubject(observer)
	// observable := demoBehavior(observer)
	observable := demoReplay(observer)

	observer.Default(func(event interface{}) {
		v := rx.ToInt(event, -1)
		if v == 99 || v == -1 {
			observable.Unsubscribe <- observer
		}
	})
}
