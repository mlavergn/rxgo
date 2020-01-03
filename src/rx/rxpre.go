package rx

// Filter export
func (id *Observable) Filter(fn func(interface{}) bool) *Observable {
	log.Println("Observable.Filter")
	id.filtered = fn
	return id
}

// Map export
func (id *Observable) Map(fn func(interface{}) interface{}) *Observable {
	log.Println("Observable.Map")
	id.mapped = fn
	return id
}
