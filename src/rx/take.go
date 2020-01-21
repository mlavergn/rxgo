package rx

import "sync"

// Take export
func (id *Observable) Take(count int) *Observable {
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
func (id *Observable) TakeWhile(cond func() bool) *Observable {
	log.Println(id.UID, "Observable.TakeWhile")

	id.takeFn = cond

	return id
}

// TakeUntil export
func (id *Observable) TakeUntil(observable *Observable) *Observable {
	log.Println(id.UID, "Observable.TakeUntil")

	watcher := NewSubscription()
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		wg.Done()
		defer func() { id.Complete <- id }()
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
func (id *Observable) TakeUntilClose(close CloseCh) *Observable {
	log.Println(id.UID, "Observable.TakeUntilCh")

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		wg.Done()
		defer func() { id.Complete <- id }()
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
