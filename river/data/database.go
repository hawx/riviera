// Package data provides an interface for saving and retrieving data from a
// key-value database arranged into buckets.
package data

// Database is a key-value store with data arranged in buckets.
type Database interface {
	// Bucket returns a namespaced bucket for storing key-value data.
	Bucket(name []byte) (Bucket, error)

	// Close releases all database resources. All transactions must be complete
	// before closing.
	Close() error
}

// Bucket is a table for storing key-value pairs.
type Bucket interface {
	// View executes the function in the context of a read-only transaction.
	View(func(ReadTx) error) error

	// Update executes the function in the context of a read-write transaction.
	Update(func(Tx) error) error
}

// ReadTx is a read-only database transaction.
type ReadTx interface {
	// Get returns the value associated with a key. Returns a nil value if the key does not exist.
	Get(key []byte) []byte

	// After returns all values listed after the key given, inclusive, to the last
	// value, in sorted key order.
	After(first []byte) [][]byte

	// KeysBefore returns all keys from the first to the key given, exlusive, in
	// sorted key order.
	KeysBefore(last []byte) [][]byte

	// All returns all values listed in sorted key order.
	All() [][]byte
}

// Tx is a database transaction.
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

	// KeysBefore returns all keys from the first to the key given, exlusive, in
	// sorted key order.
	KeysBefore(last []byte) [][]byte

	// All returns all values listed in sorted key order.
	All() [][]byte
}
