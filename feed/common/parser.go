// Package common contains structs for a unified feed format.
package common

import (
	"io"
	"net/url"
)

type Parser interface {
	CanRead(io.Reader, func(charset string, input io.Reader) (io.Reader, error)) bool
	Read(io.Reader, *url.URL, func(charset string, input io.Reader) (io.Reader, error)) ([]*Channel, error)
}
