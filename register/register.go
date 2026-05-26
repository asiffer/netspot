package register

// Callback is a generic function
type Callback[T any] func(T)

// Register is a structure that allows to store callbacks
// and then execute them
type Register[T any] struct {
	callbacks []Callback[T]
}

// NewRegister inits a new Register structure
func NewRegister[T any]() *Register[T] {
	return &Register[T]{
		callbacks: make([]Callback[T], 0),
	}
}

// Clear removes all callbacks
func (r *Register[T]) Clear() {
	r.callbacks = nil
}

// Register append a new callback
func (r *Register[T]) Register(callback Callback[T]) {
	r.callbacks = append(r.callbacks, callback)
}

// Exec calls all register callbacks
func (r *Register[T]) Exec(data T) {
	for _, callback := range r.callbacks {
		callback(data)
	}
}

// Callback2 is a generic function with 2 parameters
type Callback2[T any, U any] func(T, U)

// Register2 is a structure that allows to store callbacks with 2 parameters
// and then execute them
type Register2[T any, U any] struct {
	callbacks []Callback2[T, U]
}

// NewRegister2 inits a new Register2 structure
func NewRegister2[T any, U any]() *Register2[T, U] {
	return &Register2[T, U]{
		callbacks: make([]Callback2[T, U], 0),
	}
}

// Clear removes all callbacks
func (r *Register2[T, U]) Clear() {
	r.callbacks = nil
}

// Register append a new callback
func (r *Register2[T, U]) Register(callback Callback2[T, U]) {
	r.callbacks = append(r.callbacks, callback)
}

// Exec calls all register callbacks
func (r *Register2[T, U]) Exec(data T, data2 U) {
	for _, callback := range r.callbacks {
		callback(data, data2)
	}
}

// Callback3 is a generic function with 3 parameters
type Callback3[T any, U any, V any] func(T, U, V)

// Register3 is a structure that allows to store callbacks with 3 parameters
// and then execute them
type Register3[T any, U any, V any] struct {
	callbacks []Callback3[T, U, V]
}

// NewRegister3 inits a new Register3 structure
func NewRegister3[T any, U any, V any]() *Register3[T, U, V] {
	return &Register3[T, U, V]{
		callbacks: make([]Callback3[T, U, V], 0),
	}
}

// Clear removes all callbacks
func (r *Register3[T, U, V]) Clear() {
	r.callbacks = nil
}

// Register append a new callback
func (r *Register3[T, U, V]) Register(callback Callback3[T, U, V]) {
	r.callbacks = append(r.callbacks, callback)
}

// Exec calls all register callbacks
func (r *Register3[T, U, V]) Exec(data T, data2 U, data3 V) {
	for _, callback := range r.callbacks {
		callback(data, data2, data3)
	}
}
