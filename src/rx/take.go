package rx

// Take export
func (id *Observer) Take(count int) *Observer {
	log.Println(id.UID, "Observable.Take")

	counter := count
	id.take = func() bool {
		counter--
		return (counter > 0)
	}

	return id
}

// TakeWhile export
func (id *Observer) TakeWhile(cond func() bool) *Observer {
	log.Println(id.UID, "Observable.TakeWhile")

	id.take = cond

	return id
}

// TakeUntil export
func (id *Observer) TakeUntil(observable *Observable) *Observer {
	log.Println(id.UID, "Observable.TakeUntil")

	observer := NewObserver()
	observable.Subscribe <- observer
	go func() {
		<-observer.Next
		id.complete()
	}()

	return id
}
