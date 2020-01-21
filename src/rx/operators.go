package rx

// Pipe forwards events to the observer
func (id *Observable) Pipe(observer *Observable) *Observable {
	log.Println(id.UID, "Observable.Pipe")
	id.Subscribe <- observer.Subscription
	return id
}

// Filter export
// Emit values that PASS (return true) for the filter condition
func (id *Observable) Filter(fn func(interface{}) bool) *Observable {
	log.Println(id.UID, "Observable.Filter")
	id.operators = append(id.operators, operator{operatorFilter, fn})
	return id
}

// Map modifies the event type
func (id *Observable) Map(fn func(interface{}) interface{}) *Observable {
	log.Println(id.UID, "Observable.Map")
	id.operators = append(id.operators, operator{operatorMap, fn})
	return id
}
