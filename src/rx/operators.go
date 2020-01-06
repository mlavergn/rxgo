package rx

// Pipe operator
func (id *Observable) Pipe(pipe *Observable) *Observable {
	log.Println(id.UID, "Observable.Pipe")
	id.Subscribe <- pipe.Observer
	return id
}

// Filter export
func (id *Observable) Filter(fn func(interface{}) bool) *Observable {
	log.Println(id.UID, "Observable.Filter")
	id.filtered = fn
	return id
}

// Map export
func (id *Observable) Map(fn func(interface{}) interface{}) *Observable {
	log.Println(id.UID, "Observable.Map")
	id.mapped = fn
	return id
}
