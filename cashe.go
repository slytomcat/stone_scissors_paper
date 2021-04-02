package main

import (
	"sync"
	"time"
)

type Cache struct {
	data       sync.Map
	db         Database
	defaultexp time.Duration
}

type Data struct {
	r   *Round
	exp time.Time
}

func NewCache(db Database, exp, interval time.Duration) Database {
	cache := &Cache{
		data:       sync.Map{},
		db:         db,
		defaultexp: exp,
	}
	go cache.handler(interval)
	return cache
}

func (c *Cache) handler(interval time.Duration) {
	list := []string{}
	ex := func(key interface{}, value interface{}) bool {
		if value.(Data).exp.Sub(time.Now()) < 0 {
			list = append(list, key.(string))
		}
		return true // continue iteration
	}
	for {
		list = []string{}
		<-time.After(interval)
		c.data.Range(ex)
		for _, k := range list {
			c.data.Delete(k)
		}
	}
}

func (c *Cache) Store(r *Round) error {
	c.data.Store(r.ID, Data{
		r:   r,
		exp: time.Now().Add(c.defaultexp),
	})
	return c.db.Store(r)
}

func (c *Cache) Retrive(id string) (*Round, error) {
	d, ok := c.data.Load(id)
	if ok {
		return d.(Data).r, nil
	}
	return c.db.Retrive(id)
}
