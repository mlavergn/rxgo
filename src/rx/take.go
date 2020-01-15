package rx

import "sync"

// Take export
func (id *Subscription) Take(count int) *Subscription {
	log.Println(id.UID, "Observable.Take")

	counter := count
	id.takeFn = func() bool {
		counter--
		dlog.Println(id.UID, "Observable.Take state", (counter > 0), counter)
		return (counter > 0)
	}

	return id
}

// TakeWhile export
func (id *Subscription) TakeWhile(cond func() bool) *Subscription {
	log.Println(id.UID, "Observable.TakeWhile")

	id.takeFn = cond

	return id
}

// TakeUntil export
func (id *Subscription) TakeUntil(observable *Observable) *Subscription {
	log.Println(id.UID, "Observable.TakeUntil")

	watcher := NewSubscription()
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		wg.Done()
		defer id.complete()
		select {
		case <-watcher.Next:
			return
		case <-watcher.Error:
			return
		case <-watcher.Complete:
			return
		}
	}()

	wg.Wait()
	observable.Subscribe <- watcher

	return id
}

// TakeUntilClose export
func (id *Subscription) TakeUntilClose(close CloseCh) *Subscription {
	log.Println(id.UID, "Observable.TakeUntilCh")

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		wg.Done()
		defer id.complete()
		<-close
	}()

	wg.Wait()

	return id
}

//
// TakeUntil(close) is a common pattern, this basic channel based close keeps use consistent
//

// CloseCh type
type CloseCh chan bool

// NewCloseCh export
func NewCloseCh() CloseCh {
	return make(chan bool, 1)
}
