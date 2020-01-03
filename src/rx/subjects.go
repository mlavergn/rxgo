package rx

// -----------------------------------------------------------------------------
// Subjects

// NewSubject init
func NewSubject() *Observable {
	log.Println("Observable.NewSubject")
	id := NewObservable()
	id.multicast = true

	return id
}

// NewBehaviorSubject init
func NewBehaviorSubject(value interface{}) *Observable {
	log.Println("Observable.NewSubject")
	id := NewObservable()
	id.multicast = true
	id.Behavior(value)

	return id
}

// NewReplaySubject init
func NewReplaySubject(bufferSize int) *Observable {
	log.Println("Observable.NewSubject")
	id := NewObservable()
	id.multicast = true
	id.Replay(bufferSize)

	return id
}
