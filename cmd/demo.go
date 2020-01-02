package main

import (
	"fmt"
	"rxgo"
	"time"
)

func demoInterval(observer *rxgo.RxObserver) *rxgo.RxObservable {
	observable := rxgo.NewRxInterval(200)
	observable.Take(10)
	observable.Subscribe <- observer
	return observable
}

func demoSubject(observer *rxgo.RxObserver) *rxgo.RxObservable {
	interval := rxgo.NewRxInterval(200)
	observable := rxgo.NewRxSubject()
	interval.Pipe(observable)
	observable.Take(10)
	observable.Subscribe <- observer
	observable.Next <- 99
	return observable
}

func demoBehavior(observer *rxgo.RxObserver) *rxgo.RxObservable {
	observable := rxgo.NewRxBehaviorSubject(99)
	observable.Subscribe <- observer
	return observable
}

func demoReplay(observer *rxgo.RxObserver) *rxgo.RxObservable {
	observable := rxgo.NewRxReplaySubject(5)
	observable.Next <- 11
	observable.Next <- 22
	observable.Next <- 33
	observable.Next <- 44
	observable.Next <- 55
	observable.Next <- 66
	observable.Next <- 77
	observable.Next <- 88
	observable.Next <- 99
	// sleep just enough to allow the observable to get hot
	<-time.After(1 * time.Millisecond)
	observable.Subscribe <- observer
	return observable
}

func main() {
	rxgo.RxSetup(false)

	observer := rxgo.NewRxObserver()

	// demoInterval(observer)
	// demoSubject(observer)
	// demoBehavior(observer)
	demoReplay(observer)

	for {
		select {
		case next := <-observer.Next:
			v := rxgo.ToInt(next, -1)
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
