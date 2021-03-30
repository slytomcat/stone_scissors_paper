package main

import (
	"encoding/json"
	"sync"
	"time"

	redis "github.com/go-redis/redis"
)

type Database interface {
	Store(*Round) error
	Retrive(string) (*Round, error)
}

type redisDB struct {
	r redis.UniversalClient
}

func NewDatabse(opt redis.UniversalOptions) (Database, error) {
	db := redis.NewUniversalClient(&opt)
	DB := &redisDB{db}
	// try to ping database
	if err := DB.r.Ping().Err(); err != nil {
		return nil, err
	}
	return DB, nil
}

func (db *redisDB) Store(round *Round) error {
	data, _ := json.Marshal(round)
	return db.r.Set(round.ID, data, time.Hour*8760).Err()
}
func (db *redisDB) Retrive(id string) (*Round, error) {
	data, err := db.r.Get(id).Result()
	if err != nil {
		return nil, err
	}
	round := &Round{mx: sync.Mutex{}}
	if json.Unmarshal([]byte(data), round) != nil {
		return nil, err
	}
	return round, nil
}
