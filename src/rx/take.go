package rx

// Take export
func (id *Subscriber) Take(count int) *Subscriber {
	log.Println(id.UID, "Observable.Take")

	counter := count
	id.take = func() bool {
		counter--
		return (counter > 0)
	}

	return id
}

// TakeWhile export
func (id *Subscriber) TakeWhile(cond func() bool) *Subscriber {
	log.Println(id.UID, "Observable.TakeWhile")

	id.take = cond

	return id
}

// TakeUntil export
func (id *Subscriber) TakeUntil(observable *Observable) *Subscriber {
	log.Println(id.UID, "Observable.TakeUntil")

	subscriber := NewSubscriber()
	observable.Subscribe <- subscriber
	go func() {
		<-subscriber.Next
		id.complete()
	}()

	return id
}
