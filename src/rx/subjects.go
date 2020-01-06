package rx

// -----------------------------------------------------------------------------
// Subjects

// NewSubject init
func NewSubject() *Observable {
	log.Println("Observable.NewSubject")
	id := NewObservable()
	id.isMulticast = true

	return id
}

// NewBehaviorSubject init
func NewBehaviorSubject(value interface{}) *Observable {
	log.Println("Observable.NewSubject")
	id := NewObservable()
	id.isMulticast = true
	id.Behavior(value)

	return id
}

// NewReplaySubject init
func NewReplaySubject(bufferSize int) *Observable {
	log.Println("Observable.NewSubject")
	id := NewObservable()
	id.isMulticast = true
	id.Replay(bufferSize)

	return id
}
