// Package subscriptions implements a list of feeds along with operations to
// modify the list.
package subscriptions

type DB interface {
	AddToRiver(uri string) error
	RemoveFromRiver(uri string) error
	AddToGarden(uri string) error
	RemoveFromGarden(uri string) error
}

type Subscriptions struct {
	db DB
}

func New(db DB) *Subscriptions {
	return &Subscriptions{db: db}
}
