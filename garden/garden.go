// Package garden aggregates feeds into a gardenjs file.
//
// That format doesn't exist, yet, the purpose of this is to define it. The idea
// is to have something similiar/exactly the same as fraidycat does: a list of
// feeds ordered by most recently updated, with a compact list of recent items
// for each feed.
//
// This can and will co-exist nicely with rivers because they each solve
// different problems. A garden is for friends, a river is for things you don't
// mind missing. Neither are the dreaded inbox with all of the management of
// read status, etc.
//
// Hopefully this works out as nicely as I think it will.
package garden

import (
	"encoding/json"
	"errors"
	"hawx.me/code/riviera/garden/gardenjs"
	"hawx.me/code/riviera/river/data"
	"io"
	"sort"
	"sync"
	"time"
)

type Options struct {
	Size    int
	Refresh time.Duration
}

type Garden struct {
	store        data.Database
	size         int
	cacheTimeout time.Duration

	mu      sync.RWMutex
	flowers map[string]*Flower
}

func New(store data.Database, options Options) *Garden {
	if options.Size <= 0 {
		options.Size = 10
	}
	if options.Refresh <= 0 {
		options.Refresh = time.Hour
	}

	return &Garden{
		store:        store,
		size:         options.Size,
		cacheTimeout: options.Refresh,
		flowers:      map[string]*Flower{},
	}
}

func (g *Garden) Encode(w io.Writer) error {
	garden := gardenjs.Garden{
		Metadata: gardenjs.Metadata{
			BuiltAt: time.Now(),
		},
	}

	for _, flower := range g.flowers {
		garden.Feeds = append(garden.Feeds, flower.Latest())
	}

	sort.Slice(garden.Feeds, func(i, j int) bool {
		return garden.Feeds[i].UpdatedAt.Before(garden.Feeds[j].UpdatedAt)
	})

	return json.NewEncoder(w).Encode(garden)
}

func (g *Garden) Add(uri string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.flowers[uri]; exists {
		return errors.New("already added uri")
	}

	feedStore, err := g.store.Feed(uri)
	if err != nil {
		return err
	}

	flower, err := NewFlower(feedStore, g.cacheTimeout, uri, g.size)
	if err != nil {
		return err
	}
	flower.Start()

	g.flowers[uri] = flower
	return nil
}

func (g *Garden) Remove(uri string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	if flower, exists := g.flowers[uri]; exists {
		flower.Stop()
		delete(g.flowers, uri)
		return true
	}

	return false
}

func (g *Garden) Close() error {
	for _, flower := range g.flowers {
		flower.Stop()
	}
	return nil
}
