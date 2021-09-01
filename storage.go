package main

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/go-redis/redis"
)

// Database is an interface for the persistence layer
type Database interface {
	Store(*Round) error
	Retrieve(string) (*Round, error)
}

// redisDB is a Redis implementation of Database interface
type redisDB struct {
	r redis.UniversalClient
}

// NewDatabase returns a new instance of Database interface implementing the persistence layer via Redis
func NewDatabase(opt redis.UniversalOptions) (Database, error) {
	db := redis.NewUniversalClient(&opt)
	DB := &redisDB{db}
	// try to ping database
	if err := DB.r.Ping().Err(); err != nil {
		return nil, err
	}
	return DB, nil
}

// Store stores data to database
func (db *redisDB) Store(round *Round) error {
	data, _ := json.Marshal(round)
	return db.r.Set(round.ID, data, time.Hour*8760).Err()
}

// Retrieve reads the data from database
func (db *redisDB) Retrieve(id string) (*Round, error) {
	data, err := db.r.Get(id).Result()
	if err != nil {
		return nil, err
	}
	round := &Round{mx: sync.RWMutex{}}
	if err := json.Unmarshal([]byte(data), round); err != nil {
		return nil, err
	}
	return round, nil
}
