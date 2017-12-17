// Package mapping converts a feed item into a riverjs item.
package mapping

import (
	"hawx.me/code/riviera/feed/common"
	"hawx.me/code/riviera/river/riverjs"
)

// A Mapping takes an item from a feed and returns an item for the river, if nil
// is returned the item will not be added to the river.
type Mapping func(*common.Item) *riverjs.Item
