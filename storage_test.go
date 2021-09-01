package main

import (
	"testing"

	"github.com/go-redis/redis"
	"github.com/kelseyhightower/envconfig"
	"github.com/stretchr/testify/assert"
)

func Test1_Storage(t *testing.T) {

	config := configT{}
	_, err := NewDatabase(redis.UniversalOptions{Addrs: config.RedisAddrs, Password: config.RedisPassword})
	assert.Error(t, err)

	envSet(t) // load .env file for test environment

	assert.NoError(t, envconfig.Process("SSP", &config))

	db, err := NewDatabase(redis.UniversalOptions{Addrs: config.RedisAddrs, Password: config.RedisPassword})
	assert.NoError(t, err)

	r := NewRound("u1", "u2")

	assert.NoError(t, db.Store(r))

	rr, _ := db.Retrieve(r.ID)

	assert.Equal(t, rr, r)

	t.Logf("Stored: %+v \n retrived: %+v \n", r, rr)

	_, err = db.Retrieve("Non-existing_key")
	assert.Error(t, err)
}
