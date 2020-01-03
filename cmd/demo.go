package main

import (
	"fmt"
	"rx"
)

func demoInterval(observer *rx.Observer) *rx.Observable {
	observable := rx.NewInterval(200)
	observable.Take(10)
	observable.Subscribe <- observer
	return observable
}

func demoSubject(observer *rx.Observer) *rx.Observable {
	interval := rx.NewInterval(200)
	observable := rx.NewSubject()
	interval.Pipe(observable)
	observable.Take(10)
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
	observable.Warmup()
	observable.Subscribe <- observer
	return observable
}

func main() {
	rx.Setup(false)

	observer := rx.NewObserver()

	// demoInterval(observer)
	// demoSubject(observer)
	// demoBehavior(observer)
	demoReplay(observer)

	for {
		select {
		case next := <-observer.Next:
			v := rx.ToInt(next, -1)
			fmt.Println("next", v)
			if v == 99 || v == -1 {
				return
			}
			break
		case <-observer.Error:
			fmt.Println("error")
			return
		case <-observer.Complete:
			fmt.Println("complete")
			return
		}
	}
}
