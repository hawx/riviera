// Package data provides an interface for saving and retrieving data from a
// key-value database arranged into buckets.
package data

type Database interface {
	// Bucket returns a namespaced bucket for storing key-value data.
	Bucket(name []byte) (Bucket, error)

	// Close releases all database resources. All transactions must be complete
	// before closing.
	Close() error
}

type Bucket interface {
	// View executes the function in the context of a read-only transaction.
	View(func(Tx) error) error

	// Update executes the function in the context of a read-write transaction.
	Update(func(Tx) error) error
}

type Tx interface {
	// Get returns the value associated with a key. Returns a nil value if the key does not exist.
	Get(key []byte) []byte

	// Put sets the value associated with a key.
	Put(key, value []byte) error

	// Delete removes the value associated with a key.
	Delete(key []byte) error

	// After returns all values listed after the key given to the last value, in
	// sorted key order.
	After(start []byte) [][]byte

	// All returns all values listed in sorted key order.
	All() [][]byte
}
