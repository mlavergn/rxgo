package main

import (
	"fmt"
	"rxgo"
)

func demoInterval(observer *rxgo.RxObserver) {
	observable := rxgo.NewRxInterval(200)
	observable.Take(10)
	observable.Subscribe <- observer
}

func demoSubject(observer *rxgo.RxObserver) {
	interval := rxgo.NewRxInterval(200)
	observable := rxgo.NewRxSubject()
	interval.Pipe(observable)
	observable.Take(10)
	observable.Subscribe <- observer
	observable.Next(99)
}

func main() {
	rxgo.RxSetup(true)

	obvr := rxgo.NewRxObserver()

	go func() {
		// demoInterval(obvr)
		demoSubject(obvr)
	}()

	for {
		select {
		case next := <-obvr.OnNext:
			fmt.Println("next", next.(int))
			break
		case <-obvr.OnError:
			fmt.Println("error")
			return
		case <-obvr.OnComplete:
			fmt.Println("complete")
			return
		}
	}
}
