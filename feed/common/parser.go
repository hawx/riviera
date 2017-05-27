package common

import "io"

type Parser interface {
	CanRead(io.Reader, func(charset string, input io.Reader) (io.Reader, error)) bool
	Read(io.Reader, func(charset string, input io.Reader) (io.Reader, error)) ([]*Channel, error)
}
