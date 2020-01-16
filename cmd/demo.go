package main

import (
	"fmt"
	"github.com/mlavergn/rxgo/src/rx"
	"sync"
	"time"
)

func demoInterval(subscription *rx.Subscription) *rx.Observable {
	interval := rx.NewInterval(200)
	interval.UID = "demoInterval"
	return interval
}

func demoSubject(subscription *rx.Subscription) *rx.Observable {
	interval := rx.NewInterval(200)
	interval.UID = "demoSubjectInterval"
	observable := rx.NewSubject()
	observable.UID = "demoSubjectObservable"
	interval.Pipe(observable)
	observable.Subscribe <- subscription
	observable.Next <- 99
	return observable
}

func demoBehavior(subscription *rx.Subscription) *rx.Observable {
	return rx.NewBehaviorSubject(99)
}

func demoReplay(subscription *rx.Subscription, count int) *rx.Observable {
	observable := rx.NewReplaySubject(count)
	observable.UID = "demoReplayObservable"

	observable.Next <- 11
	observable.Next <- 22
	observable.Next <- 33
	observable.Next <- 44
	observable.Next <- 55
	observable.Next <- 66
	observable.Next <- 77
	observable.Next <- 88
	observable.Next <- 99
	return observable
}

func demoRetry(subscription *rx.Subscription, closeCh chan bool) *rx.Observable {
	rxhttp := rx.NewHTTPRequest(0)
	observable, err := rxhttp.TextSubject("http://httpbin.org/get", nil)
	if err != nil {
		fmt.Println("demoRetry", err)
		return nil
	}
	observable.UID = "demoRetryObservable"
	retry := 2
	observable.RetryWhen(func() bool {
		retry--
		<-time.After(1 * time.Second)
		return (retry != 0)
	})
	return observable
}

func main() {
	rx.Config(false)
	closeCh := make(chan bool)
	subscription := rx.NewSubscription()
	subscription.UID = "demoSubscription"
	subscription.Take(10)

	// observable := demoInterval(subscription)
	// observable := demoSubject(subscription)
	// observable := demoBehavior(subscription)
	observable := demoReplay(subscription, 4)
	parse := true

	// observable := demoRetry(subscription, closeCh)
	// parse = false

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		wg.Done()
		for {
			select {
			case event := <-subscription.Next:
				fmt.Println(subscription.UID, event)
				if parse {
					v := rx.ToInt(event, -1)
					if v == 99 || v == -1 {
						fmt.Println("Done")
						observable.Unsubscribe <- subscription
						subscription.Complete <- true
					}
				}
			}
		}
	}()

	wg.Wait()
	observable.Subscribe <- subscription

	<-observable.Finalize
	close(closeCh)
}
