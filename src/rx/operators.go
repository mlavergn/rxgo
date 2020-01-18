package rx

// Pipe operator
func (id *Observable) Pipe(pipe *Observable) *Observable {
	log.Println(id.UID, "Observable.Pipe")
	id.Subscribe <- pipe.Subscription
	return id
}

// Filter export
// Emit values that PASS (return true) for the filter condition
func (id *Observable) Filter(fn func(interface{}) bool) *Observable {
	log.Println(id.UID, "Observable.Filter")
	id.operators = append(id.operators, operator{operatorFilter, fn})
	return id
}

// Map export
func (id *Observable) Map(fn func(interface{}) interface{}) *Observable {
	log.Println(id.UID, "Observable.Map")
	id.operators = append(id.operators, operator{operatorMap, fn})
	return id
}
