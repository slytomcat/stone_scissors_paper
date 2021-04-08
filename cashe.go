package main

import (
	"sync"
	"time"
)

// Cache is an implementation of Database interface that works as passthrough memory cache
type Cache struct {
	data       sync.Map
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
		data:       sync.Map{},
		db:         db,
		defaultexp: exp,
	}
	go cache.handler(interval)
	return cache
}

// handler clears memory cache from expired data
func (c *Cache) handler(interval time.Duration) {
	list := []string{}
	ex := func(key interface{}, value interface{}) bool {
		if value.(data).exp.Sub(time.Now()) < 0 {
			list = append(list, key.(string))
		}
		return true // continue iteration
	}
	for {
		list = []string{}
		time.Sleep(interval)
		c.data.Range(ex)
		for _, k := range list {
			c.data.Delete(k)
		}
	}
}

// Store remember the Round in memory cache and save it to database
func (c *Cache) Store(r *Round) error {
	c.data.Store(r.ID, data{
		r:   r,
		exp: time.Now().Add(c.defaultexp),
	})
	return c.db.Store(r)
}

// Retrieve tries to get Round from memory cache or from db. It also stores data read from database into memory cache
func (c *Cache) Retrieve(id string) (*Round, error) {
	d, ok := c.data.Load(id)
	if ok {
		return d.(data).r, nil
	}
	r, err := c.db.Retrieve(id)
	if err != nil {
		return nil, err
	}

	c.data.Store(id, data{
		r:   r,
		exp: time.Now().Add(c.defaultexp),
	})

	return r, nil
}
