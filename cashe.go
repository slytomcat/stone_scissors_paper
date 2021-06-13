package main

import (
	"sync"
	"time"
)

// Cache is an implementation of Database interface that works as passthrough memory cache
type Cache struct {
	data       map[string]data //sync.Map
	mux        sync.RWMutex
	db         Database
	defaultexp time.Duration
}

// data is a memory cache element
type data struct {
	r   *Round    // stored element
	exp time.Time // memory cache expiration time
}

// NewCache returns a new Database interface with the passthrough memory cache as wrapping of the provided Database interface
// db - the Database interface to wrap
// exp - default data expiration in the memory cache
// interval - memery cache clearing interval
func NewCache(db Database, exp, interval time.Duration) Database {
	cache := &Cache{
		data:       map[string]data{}, // sync.Map{},
		mux:        sync.RWMutex{},
		db:         db,
		defaultexp: exp,
	}
	go cache.handler(interval)
	return cache
}

// handler clears memory cache from expired data
func (c *Cache) handler(interval time.Duration) {
	list := []string{}
	remove := func() {
		c.mux.Lock()
		defer c.mux.Unlock()
		for _, key := range list {
			delete(c.data, key)
		}
	}
	for {
		time.Sleep(interval)
		list = []string{}
		c.mux.RLock()
		for k, v := range c.data {
			if time.Until(v.exp) < 0 {
				list = append(list, k)
			}
		}
		c.mux.RUnlock()
		remove()
	}
}

// Store remember the Round in memory cache and save it to database
func (c *Cache) Store(r *Round) error {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.data[r.ID] = data{
		r:   r,
		exp: time.Now().Add(c.defaultexp),
	}
	return c.db.Store(r)
}

// Retrieve tries to get Round from memory cache or from db. It also stores data read from database into memory cache
func (c *Cache) Retrieve(id string) (*Round, error) {
	c.mux.RLock()
	d, ok := c.data[id]
	c.mux.RUnlock()
	if ok {
		return d.r, nil
	}
	r, err := c.db.Retrieve(id)
	if err != nil {
		return nil, err
	}
	c.mux.Lock()
	defer c.mux.Unlock()
	c.data[id] = data{
		r:   r,
		exp: time.Now().Add(c.defaultexp),
	}

	return r, nil
}
