package main

import (
	"fmt"
	"rx"
	"time"
	"sync"
)

func demoInterval(subscription *rx.Subscription) *rx.Observable {
	return rx.NewInterval(200)
}

func demoSubject(subscription *rx.Subscription) *rx.Observable {
	interval := rx.NewInterval(200)
	observable := rx.NewSubject()
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
	return observable
}

func demoRetry(subscription *rx.Subscription, closeCh chan bool) *rx.Observable {
	rxhttp := rx.NewRequest(0)
	observable, err := rxhttp.TextSubject("http://httpbin.org/get", nil)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	retry := 2
	observable.RetryWhen(func() bool {
		retry--
		<-time.After(1 * time.Second)
		return (retry != 0)
	})
	return observable
}

func main() {
	closeCh := make(chan bool)
	subscription := rx.NewSubscription()
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
		subscription.Default(func(event interface{}) {
			if parse {
				v := rx.ToInt(event, -1)
				if v == 99 || v == -1 {
					observable.Unsubscribe <- subscription
				}
			}
		}, closeCh)
	}()

	wg.Wait()
	observable.Subscribe <- subscription

	<-closeCh
	close(closeCh)
}
