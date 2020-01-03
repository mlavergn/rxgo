package main

import (
	"log"
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
	observer := rx.NewObserver()

	// observable := demoInterval(observer)
	// observable := demoSubject(observer)
	// observable := demoBehavior(observer)
	observable := demoReplay(observer)

	for {
		select {
		case next := <-observer.Next:
			v := rx.ToInt(next, -1)
			log.Println("next", v)
			if v == 99 || v == -1 {
				observable.Complete <- true
				break
			}
			break
		case <-observer.Error:
			log.Println("error")
			return
		case <-observer.Complete:
			log.Println("complete")
			return
		}
	}
}
