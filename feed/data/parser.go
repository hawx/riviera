package data

import "github.com/jteeuwen/go-pkg-xmlx"

type Parser interface {
	CanRead(*xmlx.Document) bool
	Read(*xmlx.Document) ([]*Channel, error)
}
