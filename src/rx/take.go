package rx

// Take export
func (id *Subscription) Take(count int) *Subscription {
	log.Println(id.UID, "Observable.Take")

	counter := count
	id.take = func() bool {
		counter--
		return (counter > 0)
	}

	return id
}

// TakeWhile export
func (id *Subscription) TakeWhile(cond func() bool) *Subscription {
	log.Println(id.UID, "Observable.TakeWhile")

	id.take = cond

	return id
}

// TakeUntil export
func (id *Subscription) TakeUntil(observable *Observable) *Subscription {
	log.Println(id.UID, "Observable.TakeUntil")

	subscription := NewSubscription()
	observable.Subscribe <- subscription
	go func() {
		<-subscription.Next
		id.complete()
	}()

	return id
}
