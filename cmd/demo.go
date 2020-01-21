package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/mlavergn/rxgo/src/rx"
)

func demoInterval(observer *rx.Observer) *rx.Observable {
	interval := rx.NewInterval(200)
	interval.UID = "demoInterval." + interval.UID
	return interval.Take(10)
}

func demoSubject(observer *rx.Observer) *rx.Observable {
	interval := rx.NewInterval(200)
	interval.UID = "demoSubjectInterval." + interval.UID
	observable := rx.NewSubject()
	observable.UID = "demoSubjectObservable." + observable.UID
	interval.Pipe(observable)
	observable.Subscribe <- observer
	observable.Next <- 99
	return observable.Take(10)
}

func demoBehavior(observer *rx.Observer) *rx.Observable {
	return rx.NewBehaviorSubject(99).Take(1)
}

func demoReplay(observer *rx.Observer, count int) *rx.Observable {
	observable := rx.NewReplaySubject(count)
	observable.UID = "demoReplayObservable." + observable.UID

	observable.Next <- 11
	observable.Next <- 22
	observable.Next <- 33
	observable.Next <- 44
	observable.Next <- 55
	observable.Next <- 66
	observable.Next <- 77
	observable.Next <- 88
	observable.Next <- 99
	return observable.Take(5)
}

func demoRetry(observer *rx.Observer) *rx.Observable {
	rxhttp := rx.NewHTTPRequest(10 * time.Second)
	observable, err := rxhttp.TextSubject("http://httpbin.org/get", nil)
	if err != nil {
		fmt.Println("demoRetry", err)
		return nil
	}
	observable.UID = "demoRetryObservable." + observable.UID
	retry := 2
	observable.RepeatWhen(func() bool {
		retry--
		<-time.After(1 * time.Second)
		return (retry != 0)
	}).Take(1)
	return observable
}

func demoSSE(observer *rx.Observer) *rx.Observable {
	rxhttp := rx.NewHTTPRequest(10 * time.Second)
	observable, err := rxhttp.SSESubject("http://demo.howopensource.com/sse/stocks.php", nil)
	if err != nil {
		fmt.Println("demoSSE", err)
		return nil
	}
	observable.UID = "demoSSE." + observable.UID
	observable.Map(func(event interface{}) interface{} {
		result := rx.ToStringMap(event, nil)
		return result
	}).Take(5)
	return observable
}

func main() {
	rx.Config(false)
	closeCh := make(chan bool)
	observer := rx.NewObserver()
	observer.UID = "demoSubscription." + observer.UID

	parse := true
	// observable := demoInterval(observer)
	// observable := demoSubject(observer)
	// observable := demoBehavior(observer)
	observable := demoReplay(observer, 4)

	// observable := demoRetry(observer)
	// observable := demoSSE(observer)
	// parse = false

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		wg.Done()
		for {
			select {
			case event := <-observer.Next:
				fmt.Println(observer.UID, event)
				if parse {
					v := rx.ToInt(event, -1)
					if v == 99 || v == -1 {
						fmt.Println("Done")
						observable.Unsubscribe <- observer
					}
				}
			}
		}
	}()

	wg.Wait()
	observable.Subscribe <- observer

	<-observable.Finalize
	close(closeCh)
}
